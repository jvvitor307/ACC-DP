from __future__ import annotations

from typing import Any

import pyarrow as pa


def infer_pa_type(value: Any) -> pa.DataType:
    if value is None:
        return pa.string()
    if isinstance(value, bool):
        return pa.bool_()
    if isinstance(value, int):
        return pa.int64()
    if isinstance(value, float):
        return pa.float64()
    if isinstance(value, str):
        return pa.string()
    if isinstance(value, list):
        if not value:
            return pa.list_(pa.int64())
        return pa.list_(infer_pa_type(value[0]))
    if isinstance(value, dict):
        return pa.struct([(k, infer_pa_type(v)) for k, v in value.items()])
    return pa.string()


def build_pyarrow_schema(records: list[dict]) -> pa.Schema:
    if not records:
        return _fallback_schema()

    all_keys: dict[str, pa.DataType] = {}
    for record in records:
        for key, value in record.items():
            if key not in all_keys:
                all_keys[key] = infer_pa_type(value)

    return pa.schema([pa.field(k, v) for k, v in all_keys.items()])


def _fallback_schema() -> pa.Schema:
    return pa.schema([
        pa.field("event_id", pa.string()),
        pa.field("event_time", pa.int64()),
        pa.field("ingestion_time", pa.int64()),
        pa.field("session_id", pa.string()),
        pa.field("usuario_id", pa.string()),
        pa.field("username", pa.string()),
        pa.field("source", pa.string()),
        pa.field("payload", pa.string()),
        pa.field("year", pa.int32()),
        pa.field("month", pa.int32()),
        pa.field("day", pa.int32()),
        pa.field("hour", pa.int32()),
    ])
