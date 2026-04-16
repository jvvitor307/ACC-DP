package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"acc-dp/backend/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)
	CreateUser(ctx context.Context, tx *sql.Tx, user *domain.User) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpsertMachine(ctx context.Context, tx *sql.Tx, machineUID, deviceName string, now time.Time) (*domain.Machine, error)
	RevokeActiveSessionsByUser(ctx context.Context, tx *sql.Tx, userID uuid.UUID, now time.Time, reason string) error
	RevokeActiveSessionsByMachine(ctx context.Context, tx *sql.Tx, machineID uuid.UUID, now time.Time, reason string) error
	CreateSession(ctx context.Context, tx *sql.Tx, session *domain.Session) error
	CreateRefreshToken(ctx context.Context, tx *sql.Tx, token *domain.RefreshToken) error
	GetRefreshTokenRecordForUpdate(ctx context.Context, tx *sql.Tx, tokenHash string) (*domain.RefreshTokenRecord, error)
	RevokeRefreshToken(ctx context.Context, tx *sql.Tx, tokenID uuid.UUID, now time.Time, reason string) error
	LinkRefreshTokenReplacement(ctx context.Context, tx *sql.Tx, oldTokenID, newTokenID uuid.UUID) error
	RevokeSessionAndTokens(ctx context.Context, tx *sql.Tx, sessionID uuid.UUID, now time.Time, reason string) error
}

type Service struct {
	repo            Repository
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	sessionTTL      time.Duration
	now             func() time.Time
}

type RegisterInput struct {
	Email       string
	DisplayName string
	Password    string
}

type RegisterResult struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type LoginInput struct {
	Email     string
	Password  string
	MachineID string
	DeviceName string
}

type RefreshInput struct {
	RefreshToken string
	MachineID    string
}

type LogoutInput struct {
	RefreshToken string
	MachineID    string
}

type Tokens struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type LoginResult struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
	MachineID string `json:"machine_id"`
	Tokens
}

type RefreshResult struct {
	SessionID string `json:"session_id"`
	MachineID string `json:"machine_id"`
	Tokens
}

func NewService(repo Repository, jwtSecret string, accessTokenTTL, refreshTokenTTL, sessionTTL time.Duration) (*Service, error) {
	secret := strings.TrimSpace(jwtSecret)
	if repo == nil {
		return nil, fmt.Errorf("create auth service: repository is nil")
	}
	if secret == "" {
		return nil, fmt.Errorf("create auth service: jwt secret is empty")
	}
	if accessTokenTTL <= 0 || refreshTokenTTL <= 0 || sessionTTL <= 0 {
		return nil, fmt.Errorf("create auth service: ttl values must be greater than zero")
	}

	return &Service{
		repo:            repo,
		jwtSecret:       []byte(secret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		sessionTTL:      sessionTTL,
		now:             time.Now,
	}, nil
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterResult, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}
	displayName, err := normalizeDisplayName(input.DisplayName)
	if err != nil {
		return nil, err
	}
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := s.now().UTC()
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		DisplayName:  displayName,
		PasswordHash: string(passwordHash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackQuiet(tx)

	if err := s.repo.CreateUser(ctx, tx, user); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit register transaction: %w", err)
	}

	return &RegisterResult{
		UserID:      user.ID.String(),
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Password) == "" {
		return nil, fmt.Errorf("login: %w: password is required", domain.ErrInvalidInput)
	}
	machineID, err := normalizeMachineID(input.MachineID)
	if err != nil {
		return nil, err
	}
	deviceName := strings.TrimSpace(input.DeviceName)

	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	now := s.now().UTC()
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackQuiet(tx)

	machine, err := s.repo.UpsertMachine(ctx, tx, machineID, deviceName, now)
	if err != nil {
		return nil, err
	}

	if err := s.repo.RevokeActiveSessionsByUser(ctx, tx, user.ID, now, "user logged in from another machine"); err != nil {
		return nil, err
	}

	if err := s.repo.RevokeActiveSessionsByMachine(ctx, tx, machine.ID, now, "machine assigned to another user"); err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:        uuid.New(),
		UserID:    user.ID,
		MachineID: machine.ID,
		CreatedAt: now,
		ExpiresAt: now.Add(s.sessionTTL),
	}

	if err := s.repo.CreateSession(ctx, tx, session); err != nil {
		return nil, err
	}

	refreshPlain, refreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshToken := &domain.RefreshToken{
		ID:        uuid.New(),
		SessionID: session.ID,
		UserID:    user.ID,
		TokenHash: refreshHash,
		CreatedAt: now,
		ExpiresAt: now.Add(s.refreshTokenTTL),
	}

	if err := s.repo.CreateRefreshToken(ctx, tx, refreshToken); err != nil {
		return nil, err
	}

	accessToken, expiresIn, err := s.signAccessToken(user.ID, session.ID, machineID, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit login transaction: %w", err)
	}

	return &LoginResult{
		UserID:    user.ID.String(),
		SessionID: session.ID.String(),
		MachineID: machineID,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshPlain,
			TokenType:    "Bearer",
			ExpiresIn:    expiresIn,
		},
	}, nil
}

func (s *Service) Refresh(ctx context.Context, input RefreshInput) (*RefreshResult, error) {
	machineID, err := normalizeMachineID(input.MachineID)
	if err != nil {
		return nil, err
	}
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh: %w: refresh_token is required", domain.ErrInvalidInput)
	}

	refreshHash := hashToken(refreshToken)
	now := s.now().UTC()

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer rollbackQuiet(tx)

	record, err := s.repo.GetRefreshTokenRecordForUpdate(ctx, tx, refreshHash)
	if err != nil {
		return nil, err
	}

	if record.RefreshRevokedAt != nil || now.After(record.RefreshExpiresAt) {
		return nil, domain.ErrInvalidRefreshToken
	}

	if record.SessionRevokedAt != nil || now.After(record.SessionExpiresAt) {
		return nil, domain.ErrInvalidRefreshToken
	}

	if record.MachineUID != machineID {
		return nil, domain.ErrMachineMismatch
	}

	if err := s.repo.RevokeRefreshToken(ctx, tx, record.RefreshTokenID, now, "refresh token rotation"); err != nil {
		return nil, err
	}

	newRefreshPlain, newRefreshHash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	newRefresh := &domain.RefreshToken{
		ID:        uuid.New(),
		SessionID: record.SessionID,
		UserID:    record.UserID,
		TokenHash: newRefreshHash,
		CreatedAt: now,
		ExpiresAt: now.Add(s.refreshTokenTTL),
	}

	if err := s.repo.CreateRefreshToken(ctx, tx, newRefresh); err != nil {
		return nil, err
	}

	if err := s.repo.LinkRefreshTokenReplacement(ctx, tx, record.RefreshTokenID, newRefresh.ID); err != nil {
		return nil, err
	}

	accessToken, expiresIn, err := s.signAccessToken(record.UserID, record.SessionID, record.MachineUID, now)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit refresh transaction: %w", err)
	}

	return &RefreshResult{
		SessionID: record.SessionID.String(),
		MachineID: record.MachineUID,
		Tokens: Tokens{
			AccessToken:  accessToken,
			RefreshToken: newRefreshPlain,
			TokenType:    "Bearer",
			ExpiresIn:    expiresIn,
		},
	}, nil
}

func (s *Service) Logout(ctx context.Context, input LogoutInput) error {
	machineID, err := normalizeMachineID(input.MachineID)
	if err != nil {
		return err
	}
	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return fmt.Errorf("logout: %w: refresh_token is required", domain.ErrInvalidInput)
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer rollbackQuiet(tx)

	record, err := s.repo.GetRefreshTokenRecordForUpdate(ctx, tx, hashToken(refreshToken))
	if err != nil {
		return err
	}

	if record.MachineUID != machineID {
		return domain.ErrMachineMismatch
	}

	now := s.now().UTC()
	if err := s.repo.RevokeSessionAndTokens(ctx, tx, record.SessionID, now, "logout"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit logout transaction: %w", err)
	}

	return nil
}

func (s *Service) signAccessToken(userID, sessionID uuid.UUID, machineID string, now time.Time) (string, int64, error) {
	expiresAt := now.Add(s.accessTokenTTL)
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"sid": sessionID.String(),
		"mid": machineID,
		"iat": now.Unix(),
		"exp": expiresAt.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, fmt.Errorf("sign access token: %w", err)
	}

	return signed, int64(s.accessTokenTTL.Seconds()), nil
}

func generateRefreshToken() (string, string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}

	plain := base64.RawURLEncoding.EncodeToString(buffer)
	return plain, hashToken(plain), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || !strings.Contains(normalized, "@") {
		return "", fmt.Errorf("email: %w", domain.ErrInvalidInput)
	}
	return normalized, nil
}

func normalizeDisplayName(name string) (string, error) {
	normalized := strings.TrimSpace(name)
	if normalized == "" {
		return "", fmt.Errorf("display_name: %w", domain.ErrInvalidInput)
	}
	if len(normalized) > 120 {
		return "", fmt.Errorf("display_name: %w", domain.ErrInvalidInput)
	}
	return normalized, nil
}

func normalizeMachineID(machineID string) (string, error) {
	normalized := strings.TrimSpace(machineID)
	if normalized == "" {
		return "", fmt.Errorf("machine_id: %w", domain.ErrInvalidInput)
	}
	if len(normalized) > 255 {
		return "", fmt.Errorf("machine_id: %w", domain.ErrInvalidInput)
	}
	return normalized, nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password: %w", domain.ErrInvalidInput)
	}
	return nil
}

func rollbackQuiet(tx *sql.Tx) {
	if tx == nil {
		return
	}
	_ = tx.Rollback()
}

func IsAuthError(err error) bool {
	return errors.Is(err, domain.ErrInvalidInput) ||
		errors.Is(err, domain.ErrUserAlreadyExists) ||
		errors.Is(err, domain.ErrInvalidCredentials) ||
		errors.Is(err, domain.ErrInvalidRefreshToken) ||
		errors.Is(err, domain.ErrMachineMismatch)
}
