package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	domainuser "acc-dp/producer/internal/domain/user"
)

var ErrNotFound = errors.New("not found")

type Repository struct {
	db *sqlx.DB
}

func New(databaseURL string) (*Repository, error) {
	db, err := sqlx.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	return &Repository{db: db}, nil
}

func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}
	return r.db.Close()
}

func (r *Repository) Ping(ctx context.Context) error {
	if err := r.db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}

func (r *Repository) RunMigrations(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'operator',
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS user_sessions (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash TEXT NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions (expires_at);`,
		`CREATE TABLE IF NOT EXISTS active_users (
			machine_id TEXT PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			updated_at TIMESTAMPTZ NOT NULL
		);`,
	}

	for _, stmt := range statements {
		if _, err := r.db.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("run migration statement: %w", err)
		}
	}

	return nil
}

func (r *Repository) CreateUser(ctx context.Context, user *domainuser.User) (*domainuser.User, error) {
	if user == nil {
		return nil, fmt.Errorf("create user: user is nil")
	}

	const query = `
		INSERT INTO users (id, username, name, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, username, name, password_hash, role, created_at, updated_at;
	`

	created := &domainuser.User{}
	err := r.db.GetContext(
		ctx,
		created,
		query,
		user.ID,
		user.Username,
		user.Name,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, fmt.Errorf("create user: username already exists")
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (*domainuser.User, error) {
	const query = `
		SELECT id, username, name, password_hash, role, created_at, updated_at
		FROM users
		WHERE username = $1;
	`

	user := &domainuser.User{}
	if err := r.db.GetContext(ctx, user, query, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (*domainuser.User, error) {
	const query = `
		SELECT id, username, name, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1;
	`

	user := &domainuser.User{}
	if err := r.db.GetContext(ctx, user, query, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (r *Repository) ListUsers(ctx context.Context) ([]domainuser.User, error) {
	const query = `
		SELECT id, username, name, password_hash, role, created_at, updated_at
		FROM users
		ORDER BY created_at DESC;
	`

	users := make([]domainuser.User, 0)
	if err := r.db.SelectContext(ctx, &users, query); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (r *Repository) SetActiveUserForMachine(ctx context.Context, machineID string, userID string, updatedAt time.Time) error {
	const query = `
		INSERT INTO active_users (machine_id, user_id, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (machine_id)
		DO UPDATE SET user_id = EXCLUDED.user_id, updated_at = EXCLUDED.updated_at;
	`

	if _, err := r.db.ExecContext(ctx, query, machineID, userID, updatedAt); err != nil {
		return fmt.Errorf("set active user for machine: %w", err)
	}

	return nil
}

func (r *Repository) GetActiveUserIDForMachine(ctx context.Context, machineID string) (string, error) {
	const query = `
		SELECT user_id
		FROM active_users
		WHERE machine_id = $1;
	`

	var userID string
	if err := r.db.GetContext(ctx, &userID, query, machineID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("get active user id for machine: %w", err)
	}

	return userID, nil
}

func (r *Repository) ListActiveUsers(ctx context.Context) ([]domainuser.ActiveUserAssignment, error) {
	const query = `
		SELECT machine_id, user_id, updated_at
		FROM active_users
		ORDER BY updated_at DESC;
	`

	assignments := make([]domainuser.ActiveUserAssignment, 0)
	if err := r.db.SelectContext(ctx, &assignments, query); err != nil {
		return nil, fmt.Errorf("list active users: %w", err)
	}

	return assignments, nil
}

func (r *Repository) CreateSession(ctx context.Context, session *domainuser.Session) (*domainuser.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("create session: session is nil")
	}

	const query = `
		INSERT INTO user_sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, token_hash, expires_at, created_at;
	`

	created := &domainuser.Session{}
	err := r.db.GetContext(
		ctx,
		created,
		query,
		session.ID,
		session.UserID,
		session.TokenHash,
		session.ExpiresAt,
		session.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return created, nil
}

func (r *Repository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*domainuser.Session, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM user_sessions
		WHERE token_hash = $1;
	`

	session := &domainuser.Session{}
	if err := r.db.GetContext(ctx, session, query, tokenHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get session by token hash: %w", err)
	}

	return session, nil
}

func (r *Repository) DeleteSession(ctx context.Context, sessionID string) error {
	const query = `
		DELETE FROM user_sessions
		WHERE id = $1;
	`

	if _, err := r.db.ExecContext(ctx, query, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func (r *Repository) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	const query = `
		DELETE FROM user_sessions
		WHERE expires_at <= $1;
	`

	result, err := r.db.ExecContext(ctx, query, now)
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete expired sessions rows affected: %w", err)
	}

	return rowsAffected, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == "23505"
}
