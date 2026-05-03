from __future__ import annotations

import io
import logging
from typing import Any

import fastavro

from src.filewriter.schema import build_event_schema

logger = logging.getLogger("acc-dp-consumer.filewriter")

_TOPIC_EVENT_NAMES: dict[str, str] = {
    "acc.physics.v1": "acc_physics_event",
    "acc.graphics.v1": "acc_graphics_event",
    "acc.static.v1": "acc_static_event",
}


class AvroWriter:
    def __init__(self) -> None:
        self._schema_cache: dict[str, Any] = {}

    def write_window(self, topic: str, records: list[dict]) -> bytes:
        if not records:
            return b""

        schema = self._resolve_schema(topic, records)
        normalized = [_normalize(r) for r in records]

        buffer = io.BytesIO()
        try:
            fastavro.writer(buffer, schema, normalized, codec="deflate")
        except Exception as exc:
            logger.error(
                "avro write failed topic=%s records=%d: %s",
                topic,
                len(normalized),
                exc,
            )
            raise

        result = buffer.getvalue()
        logger.debug(
            "wrote avro topic=%s records=%d bytes=%d schema_fields=%d",
            topic,
            len(normalized),
            len(result),
            len(schema.get("fields", [])),
        )
        return result

    def _resolve_schema(self, topic: str, records: list[dict]) -> dict[str, Any]:
        if topic in self._schema_cache:
            return self._schema_cache[topic]

        event_name = _TOPIC_EVENT_NAMES.get(topic, topic.replace(".", "_"))
        schema = build_event_schema(event_name, records)
        parsed = fastavro.parse_schema(schema)
        self._schema_cache[topic] = parsed

        field_count = len(schema.get("fields", []))
        logger.info("built schema for topic=%s event=%s fields=%d", topic, event_name, field_count)
        return parsed


def _normalize(record: dict) -> dict[str, Any]:
    out: dict[str, Any] = {}
    for key, value in record.items():
        out[key] = _normalize_value(value)
    return out


def _normalize_value(value: Any) -> Any:
    if value is None:
        return None
    if isinstance(value, bool):
        return value
    if isinstance(value, int):
        return value
    if isinstance(value, float):
        return value
    if isinstance(value, str):
        return value
    if isinstance(value, list):
        return [_normalize_value(v) for v in value]
    if isinstance(value, dict):
        return {k: _normalize_value(v) for k, v in value.items()}
    return str(value)
