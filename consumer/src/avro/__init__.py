from __future__ import annotations

import logging
import struct
import threading
from dataclasses import dataclass

from confluent_kafka.schema_registry import SchemaRegistryClient, Schema
from confluent_kafka.schema_registry.avro import AvroDeserializer
from confluent_kafka.serialization import SerializationContext, MessageField

from src.avro.envelope import Envelope

logger = logging.getLogger("acc-dp-consumer.avro")

_CONFLUENT_MAGIC = 0x00
_HEADER_SIZE = 5


@dataclass(frozen=True)
class SchemaInfo:
    schema_id: int
    version: int
    subject: str
    schema_str: str


@dataclass
class DeserializedRecord:
    envelope: Envelope
    topic: str
    partition: int
    offset: int
    key: str | None
    schema_id: int
    schema_version: int

    @property
    def value(self) -> dict:
        return self.envelope.to_dict()


class SchemaVersionCache:
    def __init__(self, registry: SchemaRegistryClient) -> None:
        self._registry = registry
        self._lock = threading.Lock()
        self._by_id: dict[int, SchemaInfo] = {}
        self._subjects_seen: set[str] = set()

    def resolve(self, schema_id: int) -> SchemaInfo | None:
        with self._lock:
            if schema_id in self._by_id:
                return self._by_id[schema_id]

        try:
            schema_obj = self._registry.get_schema(schema_id)
        except Exception as exc:
            logger.warning("failed to fetch schema id=%d from registry: %s", schema_id, exc)
            return None

        subject, version = self._resolve_subject_and_version(schema_id, schema_obj)

        info = SchemaInfo(
            schema_id=schema_id,
            version=version,
            subject=subject,
            schema_str=schema_obj.schema_str,
        )

        with self._lock:
            self._by_id[schema_id] = info
            if subject:
                self._subjects_seen.add(subject)

        logger.info(
            "resolved schema id=%d subject=%s version=%d",
            schema_id,
            subject,
            version,
        )
        return info

    def _resolve_subject_and_version(self, schema_id: int, schema_obj: Schema) -> tuple[str, int]:
        subjects = self._registry.get_subjects()
        for subject in subjects:
            if subject.endswith("-value"):
                try:
                    versions = self._registry.get_versions(subject)
                    for v in versions:
                        sid = self._registry.get_version(subject, v, False).schema_id
                        if sid == schema_id:
                            return subject, v
                except Exception:
                    continue
        return "", 0

    @property
    def known_schemas(self) -> dict[int, SchemaInfo]:
        with self._lock:
            return dict(self._by_id)


class AvroDecoder:
    def __init__(self, schema_registry_url: str):
        self._registry = SchemaRegistryClient({"url": schema_registry_url})
        self._deserializer = AvroDeserializer(self._registry)
        self._schema_cache = SchemaVersionCache(self._registry)
        self._known_schema_ids: set[int] = set()

    @property
    def schema_cache(self) -> SchemaVersionCache:
        return self._schema_cache

    def decode(self, msg) -> DeserializedRecord | None:
        if msg is None or msg.value() is None:
            return None

        raw = msg.value()
        if len(raw) < _HEADER_SIZE or raw[0] != _CONFLUENT_MAGIC:
            logger.warning(
                "invalid confluent wire format topic=%s partition=%d offset=%d len=%d",
                msg.topic(),
                msg.partition(),
                msg.offset(),
                len(raw) if raw else 0,
            )
            return None

        schema_id = struct.unpack(">I", raw[1:_HEADER_SIZE])[0]
        topic = msg.topic()
        partition = msg.partition()
        offset = msg.offset()

        ctx = SerializationContext(topic, MessageField.VALUE)
        try:
            data = self._deserializer(raw, ctx)
        except Exception as exc:
            logger.warning(
                "deserialization failed topic=%s partition=%d offset=%d schema_id=%d: %s",
                topic,
                partition,
                offset,
                schema_id,
                exc,
            )
            return None

        if data is None:
            return None

        self._track_schema_id(schema_id, topic)

        schema_info = self._schema_cache.resolve(schema_id)
        schema_version = schema_info.version if schema_info else 0

        envelope = Envelope.from_dict(data)

        return DeserializedRecord(
            envelope=envelope,
            topic=topic,
            partition=partition,
            offset=offset,
            key=msg.key().decode("utf-8") if msg.key() else None,
            schema_id=schema_id,
            schema_version=schema_version,
        )

    def _track_schema_id(self, schema_id: int, topic: str) -> None:
        if schema_id not in self._known_schema_ids:
            self._known_schema_ids.add(schema_id)
            logger.info("new schema version detected topic=%s schema_id=%d", topic, schema_id)
