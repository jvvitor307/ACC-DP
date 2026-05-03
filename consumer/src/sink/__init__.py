from __future__ import annotations

import io
import logging

from minio import Minio
from minio.error import S3Error

from src.config.settings import MinioSettings
from src.filewriter import AvroWriter
from src.window import Window

logger = logging.getLogger("acc-dp-consumer.sink")


class MinioSink:
    def __init__(self, settings: MinioSettings):
        self._client = Minio(
            endpoint=settings.endpoint,
            access_key=settings.access_key,
            secret_key=settings.secret_key,
            secure=settings.secure,
        )
        self._bucket = settings.bucket
        self._writer = AvroWriter()
        self._ensure_bucket()

    def _ensure_bucket(self) -> None:
        if not self._client.bucket_exists(self._bucket):
            self._client.make_bucket(self._bucket)
            logger.info("created bucket %s", self._bucket)

    def upload_window(self, window: Window) -> str | None:
        if not window.records:
            return None

        object_key = window.object_key

        try:
            avro_bytes = self._writer.write_window(window.topic, window.records)
        except Exception as exc:
            logger.error("failed to encode avro for %s: %s", object_key, exc)
            return None

        if not avro_bytes:
            return None

        try:
            self._client.put_object(
                self._bucket,
                object_key,
                io.BytesIO(avro_bytes),
                length=len(avro_bytes),
                content_type="application/octet-stream",
            )
            logger.info(
                "uploaded %s (%d records, %d bytes, schema_versions=%s)",
                object_key,
                len(window.records),
                len(avro_bytes),
                window.schema_versions,
            )
            return object_key
        except S3Error as exc:
            logger.error("failed to upload %s: %s", object_key, exc)
            return None
