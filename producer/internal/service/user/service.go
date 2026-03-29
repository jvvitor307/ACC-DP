package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"acc-dp/producer/internal/domain/user"
	"acc-dp/producer/internal/repository/postgres"
)

var (
	ErrInvalidUserInput = errors.New("invalid user input")
	ErrUserNotFound     = errors.New("user not found")
	ErrActiveUserNotSet = errors.New("active user not set")
)

const DefaultMachineID = "default"

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

func (s *Service) ListUsers(ctx context.Context) ([]user.User, error) {
	users, err := s.repo.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (s *Service) SetActiveUserForMachine(ctx context.Context, machineID string, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserInput
	}

	machineID = normalizeMachineID(machineID)

	_, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("set active user for machine: %w", err)
	}

	if err := s.repo.SetActiveUserForMachine(ctx, machineID, userID, s.now().UTC()); err != nil {
		return fmt.Errorf("set active user for machine: %w", err)
	}

	return nil
}

func (s *Service) GetActiveUserForMachine(ctx context.Context, machineID string) (*user.User, error) {
	machineID = normalizeMachineID(machineID)

	userID, err := s.repo.GetActiveUserIDForMachine(ctx, machineID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrActiveUserNotSet
		}
		return nil, fmt.Errorf("get active user id for machine: %w", err)
	}

	activeUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get active user for machine: %w", err)
	}

	return activeUser, nil
}

func (s *Service) ListActiveUsers(ctx context.Context) ([]user.ActiveUserAssignment, error) {
	assignments, err := s.repo.ListActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active users: %w", err)
	}

	return assignments, nil
}

func (s *Service) SetActiveUser(ctx context.Context, userID string) error {
	return s.SetActiveUserForMachine(ctx, DefaultMachineID, userID)
}

func (s *Service) GetActiveUser(ctx context.Context) (*user.User, error) {
	return s.GetActiveUserForMachine(ctx, DefaultMachineID)
}

func normalizeMachineID(machineID string) string {
	machineID = strings.TrimSpace(machineID)
	if machineID == "" {
		return DefaultMachineID
	}

	return machineID
}
