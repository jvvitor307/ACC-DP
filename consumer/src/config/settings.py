from __future__ import annotations

import os
import re
from dataclasses import dataclass
from pathlib import Path

from dotenv import load_dotenv

_REQUIRED_KEYS = [
    "REDPANDA_BROKERS",
    "SCHEMA_REGISTRY_URL",
    "TOPIC_PHYSICS",
    "TOPIC_GRAPHICS",
    "TOPIC_STATIC",
    "MINIO_ENDPOINT",
    "MINIO_ACCESS_KEY",
    "MINIO_SECRET_KEY",
    "MINIO_BUCKET",
]

_VALID_OFFSET_RESET = {"earliest", "latest", "none"}
_VALID_LOG_LEVELS = {"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
_DURATION_RE = re.compile(r"^(\d+)(ms|s|m|h)$")


@dataclass(frozen=True)
class KafkaSettings:
    brokers: str
    group_id: str
    topics: list[str]
    schema_registry_url: str
    session_timeout_ms: int
    auto_offset_reset: str
    enable_auto_commit: bool


@dataclass(frozen=True)
class MinioSettings:
    endpoint: str
    access_key: str
    secret_key: str
    bucket: str
    secure: bool


@dataclass(frozen=True)
class WindowSettings:
    duration_seconds: int


@dataclass(frozen=True)
class LogSettings:
    level: str


@dataclass(frozen=True)
class Settings:
    kafka: KafkaSettings
    minio: MinioSettings
    window: WindowSettings
    log: LogSettings


class ConfigError(Exception):
    def __init__(self, errors: list[str]) -> None:
        self.errors = errors
        super().__init__("invalid configuration:\n  - " + "\n  - ".join(errors))


def _collect_required(env: dict[str, str], errors: list[str]) -> None:
    missing = [k for k in _REQUIRED_KEYS if not env.get(k)]
    if missing:
        errors.append(f"missing required environment variables: {', '.join(missing)}")


def _parse_duration(raw: str) -> int:
    raw = raw.strip()
    match = _DURATION_RE.match(raw)
    if not match:
        raise ValueError(f"invalid duration format: {raw!r} (expected e.g. 5m, 300s, 1h)")
    value = int(match.group(1))
    unit = match.group(2)
    multipliers = {"ms": 0.001, "s": 1, "m": 60, "h": 3600}
    return int(value * multipliers[unit])


def _validate_kafka(env: dict[str, str], errors: list[str]) -> None:
    brokers = env.get("REDPANDA_BROKERS", "")
    if brokers:
        parts = [b.strip() for b in brokers.split(",") if b.strip()]
        if not parts:
            errors.append("REDPANDA_BROKERS must contain at least one broker address")
        for part in parts:
            if ":" not in part:
                errors.append(f"REDPANDA_BROKERS broker missing port: {part!r}")

    sr_url = env.get("SCHEMA_REGISTRY_URL", "")
    if sr_url and not sr_url.startswith(("http://", "https://")):
        errors.append("SCHEMA_REGISTRY_URL must start with http:// or https://")

    offset = env.get("CONSUMER_AUTO_OFFSET_RESET", "earliest").lower()
    if offset not in _VALID_OFFSET_RESET:
        errors.append(f"CONSUMER_AUTO_OFFSET_RESET must be one of {sorted(_VALID_OFFSET_RESET)}, got {offset!r}")

    timeout_raw = env.get("CONSUMER_SESSION_TIMEOUT_MS", "30000")
    try:
        timeout = int(timeout_raw)
        if timeout <= 0:
            errors.append("CONSUMER_SESSION_TIMEOUT_MS must be greater than zero")
    except ValueError:
        errors.append(f"CONSUMER_SESSION_TIMEOUT_MS must be an integer, got {timeout_raw!r}")


def _validate_minio(env: dict[str, str], errors: list[str]) -> None:
    endpoint = env.get("MINIO_ENDPOINT", "")
    if endpoint and ":" not in endpoint:
        errors.append("MINIO_ENDPOINT must include a port (e.g. localhost:19000)")

    bucket = env.get("MINIO_BUCKET", "")
    if bucket and not re.match(r"^[a-z0-9][a-z0-9.\-]{1,61}[a-z0-9]$", bucket):
        errors.append(f"MINIO_BUCKET has invalid bucket name: {bucket!r}")


def _validate_window(env: dict[str, str], errors: list[str]) -> None:
    raw = env.get("CONSUMER_WINDOW", "300")
    try:
        seconds = _parse_duration(raw) if _DURATION_RE.match(raw) else int(raw)
        if seconds <= 0:
            errors.append("CONSUMER_WINDOW must be greater than zero")
    except ValueError:
        errors.append(f"CONSUMER_WINDOW must be a positive integer or duration (e.g. 5m), got {raw!r}")


def _validate_log(env: dict[str, str], errors: list[str]) -> None:
    level = env.get("CONSUMER_LOG_LEVEL", "INFO").upper()
    if level not in _VALID_LOG_LEVELS:
        errors.append(f"CONSUMER_LOG_LEVEL must be one of {sorted(_VALID_LOG_LEVELS)}, got {level!r}")


def _load_env(env_file: str | Path | None) -> None:
    if env_file is None:
        candidates = [Path("../../.env"), Path("../.env"), Path(".env")]
    else:
        candidates = [Path(env_file)]

    for candidate in candidates:
        resolved = candidate.resolve()
        if resolved.exists():
            load_dotenv(resolved, override=False)
            return


def load_settings(env_file: str | Path | None = None) -> Settings:
    _load_env(env_file)

    env: dict[str, str] = {}
    for key in os.environ:
        env[key] = os.environ[key].strip()

    errors: list[str] = []

    _collect_required(env, errors)
    _validate_kafka(env, errors)
    _validate_minio(env, errors)
    _validate_window(env, errors)
    _validate_log(env, errors)

    if errors:
        raise ConfigError(errors)

    window_raw = env.get("CONSUMER_WINDOW", "300")
    if _DURATION_RE.match(window_raw):
        window_seconds = _parse_duration(window_raw)
    else:
        window_seconds = int(window_raw)

    return Settings(
        kafka=KafkaSettings(
            brokers=env["REDPANDA_BROKERS"],
            group_id=env.get("CONSUMER_GROUP_ID", "acc-dp-consumer"),
            topics=[env["TOPIC_PHYSICS"], env["TOPIC_GRAPHICS"], env["TOPIC_STATIC"]],
            schema_registry_url=env["SCHEMA_REGISTRY_URL"],
            session_timeout_ms=int(env.get("CONSUMER_SESSION_TIMEOUT_MS", "30000")),
            auto_offset_reset=env.get("CONSUMER_AUTO_OFFSET_RESET", "earliest"),
            enable_auto_commit=env.get("CONSUMER_ENABLE_AUTO_COMMIT", "false").lower() == "true",
        ),
        minio=MinioSettings(
            endpoint=env["MINIO_ENDPOINT"],
            access_key=env["MINIO_ACCESS_KEY"],
            secret_key=env["MINIO_SECRET_KEY"],
            bucket=env["MINIO_BUCKET"],
            secure=env.get("MINIO_USE_SSL", "false").lower() == "true",
        ),
        window=WindowSettings(duration_seconds=window_seconds),
        log=LogSettings(level=env.get("CONSUMER_LOG_LEVEL", "INFO").upper()),
    )
