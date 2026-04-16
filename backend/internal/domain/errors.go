package domain

import "errors"

var (
	ErrUserAlreadyExists   = errors.New("user already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrMachineMismatch     = errors.New("machine id does not match session")
	ErrInvalidInput        = errors.New("invalid input")
)
