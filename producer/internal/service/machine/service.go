package machine

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

	domainmachine "acc-dp/producer/internal/domain/machine"
	"acc-dp/producer/internal/repository/postgres"

	"github.com/google/uuid"
)

var (
	ErrInvalidMachineInput       = errors.New("invalid machine input")
	ErrMachineAlreadyRegistered  = errors.New("machine already registered")
	ErrMachineNotFound           = errors.New("machine not found")
	ErrInvalidMachineCredentials = errors.New("invalid machine credentials")
	ErrMachineInactive           = errors.New("machine inactive")
)

const statusActive = "active"

type Service struct {
	repo *postgres.Repository
	now  func() time.Time
}

func New(repo *postgres.Repository) *Service {
	return &Service{
		repo: repo,
		now:  time.Now,
	}
}

func (s *Service) Register(ctx context.Context, name, fingerprint string) (*domainmachine.RegisterResult, error) {
	name = strings.TrimSpace(name)
	fingerprintHash := hashFingerprint(strings.TrimSpace(fingerprint))

	if fingerprintHash != nil {
		_, err := s.repo.GetMachineByFingerprintHash(ctx, *fingerprintHash)
		if err == nil {
			return nil, ErrMachineAlreadyRegistered
		}
		if !errors.Is(err, postgres.ErrNotFound) {
			return nil, fmt.Errorf("register machine: %w", err)
		}
	}

	machineToken, tokenHash, err := newMachineToken()
	if err != nil {
		return nil, fmt.Errorf("register machine: %w", err)
	}

	now := s.now().UTC()
	created, err := s.repo.CreateMachine(ctx, &domainmachine.Machine{
		ID:              uuid.NewString(),
		Name:            name,
		FingerprintHash: fingerprintHash,
		TokenHash:       tokenHash,
		Status:          statusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
		LastSeenAt:      now,
	})
	if err != nil {
		if errors.Is(err, postgres.ErrConflict) {
			return nil, ErrMachineAlreadyRegistered
		}
		return nil, fmt.Errorf("register machine: %w", err)
	}

	return &domainmachine.RegisterResult{
		Machine:      created,
		MachineToken: machineToken,
	}, nil
}

func (s *Service) Authenticate(ctx context.Context, machineID string, machineToken string) (*domainmachine.Machine, error) {
	machineID = strings.TrimSpace(machineID)
	machineToken = strings.TrimSpace(machineToken)
	if machineID == "" || machineToken == "" {
		return nil, ErrInvalidMachineInput
	}

	tokenHash := hashToken(machineToken)
	machine, err := s.repo.GetMachineByIDAndTokenHash(ctx, machineID, tokenHash)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrInvalidMachineCredentials
		}
		return nil, fmt.Errorf("authenticate machine: %w", err)
	}

	if !strings.EqualFold(machine.Status, statusActive) {
		return nil, ErrMachineInactive
	}

	return machine, nil
}

func (s *Service) Touch(ctx context.Context, machineID string) (time.Time, error) {
	machineID = strings.TrimSpace(machineID)
	if machineID == "" {
		return time.Time{}, ErrInvalidMachineInput
	}

	now := s.now().UTC()
	if err := s.repo.UpdateMachineLastSeen(ctx, machineID, now); err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return time.Time{}, ErrMachineNotFound
		}
		return time.Time{}, fmt.Errorf("touch machine: %w", err)
	}

	return now, nil
}

func (s *Service) GetByID(ctx context.Context, machineID string) (*domainmachine.Machine, error) {
	machineID = strings.TrimSpace(machineID)
	if machineID == "" {
		return nil, ErrInvalidMachineInput
	}

	machine, err := s.repo.GetMachineByID(ctx, machineID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrMachineNotFound
		}
		return nil, fmt.Errorf("get machine by id: %w", err)
	}

	return machine, nil
}

func (s *Service) List(ctx context.Context) ([]domainmachine.Machine, error) {
	machines, err := s.repo.ListMachines(ctx)
	if err != nil {
		return nil, fmt.Errorf("list machines: %w", err)
	}

	return machines, nil
}

func hashFingerprint(fingerprint string) *string {
	if fingerprint == "" {
		return nil
	}

	sum := sha256.Sum256([]byte(fingerprint))
	hash := hex.EncodeToString(sum[:])

	return &hash
}

func newMachineToken() (token string, tokenHash string, err error) {
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
