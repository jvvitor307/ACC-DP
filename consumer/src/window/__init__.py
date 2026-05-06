from __future__ import annotations

import time
from dataclasses import dataclass, field
from datetime import datetime, timezone

from src.avro import DeserializedRecord


@dataclass
class Window:
    topic: str
    usuario_id: str
    session_id: str
    start: datetime
    end: datetime
    records: list[dict] = field(default_factory=list)
    schema_versions: set[int] = field(default_factory=set)

    @property
    def table_path(self) -> str:
        return f"bronze/{self._topic_label()}"

    @property
    def partition_cols(self) -> list[str]:
        return ["usuario_id", "session_id", "year", "month", "day", "hour"]

    def enriched_records(self) -> list[dict]:
        result: list[dict] = []
        for record in self.records:
            row = dict(record)
            row["usuario_id"] = self.usuario_id
            row["session_id"] = self.session_id
            row["year"] = self.start.year
            row["month"] = self.start.month
            row["day"] = self.start.day
            row["hour"] = self.start.hour
            result.append(row)
        return result

    def _topic_label(self) -> str:
        mapping = {
            "acc.physics.v1": "acc_physics",
            "acc.graphics.v1": "acc_graphics",
            "acc.static.v1": "acc_statics",
        }
        return mapping.get(self.topic, self.topic.replace(".", "_"))


class WindowManager:
    def __init__(self, window_duration_seconds: int):
        self._duration = window_duration_seconds
        self._buckets: dict[tuple[str, str, str], Window] = {}

    def add(self, record: DeserializedRecord) -> None:
        usuario_id = record.envelope.usuario_id or "unknown"
        session_id = record.envelope.session_id or "no_session"
        event_time_ms = record.envelope.event_time
        event_dt = datetime.fromtimestamp(event_time_ms / 1000.0, tz=timezone.utc)
        window_start = self._truncate(event_dt)
        window_end = datetime.fromtimestamp(
            window_start.timestamp() + self._duration, tz=timezone.utc
        )

        key = (record.topic, usuario_id, session_id)
        if key not in self._buckets:
            self._buckets[key] = Window(
                topic=record.topic,
                usuario_id=usuario_id,
                session_id=session_id,
                start=window_start,
                end=window_end,
            )

        self._buckets[key].records.append(record.value)
        self._buckets[key].schema_versions.add(record.schema_version)

    def flush_ready(self) -> list[Window]:
        now = time.time()
        ready: list[Window] = []
        remaining: dict[tuple[str, str, str], Window] = {}

        for key, window in self._buckets.items():
            if window.end.timestamp() <= now:
                ready.append(window)
            else:
                remaining[key] = window

        self._buckets = remaining
        return ready

    def flush_all(self) -> list[Window]:
        windows = list(self._buckets.values())
        self._buckets.clear()
        return windows

    def _truncate(self, dt: datetime) -> datetime:
        ts = dt.timestamp()
        truncated = int(ts // self._duration) * self._duration
        return datetime.fromtimestamp(truncated, tz=timezone.utc)
