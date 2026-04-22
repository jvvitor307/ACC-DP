package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	defaultLogLevel = "info"
)

var requiredEnvKeys = []string{
	"REDPANDA_BROKERS",
	"TOPIC_PHYSICS",
	"TOPIC_GRAPHICS",
	"TOPIC_STATIC",
	"SCHEMA_REGISTRY_URL",
	"PRODUCER_FLUSH_INTERVAL",
	"BADGER_PATH",
	"USER_STORAGE_PATH",
}

type Config struct {
	Brokers           []string
	TopicPhysics      string
	TopicGraphics     string
	TopicStatic       string
	SchemaRegistryURL string
	PhysicsInterval   time.Duration
	GraphicsInterval  time.Duration
	StaticInterval    time.Duration
	BadgerPath        string
	UserStoragePath   string
	LogLevel          string
}

func Load() (Config, error) {
	envValues, err := readRequiredEnvs(requiredEnvKeys)
	if err != nil {
		return Config{}, err
	}

	brokers, err := parseBrokers(envValues["REDPANDA_BROKERS"])
	if err != nil {
		return Config{}, err
	}

	physicsInterval, err := parseConfiguredInterval("PRODUCER_PHYSICS_INTERVAL", envValues["PRODUCER_FLUSH_INTERVAL"])
	if err != nil {
		return Config{}, err
	}

	graphicsInterval, err := parseConfiguredInterval("PRODUCER_GRAPHICS_INTERVAL", envValues["PRODUCER_FLUSH_INTERVAL"])
	if err != nil {
		return Config{}, err
	}

	staticInterval, err := parseConfiguredInterval("PRODUCER_STATIC_INTERVAL", envValues["PRODUCER_FLUSH_INTERVAL"])
	if err != nil {
		return Config{}, err
	}

	return Config{
		Brokers:           brokers,
		TopicPhysics:      envValues["TOPIC_PHYSICS"],
		TopicGraphics:     envValues["TOPIC_GRAPHICS"],
		TopicStatic:       envValues["TOPIC_STATIC"],
		SchemaRegistryURL: envValues["SCHEMA_REGISTRY_URL"],
		PhysicsInterval:   physicsInterval,
		GraphicsInterval:  graphicsInterval,
		StaticInterval:    staticInterval,
		BadgerPath:        envValues["BADGER_PATH"],
		UserStoragePath:   envValues["USER_STORAGE_PATH"],
		LogLevel:          getEnv("PRODUCER_LOG_LEVEL", defaultLogLevel),
	}, nil
}

func parseConfiguredInterval(key, fallbackRaw string) (time.Duration, error) {
	raw := getEnv(key, fallbackRaw)

	interval, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}

	if interval <= 0 {
		return 0, fmt.Errorf("%s must be greater than zero", key)
	}

	return interval, nil
}

func readRequiredEnvs(keys []string) (map[string]string, error) {
	values := make(map[string]string, len(keys))
	missing := make([]string, 0)

	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value == "" {
			missing = append(missing, key)
			continue
		}

		values[key] = value
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return values, nil
}

func parseBrokers(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	brokers := make([]string, 0, len(parts))

	for _, part := range parts {
		broker := strings.TrimSpace(part)
		if broker == "" {
			continue
		}

		brokers = append(brokers, broker)
	}

	if len(brokers) == 0 {
		return nil, fmt.Errorf("REDPANDA_BROKERS must contain at least one broker")
	}

	return brokers, nil
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	return value
}
