from __future__ import annotations

from typing import Any


def build_event_schema(event_name: str, records: list[dict]) -> dict[str, Any]:
    if not records:
        return _fallback_schema()

    sample = records[0]
    payload = sample.get("payload", {})

    payload_schema = _infer_record_schema(f"{event_name}_payload", payload)
    event_schema = _infer_event_schema(event_name, sample, payload_schema)

    return event_schema


def _infer_event_schema(name: str, record: dict, payload_schema: dict) -> dict[str, Any]:
    fields: list[dict[str, Any]] = []

    header_fields = [
        ("event_id", "string", None),
        ("event_time", "long", None),
        ("ingestion_time", "long", None),
        ("session_id", "string", ""),
        ("usuario_id", "string", None),
        ("username", "string", None),
        ("source", "string", None),
    ]

    for field_name, avro_type, default in header_fields:
        field_def: dict[str, Any] = {"name": field_name, "type": avro_type}
        if default is not None:
            field_def["default"] = default
        fields.append(field_def)

    fields.append({"name": "payload", "type": payload_schema})

    for key, value in record.items():
        if key in ("event_id", "event_time", "ingestion_time", "session_id",
                    "usuario_id", "username", "source", "payload"):
            continue
        fields.append({"name": key, "type": _infer_type(value)})

    return {"type": "record", "name": name, "fields": fields}


def _infer_record_schema(name: str, record: dict) -> dict[str, Any]:
    fields: list[dict[str, Any]] = []
    for key, value in record.items():
        fields.append({"name": key, "type": _infer_type(value)})
    return {"type": "record", "name": name, "fields": fields}


def _infer_type(value: Any) -> Any:
    if value is None:
        return ["null", "string"]
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, int):
        return "long"
    if isinstance(value, float):
        return "double"
    if isinstance(value, str):
        return "string"
    if isinstance(value, list):
        if not value:
            return {"type": "array", "items": "long"}
        return {"type": "array", "items": _infer_type(value[0])}
    if isinstance(value, dict):
        return _infer_record_schema("nested", value)
    return "string"


def _fallback_schema() -> dict[str, Any]:
    return {
        "type": "record",
        "name": "acc_event_fallback",
        "fields": [
            {"name": "event_id", "type": "string"},
            {"name": "event_time", "type": "long"},
            {"name": "ingestion_time", "type": "long"},
            {"name": "session_id", "type": ["null", "string"], "default": None},
            {"name": "usuario_id", "type": "string"},
            {"name": "username", "type": "string"},
            {"name": "source", "type": "string"},
            {"name": "payload", "type": {"type": "map", "values": ["null", "long", "double", "string"]}},
        ],
    }
