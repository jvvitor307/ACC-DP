from __future__ import annotations

import json
import logging
from typing import Any

import pyarrow as pa

logger = logging.getLogger("acc-dp-consumer.filewriter")


class DeltaWriter:
    def __init__(self) -> None:
        self._schema_cache: dict[str, pa.Schema] = {}

    def build_table(self, topic: str, records: list[dict]) -> pa.Table:
        if not records:
            raise ValueError("cannot build table from empty records")

        rows = [_normalize_row(r) for r in records]

        if topic not in self._schema_cache:
            self._schema_cache[topic] = pa.Table.from_pylist(rows[:1]).schema

        table = pa.Table.from_pylist(rows, schema=self._schema_cache[topic])

        logger.debug(
            "built arrow table topic=%s rows=%d cols=%d",
            topic,
            table.num_rows,
            table.num_columns,
        )
        return table


def _normalize_row(record: dict) -> dict[str, Any]:
    out: dict[str, Any] = {}
    for key, value in record.items():
        if key == "payload":
            out[key] = json.dumps(value, default=str) if isinstance(value, (dict, list)) else str(value)
        else:
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
