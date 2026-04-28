package event

import "acc-dp/producer/internal/source/acc_shm"

type PhysicsEvent struct {
	EventID       string              `avro:"event_id"`
	EventTime     int64               `avro:"event_time"`
	IngestionTime int64               `avro:"ingestion_time"`
	UsuarioID     string              `avro:"usuario_id"`
	Username      string              `avro:"username"`
	Source        string              `avro:"source"`
	Payload       acc_shm.PhysicsPage `avro:"payload"`
}

type GraphicsEvent struct {
	EventID       string               `avro:"event_id"`
	EventTime     int64                `avro:"event_time"`
	IngestionTime int64                `avro:"ingestion_time"`
	UsuarioID     string               `avro:"usuario_id"`
	Username      string               `avro:"username"`
	Source        string               `avro:"source"`
	Payload       acc_shm.GraphicsPage `avro:"payload"`
}

type StaticEvent struct {
	EventID       string             `avro:"event_id"`
	EventTime     int64              `avro:"event_time"`
	IngestionTime int64              `avro:"ingestion_time"`
	UsuarioID     string             `avro:"usuario_id"`
	Username      string             `avro:"username"`
	Source        string             `avro:"source"`
	Payload       acc_shm.StaticPage `avro:"payload"`
}
