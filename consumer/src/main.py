from __future__ import annotations

import logging
import signal
import sys
from types import FrameType

from src.config import ConfigError, Settings
from src.config.settings import load_settings
from src.consumer import TopicConsumer
from src.logging_setup import setup
from src.sink import MinioSink
from src.window import WindowManager

logger = logging.getLogger("acc-dp-consumer")

_COMMIT_INTERVAL_RECORDS = 500


def run(settings: Settings) -> None:
    consumer = TopicConsumer(settings.kafka)
    window_mgr = WindowManager(settings.window.duration_seconds)
    sink = MinioSink(settings.minio)

    consumer.subscribe()
    consumer.running = True

    if not consumer.wait_for_assignment(timeout=30.0):
        logger.error("timed out waiting for partition assignment")
        consumer.close()
        return

    def _shutdown(signum: int, frame: FrameType | None) -> None:
        logger.info("received signal %s, shutting down...", signum)
        consumer.running = False

    signal.signal(signal.SIGINT, _shutdown)
    signal.signal(signal.SIGTERM, _shutdown)

    logger.info(
        "consumer started — topics=%s window=%ds",
        settings.kafka.topics,
        settings.window.duration_seconds,
    )

    records_since_commit = 0

    try:
        while consumer.running:
            record = consumer.poll(timeout=1.0)
            if record is None:
                _flush_ready_windows(window_mgr, sink)
                continue

            window_mgr.add(record)
            records_since_commit += 1

            _flush_ready_windows(window_mgr, sink)

            if records_since_commit >= _COMMIT_INTERVAL_RECORDS:
                consumer.commit(asynchronous=True)
                records_since_commit = 0
    finally:
        logger.info("flushing remaining windows...")
        for window in window_mgr.flush_all():
            sink.upload_window(window)
        consumer.commit(asynchronous=False)
        consumer.close()
        logger.info("shutdown complete")


def _flush_ready_windows(window_mgr: WindowManager, sink: MinioSink) -> None:
    for window in window_mgr.flush_ready():
        sink.upload_window(window)


def main() -> None:
    try:
        settings = load_settings(env_file=None)
    except ConfigError as exc:
        setup("ERROR")
        logging.getLogger("acc-dp-consumer").error("startup aborted:\n%s", exc)
        sys.exit(1)

    log = setup(settings.log.level)
    log.info("starting ACC-DP consumer")

    try:
        run(settings)
    except Exception:
        log.exception("consumer failed")
        sys.exit(1)


if __name__ == "__main__":
    main()
