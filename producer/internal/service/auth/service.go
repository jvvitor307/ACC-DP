package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	domainuser "acc-dp/producer/internal/domain/user"
	"acc-dp/producer/internal/repository/postgres"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidInput       = errors.New("invalid input")
	ErrSessionExpired     = errors.New("session expired")
)

type LoginResult struct {
	User    *domainuser.User
	Session *domainuser.Session
	Token   string
}

type Service struct {
	repo       *postgres.Repository
	now        func() time.Time
	sessionTTL time.Duration
}

func New(repo *postgres.Repository, sessionTTL time.Duration) *Service {
	return &Service{
		repo:       repo,
		now:        time.Now,
		sessionTTL: sessionTTL,
	}
}

func (s *Service) Register(ctx context.Context, username, name, password, role string) (*domainuser.User, error) {
	username = strings.TrimSpace(username)
	name = strings.TrimSpace(name)
	password = strings.TrimSpace(password)
	role = strings.TrimSpace(role)

	if username == "" || name == "" || password == "" {
		return nil, ErrInvalidInput
	}

	if role == "" {
		role = "operator"
	}

	hash, err := hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	now := s.now().UTC()
	created, err := s.repo.CreateUser(ctx, &domainuser.User{
		ID:           uuid.NewString(),
		Username:     username,
		Name:         name,
		PasswordHash: hash,
		Role:         role,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	return created, nil
}

func (s *Service) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	username = strings.TrimSpace(username)
	password = strings.TrimSpace(password)
	if username == "" || password == "" {
		return nil, ErrInvalidInput
	}

	account, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, tokenHash, err := newSessionToken()
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	now := s.now().UTC()
	session, err := s.repo.CreateSession(ctx, &domainuser.Session{
		ID:        uuid.NewString(),
		UserID:    account.ID,
		TokenHash: tokenHash,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}

	return &LoginResult{
		User:    account,
		Session: session,
		Token:   token,
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*domainuser.User, *domainuser.Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, nil, ErrInvalidCredentials
	}

	tokenHash := hashToken(token)
	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("authenticate: %w", err)
	}

	now := s.now().UTC()
	if now.After(session.ExpiresAt) {
		_ = s.repo.DeleteSession(ctx, session.ID)
		return nil, nil, ErrSessionExpired
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, nil, ErrInvalidCredentials
		}
		return nil, nil, fmt.Errorf("authenticate: %w", err)
	}

	return user, session, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil
	}

	if err := s.repo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("logout: %w", err)
	}

	return nil
}

func (s *Service) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	deleted, err := s.repo.DeleteExpiredSessions(ctx, s.now().UTC())
	if err != nil {
		return 0, fmt.Errorf("cleanup expired sessions: %w", err)
	}

	return deleted, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}

	return string(hash), nil
}

func newSessionToken() (token string, tokenHash string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", fmt.Errorf("read random bytes: %w", err)
	}

	token = base64.RawURLEncoding.EncodeToString(raw)
	tokenHash = hashToken(token)

	return token, tokenHash, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
