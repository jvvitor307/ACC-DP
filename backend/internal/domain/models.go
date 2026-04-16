package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Machine struct {
	ID         uuid.UUID
	MachineUID string
	DeviceName string
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type Session struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	MachineID    uuid.UUID
	CreatedAt    time.Time
	ExpiresAt    time.Time
	RevokedAt    *time.Time
	RevokedReason string
}

type RefreshToken struct {
	ID                uuid.UUID
	SessionID         uuid.UUID
	UserID            uuid.UUID
	TokenHash         string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	RevokedAt         *time.Time
	RevokedReason     string
	ReplacedByTokenID *uuid.UUID
}

type RefreshTokenRecord struct {
	RefreshTokenID   uuid.UUID
	SessionID        uuid.UUID
	UserID           uuid.UUID
	MachineUID       string
	RefreshExpiresAt time.Time
	RefreshRevokedAt *time.Time
	SessionExpiresAt time.Time
	SessionRevokedAt *time.Time
}
