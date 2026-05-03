from __future__ import annotations

import logging
import threading
from dataclasses import dataclass, field

from confluent_kafka import (
    Consumer,
    KafkaError,
    KafkaException,
    TopicPartition,
)

from src.avro import AvroDecoder, DeserializedRecord
from src.config.settings import KafkaSettings

logger = logging.getLogger("acc-dp-consumer.kafka")


@dataclass
class TopicMetrics:
    consumed: int = 0
    errors: int = 0
    bytes_consumed: int = 0
    last_offset: dict[int, int] = field(default_factory=dict)
    last_partition: dict[int, int] = field(default_factory=dict)


class TopicConsumer:
    def __init__(self, settings: KafkaSettings):
        self._settings = settings
        self._topics = settings.topics
        self._running = False
        self._assigned = threading.Event()
        self._commit_lock = threading.Lock()
        self._metrics: dict[str, TopicMetrics] = {t: TopicMetrics() for t in self._topics}

        config = {
            "bootstrap.servers": settings.brokers,
            "group.id": settings.group_id,
            "session.timeout.ms": settings.session_timeout_ms,
            "auto.offset.reset": settings.auto_offset_reset,
            "enable.auto.commit": False,
            "log.connection.close": False,
            "on_commit": self._on_commit,
        }

        self._decoder = AvroDecoder(settings.schema_registry_url)
        self._consumer = Consumer(config)

    def subscribe(self) -> None:
        self._consumer.subscribe(self._topics, on_assign=self._on_assign, on_revoke=self._on_revoke)
        logger.info("subscribed to topics: %s", self._topics)

    def poll(self, timeout: float = 1.0) -> DeserializedRecord | None:
        msg = self._consumer.poll(timeout)
        if msg is None:
            return None

        if msg.error():
            return self._handle_error(msg.error())

        topic = msg.topic()
        metrics = self._metrics.get(topic)
        record = self._decoder.decode(msg)

        if record is not None and metrics is not None:
            metrics.consumed += 1
            metrics.bytes_consumed += len(msg.value()) if msg.value() else 0
            metrics.last_offset[msg.partition()] = msg.offset()
            metrics.last_partition[msg.partition()] = msg.partition()

        if metrics is not None and metrics.consumed % 1000 == 0:
            logger.info(
                "topic=%s consumed=%d errors=%d bytes=%d partitions=%s",
                topic,
                metrics.consumed,
                metrics.errors,
                metrics.bytes_consumed,
                len(metrics.last_offset),
            )

        return record

    def commit(self, asynchronous: bool = False) -> None:
        with self._commit_lock:
            try:
                self._consumer.commit(asynchronous=asynchronous)
            except KafkaException as exc:
                logger.warning("commit failed: %s", exc)

    def close(self) -> None:
        self._running = False
        try:
            self._consumer.close()
        except Exception:
            pass
        self._log_final_metrics()
        logger.info("consumer closed")

    def metrics(self) -> dict[str, TopicMetrics]:
        return dict(self._metrics)

    @property
    def running(self) -> bool:
        return self._running

    @running.setter
    def running(self, value: bool) -> None:
        self._running = value

    def wait_for_assignment(self, timeout: float = 30.0) -> bool:
        return self._assigned.wait(timeout=timeout)

    def _handle_error(self, error: KafkaError) -> None:
        if error.code() == KafkaError._PARTITION_EOF:
            return None
        if error.code() == KafkaError._TRANSPORT:
            logger.error("transport error, will retry: %s", error)
            return None
        if error.fatal():
            logger.error("fatal kafka error: %s", error)
            raise KafkaException(error)
        logger.warning("kafka error (code=%d): %s", error.code(), error)
        return None

    def _on_assign(self, consumer: Consumer, partitions: list[TopicPartition]) -> None:
        for tp in partitions:
            logger.info("assigned %s [%d] offset=%s", tp.topic, tp.partition, tp.offset)
        self._assigned.set()

    def _on_revoke(self, consumer: Consumer, partitions: list[TopicPartition]) -> None:
        for tp in partitions:
            logger.info("revoked %s [%d]", tp.topic, tp.partition)
        self._assigned.clear()
        try:
            consumer.commit(asynchronous=False)
        except KafkaException as exc:
            logger.warning("commit on revoke failed: %s", exc)

    @staticmethod
    def _on_commit(err: KafkaError | None, partitions: list[TopicPartition]) -> None:
        if err is not None:
            logger.warning("async commit error: %s", err)
            return
        for tp in partitions:
            logger.debug("committed %s [%d] offset=%s", tp.topic, tp.partition, tp.offset)

    def _log_final_metrics(self) -> None:
        for topic, m in self._metrics.items():
            logger.info(
                "final metrics topic=%s consumed=%d errors=%d bytes=%d",
                topic,
                m.consumed,
                m.errors,
                m.bytes_consumed,
            )
