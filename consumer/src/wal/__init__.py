from __future__ import annotations

import json
import logging
import os
import time
from dataclasses import asdict
from datetime import datetime, timezone
from pathlib import Path

from src.window import Window

logger = logging.getLogger("acc-dp-consumer.wal")

_DEFAULT_DIR = Path("wal_data")


def persist(window: Window, wal_dir: Path = _DEFAULT_DIR) -> Path:
    wal_dir.mkdir(parents=True, exist_ok=True)

    ts = int(time.time() * 1000)
    safe_topic = window.topic.replace(".", "_")
    filename = f"{safe_topic}_{window.usuario_id}_{ts}.wal"
    filepath = wal_dir / filename

    data = _serialize_window(window)
    tmp = filepath.with_suffix(".tmp")
    tmp.write_text(json.dumps(data, default=str), encoding="utf-8")
    os.replace(tmp, filepath)

    logger.debug("persisted WAL %s (%d records)", filepath.name, len(window.records))
    return filepath


def recover_all(wal_dir: Path = _DEFAULT_DIR) -> list[Window]:
    if not wal_dir.exists():
        return []

    windows: list[Window] = []
    for filepath in sorted(wal_dir.glob("*.wal")):
        try:
            data = json.loads(filepath.read_text(encoding="utf-8"))
            windows.append(_deserialize_window(data))
            logger.info("recovered WAL %s (%d records)", filepath.name, len(data.get("records", [])))
        except Exception as exc:
            logger.error("failed to recover WAL %s: %s", filepath.name, exc)

    return windows


def remove(filepath: Path) -> None:
    try:
        filepath.unlink()
        logger.debug("removed WAL %s", filepath.name)
    except FileNotFoundError:
        pass


def cleanup_all(wal_dir: Path = _DEFAULT_DIR) -> None:
    if not wal_dir.exists():
        return
    for filepath in wal_dir.glob("*.wal"):
        filepath.unlink(missing_ok=True)
    logger.info("cleaned up WAL directory %s", wal_dir)


def _serialize_window(window: Window) -> dict:
    return {
        "topic": window.topic,
        "usuario_id": window.usuario_id,
        "session_id": window.session_id,
        "start": window.start.isoformat(),
        "end": window.end.isoformat(),
        "records": window.records,
        "schema_versions": sorted(window.schema_versions),
    }


def _deserialize_window(data: dict) -> Window:
    return Window(
        topic=data["topic"],
        usuario_id=data["usuario_id"],
        session_id=data["session_id"],
        start=datetime.fromisoformat(data["start"]),
        end=datetime.fromisoformat(data["end"]),
        records=data["records"],
        schema_versions=set(data.get("schema_versions", [])),
    )
