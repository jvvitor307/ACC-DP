package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"acc-dp/backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	return tx, nil
}

func (r *AuthRepository) CreateUser(ctx context.Context, tx *sql.Tx, user *domain.User) error {
	const query = `
		INSERT INTO users (id, email, display_name, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`

	_, err := tx.ExecContext(ctx, query, user.ID, user.Email, user.DisplayName, user.PasswordHash, user.CreatedAt)
	if isUniqueViolation(err) {
		return domain.ErrUserAlreadyExists
	}
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, email, display_name, password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.DisplayName,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvalidCredentials
	}
	if err != nil {
		return nil, fmt.Errorf("select user by email: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) UpsertMachine(ctx context.Context, tx *sql.Tx, machineUID, deviceName string, now time.Time) (*domain.Machine, error) {
	const query = `
		INSERT INTO machines (id, machine_uid, device_name, created_at, last_seen_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $4)
		ON CONFLICT (machine_uid)
		DO UPDATE SET
			device_name = COALESCE(NULLIF(EXCLUDED.device_name, ''), machines.device_name),
			last_seen_at = EXCLUDED.last_seen_at
		RETURNING id, machine_uid, COALESCE(device_name, ''), created_at, last_seen_at
	`

	machine := &domain.Machine{}
	err := tx.QueryRowContext(ctx, query, uuid.New(), machineUID, deviceName, now).Scan(
		&machine.ID,
		&machine.MachineUID,
		&machine.DeviceName,
		&machine.CreatedAt,
		&machine.LastSeenAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert machine: %w", err)
	}

	return machine, nil
}

func (r *AuthRepository) RevokeActiveSessionsByUser(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time, reason string) error {
	const query = `
		WITH revoked AS (
			UPDATE sessions
			SET revoked_at = $2, revoked_reason = $3
			WHERE user_id = $1 AND revoked_at IS NULL
			RETURNING id
		)
		UPDATE refresh_tokens
		SET revoked_at = $2, revoked_reason = $3
		WHERE session_id IN (SELECT id FROM revoked)
		  AND revoked_at IS NULL
	`

	if _, err := tx.ExecContext(ctx, query, userID, now, reason); err != nil {
		return fmt.Errorf("revoke sessions by user: %w", err)
	}

	return nil
}

func (r *AuthRepository) RevokeActiveSessionsByMachine(ctx context.Context, tx *sql.Tx, machineID uuid.UUID, now time.Time, reason string) error {
	const query = `
		WITH revoked AS (
			UPDATE sessions
			SET revoked_at = $2, revoked_reason = $3
			WHERE machine_id = $1 AND revoked_at IS NULL
			RETURNING id
		)
		UPDATE refresh_tokens
		SET revoked_at = $2, revoked_reason = $3
		WHERE session_id IN (SELECT id FROM revoked)
		  AND revoked_at IS NULL
	`

	if _, err := tx.ExecContext(ctx, query, machineID, now, reason); err != nil {
		return fmt.Errorf("revoke sessions by machine: %w", err)
	}

	return nil
}

func (r *AuthRepository) CreateSession(ctx context.Context, tx *sql.Tx, session *domain.Session) error {
	const query = `
		INSERT INTO sessions (id, user_id, machine_id, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := tx.ExecContext(ctx, query, session.ID, session.UserID, session.MachineID, session.CreatedAt, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	return nil
}

func (r *AuthRepository) CreateRefreshToken(ctx context.Context, tx *sql.Tx, token *domain.RefreshToken) error {
	const query = `
		INSERT INTO refresh_tokens (id, session_id, user_id, token_hash, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.ExecContext(ctx, query, token.ID, token.SessionID, token.UserID, token.TokenHash, token.CreatedAt, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("insert refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetRefreshTokenRecordForUpdate(ctx context.Context, tx *sql.Tx, tokenHash string) (*domain.RefreshTokenRecord, error) {
	const query = `
		SELECT
			rt.id,
			rt.session_id,
			rt.user_id,
			m.machine_uid,
			rt.expires_at,
			rt.revoked_at,
			s.expires_at,
			s.revoked_at
		FROM refresh_tokens rt
		JOIN sessions s ON s.id = rt.session_id
		JOIN machines m ON m.id = s.machine_id
		WHERE rt.token_hash = $1
		FOR UPDATE OF rt
	`

	record := &domain.RefreshTokenRecord{}
	var refreshRevokedAt sql.NullTime
	var sessionRevokedAt sql.NullTime

	err := tx.QueryRowContext(ctx, query, tokenHash).Scan(
		&record.RefreshTokenID,
		&record.SessionID,
		&record.UserID,
		&record.MachineUID,
		&record.RefreshExpiresAt,
		&refreshRevokedAt,
		&record.SessionExpiresAt,
		&sessionRevokedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrInvalidRefreshToken
	}
	if err != nil {
		return nil, fmt.Errorf("select refresh token: %w", err)
	}

	record.RefreshRevokedAt = nullTime(refreshRevokedAt)
	record.SessionRevokedAt = nullTime(sessionRevokedAt)

	return record, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tx *sql.Tx, tokenID uuid.UUID, now time.Time, reason string) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = $2, revoked_reason = $3
		WHERE id = $1 AND revoked_at IS NULL
	`

	if _, err := tx.ExecContext(ctx, query, tokenID, now, reason); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) LinkRefreshTokenReplacement(ctx context.Context, tx *sql.Tx, oldTokenID, newTokenID uuid.UUID) error {
	const query = `
		UPDATE refresh_tokens
		SET replaced_by_token_id = $2
		WHERE id = $1
	`

	if _, err := tx.ExecContext(ctx, query, oldTokenID, newTokenID); err != nil {
		return fmt.Errorf("link refresh token replacement: %w", err)
	}

	return nil
}

func (r *AuthRepository) RevokeSessionAndTokens(ctx context.Context, tx *sql.Tx, sessionID uuid.UUID, now time.Time, reason string) error {
	const revokeSession = `
		UPDATE sessions
		SET revoked_at = $2, revoked_reason = $3
		WHERE id = $1 AND revoked_at IS NULL
	`

	if _, err := tx.ExecContext(ctx, revokeSession, sessionID, now, reason); err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	const revokeTokens = `
		UPDATE refresh_tokens
		SET revoked_at = $2, revoked_reason = $3
		WHERE session_id = $1 AND revoked_at IS NULL
	`

	if _, err := tx.ExecContext(ctx, revokeTokens, sessionID, now, reason); err != nil {
		return fmt.Errorf("revoke refresh tokens by session: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505"
}

func nullTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	at := value.Time
	return &at
}
