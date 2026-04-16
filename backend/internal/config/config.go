package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	PostgresHost    string
	PostgresPort    int
	PostgresUser    string
	PostgresPassword string
	PostgresDB      string
	PostgresSSLMode string
	JWTSecret       string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	SessionTTL      time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:         getEnv("BACKEND_HTTP_ADDR", ":8088"),
		PostgresHost:     getEnv("POSTGRES_HOST", "localhost"),
		PostgresPort:     getEnvInt("POSTGRES_PORT", 15432),
		PostgresUser:     getEnv("POSTGRES_USER", "acc_user"),
		PostgresPassword: getEnv("POSTGRES_PASSWORD", "acc_password"),
		PostgresDB:       getEnv("POSTGRES_DB", "acc_auth"),
		PostgresSSLMode:  getEnv("POSTGRES_SSLMODE", "disable"),
		JWTSecret:        strings.TrimSpace(os.Getenv("BACKEND_JWT_SECRET")),
	}

	accessTokenTTL, err := time.ParseDuration(getEnv("BACKEND_ACCESS_TOKEN_TTL", "15m"))
	if err != nil {
		return Config{}, fmt.Errorf("parse BACKEND_ACCESS_TOKEN_TTL: %w", err)
	}
	cfg.AccessTokenTTL = accessTokenTTL

	refreshTokenTTL, err := time.ParseDuration(getEnv("BACKEND_REFRESH_TOKEN_TTL", "168h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse BACKEND_REFRESH_TOKEN_TTL: %w", err)
	}
	cfg.RefreshTokenTTL = refreshTokenTTL

	sessionTTL, err := time.ParseDuration(getEnv("BACKEND_SESSION_TTL", "720h"))
	if err != nil {
		return Config{}, fmt.Errorf("parse BACKEND_SESSION_TTL: %w", err)
	}
	cfg.SessionTTL = sessionTTL

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("BACKEND_JWT_SECRET is required")
	}

	if cfg.AccessTokenTTL <= 0 {
		return Config{}, fmt.Errorf("BACKEND_ACCESS_TOKEN_TTL must be greater than zero")
	}

	if cfg.RefreshTokenTTL <= 0 {
		return Config{}, fmt.Errorf("BACKEND_REFRESH_TOKEN_TTL must be greater than zero")
	}

	if cfg.SessionTTL <= 0 {
		return Config{}, fmt.Errorf("BACKEND_SESSION_TTL must be greater than zero")
	}

	return cfg, nil
}

func (c Config) PostgresDSN() string {
	username := url.QueryEscape(c.PostgresUser)
	password := url.QueryEscape(c.PostgresPassword)
	database := url.QueryEscape(c.PostgresDB)
	sslmode := url.QueryEscape(c.PostgresSSLMode)

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", username, password, c.PostgresHost, c.PostgresPort, database, sslmode)
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
