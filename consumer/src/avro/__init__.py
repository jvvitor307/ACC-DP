from __future__ import annotations

from dataclasses import dataclass

from confluent_kafka.schema_registry import SchemaRegistryClient
from confluent_kafka.schema_registry.avro import AvroDeserializer
from confluent_kafka.serialization import SerializationContext, MessageField


@dataclass
class DeserializedRecord:
    value: dict
    topic: str
    partition: int
    offset: int
    key: str | None


class AvroDecoder:
    def __init__(self, schema_registry_url: str):
        self._registry = SchemaRegistryClient({"url": schema_registry_url})
        self._deserializer = AvroDeserializer(self._registry)

    def decode(self, msg) -> DeserializedRecord | None:
        if msg is None or msg.value() is None:
            return None

        topic = msg.topic()
        ctx = SerializationContext(topic, MessageField.VALUE)
        try:
            value = self._deserializer(msg.value(), ctx)
        except Exception:
            return None

        if value is None:
            return None

        return DeserializedRecord(
            value=value,
            topic=topic,
            partition=msg.partition(),
            offset=msg.offset(),
            key=msg.key().decode("utf-8") if msg.key() else None,
        )
