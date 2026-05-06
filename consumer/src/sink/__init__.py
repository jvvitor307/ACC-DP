from __future__ import annotations

import logging
import time

import pyarrow as pa
from deltalake import write_deltalake
from minio import Minio
from minio.error import S3Error

from src.config.settings import MinioSettings
from src.filewriter import DeltaWriter
from src.window import Window

logger = logging.getLogger("acc-dp-consumer.sink")

_MAX_RETRIES = 3
_RETRY_BASE_DELAY = 1.0


class DeltaSink:
    def __init__(self, settings: MinioSettings):
        self._bucket = settings.bucket
        self._writer = DeltaWriter()
        self._written_tables: set[str] = set()
        self._storage_options = _build_storage_options(settings)
        self._minio = Minio(
            endpoint=settings.endpoint,
            access_key=settings.access_key,
            secret_key=settings.secret_key,
            secure=settings.secure,
        )
        self._ensure_bucket()

    def _ensure_bucket(self) -> None:
        try:
            if not self._minio.bucket_exists(self._bucket):
                self._minio.make_bucket(self._bucket)
                logger.info("created bucket %s", self._bucket)
        except S3Error as exc:
            logger.error("bucket check/create failed for %s: %s", self._bucket, exc)
            raise

    def write_window(self, window: Window) -> str | None:
        if not window.records:
            return None

        table_uri = f"s3://{self._bucket}/{window.table_path}"
        enriched = window.enriched_records()

        try:
            table = self._writer.build_table(window.topic, enriched)
        except Exception as exc:
            logger.error("failed to build arrow table for %s: %s", table_uri, exc)
            return None

        if not self._write_with_retry(table_uri, table, window.partition_cols):
            return None

        self._written_tables.add(table_uri)

        logger.info(
            "wrote delta %s (%d records, %d cols, schema_versions=%s)",
            table_uri,
            table.num_rows,
            table.num_columns,
            window.schema_versions,
        )
        return table_uri

    def _write_with_retry(self, table_uri: str, table: pa.Table, partition_cols: list[str]) -> bool:
        for attempt in range(1, _MAX_RETRIES + 1):
            try:
                write_deltalake(
                    table_or_uri=table_uri,
                    data=table,
                    mode="append",
                    partition_by=partition_cols,
                    storage_options=self._storage_options,
                )
                return True
            except Exception as exc:
                delay = _RETRY_BASE_DELAY * (2 ** (attempt - 1))
                logger.warning(
                    "delta write attempt %d/%d failed for %s (retry in %.1fs): %s",
                    attempt,
                    _MAX_RETRIES,
                    table_uri,
                    delay,
                    exc,
                )
                if attempt < _MAX_RETRIES:
                    time.sleep(delay)

        logger.error("delta write exhausted %d retries for %s", _MAX_RETRIES, table_uri)
        return False

    def health_check(self) -> bool:
        try:
            return self._minio.bucket_exists(self._bucket)
        except Exception:
            return False


def _build_storage_options(settings: MinioSettings) -> dict[str, str]:
    scheme = "https" if settings.secure else "http"
    return {
        "AWS_ACCESS_KEY_ID": settings.access_key,
        "AWS_SECRET_ACCESS_KEY": settings.secret_key,
        "AWS_ENDPOINT_URL": f"{scheme}://{settings.endpoint}",
        "AWS_ALLOW_HTTP": "true" if not settings.secure else "false",
        "AWS_S3_ALLOW_UNSAFE_WRITE": "true",
    }
