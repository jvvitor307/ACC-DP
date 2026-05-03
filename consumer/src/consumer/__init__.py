from __future__ import annotations

import logging
import threading

from confluent_kafka import Consumer, KafkaError, KafkaException

from src.avro import AvroDecoder, DeserializedRecord
from src.config.settings import KafkaSettings

logger = logging.getLogger("acc-dp-consumer.kafka")


class TopicConsumer:
    def __init__(self, settings: KafkaSettings):
        config = {
            "bootstrap.servers": settings.brokers,
            "group.id": settings.group_id,
            "session.timeout.ms": settings.session_timeout_ms,
            "auto.offset.reset": settings.auto_offset_reset,
            "enable.auto.commit": settings.enable_auto_commit,
            "log.connection.close": False,
        }
        self._consumer = Consumer(config)
        self._decoder = AvroDecoder(settings.schema_registry_url)
        self._topics = settings.topics
        self._running = False
        self._commit_lock = threading.Lock()

    def subscribe(self) -> None:
        self._consumer.subscribe(self._topics)
        logger.info("subscribed to topics: %s", self._topics)

    def poll(self, timeout: float = 1.0) -> DeserializedRecord | None:
        msg = self._consumer.poll(timeout)
        if msg is None:
            return None

        if msg.error():
            if msg.error().code() == KafkaError._PARTITION_EOF:
                return None
            raise KafkaException(msg.error())

        return self._decoder.decode(msg)

    def commit(self, asynchronous: bool = False) -> None:
        with self._commit_lock:
            self._consumer.commit(asynchronous=asynchronous)

    def close(self) -> None:
        self._running = False
        self._consumer.close()
        logger.info("consumer closed")

    @property
    def running(self) -> bool:
        return self._running

    @running.setter
    def running(self, value: bool) -> None:
        self._running = value
