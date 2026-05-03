from __future__ import annotations

import io
import json
import logging

import avro.io
import avro.schema
from minio import Minio
from minio.error import S3Error

from src.config.settings import MinioSettings
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
        self._ensure_bucket()

    def _ensure_bucket(self) -> None:
        if not self._client.bucket_exists(self._bucket):
            self._client.make_bucket(self._bucket)
            logger.info("created bucket %s", self._bucket)

    def upload_window(self, window: Window) -> str | None:
        if not window.records:
            return None

        avro_bytes = self._encode_avro(window.records)
        object_key = window.object_key

        try:
            self._client.put_object(
                self._bucket,
                object_key,
                io.BytesIO(avro_bytes),
                length=len(avro_bytes),
                content_type="application/octet-stream",
            )
            logger.info(
                "uploaded %s (%d records, %d bytes)",
                object_key,
                len(window.records),
                len(avro_bytes),
            )
            return object_key
        except S3Error as exc:
            logger.error("failed to upload %s: %s", object_key, exc)
            return None

    def _encode_avro(self, records: list[dict]) -> bytes:
        schema = self._infer_schema(records[0])
        parsed = avro.schema.parse(json.dumps(schema))
        buffer = io.BytesIO()
        writer = avro.io.DatumWriter(parsed)
        encoder = avro.io.BinaryEncoder(buffer)
        for record in records:
            writer.write(record, encoder)
        return buffer.getvalue()

    @staticmethod
    def _infer_schema(record: dict) -> dict:
        fields = []
        for name, value in record.items():
            avro_type = _python_to_avro(value)
            fields.append({"name": name, "type": avro_type})
        return {"type": "record", "name": "window_record", "fields": fields}


def _python_to_avro(value) -> str | dict:
    if isinstance(value, bool):
        return "boolean"
    if isinstance(value, int):
        return "long"
    if isinstance(value, float):
        return "double"
    if isinstance(value, str):
        return "string"
    if isinstance(value, list):
        return {"type": "array", "items": "long"}
    if isinstance(value, dict):
        fields = []
        for k, v in value.items():
            fields.append({"name": k, "type": _python_to_avro(v)})
        return {"type": "record", "name": "nested", "fields": fields}
    return "string"
