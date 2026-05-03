from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Envelope:
    event_id: str
    event_time: int
    ingestion_time: int
    session_id: str
    usuario_id: str
    username: str
    source: str
    payload: dict

    @staticmethod
    def from_dict(data: dict) -> Envelope:
        return Envelope(
            event_id=data.get("event_id", ""),
            event_time=data.get("event_time", 0),
            ingestion_time=data.get("ingestion_time", 0),
            session_id=data.get("session_id", ""),
            usuario_id=data.get("usuario_id", ""),
            username=data.get("username", ""),
            source=data.get("source", ""),
            payload=data.get("payload", {}),
        )

    def to_dict(self) -> dict:
        return {
            "event_id": self.event_id,
            "event_time": self.event_time,
            "ingestion_time": self.ingestion_time,
            "session_id": self.session_id,
            "usuario_id": self.usuario_id,
            "username": self.username,
            "source": self.source,
            "payload": self.payload,
        }
