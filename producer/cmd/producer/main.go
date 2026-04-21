package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.uber.org/zap"

	"acc-dp/producer/internal/config"
	"acc-dp/producer/internal/service/avro"
	"acc-dp/producer/internal/service/normalizer"
	"acc-dp/producer/internal/source/acc_shm"
)

const (
	schemaVersion = int32(1)
	userID        = "local-user"
	username      = "local"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "producer failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger, err := newLogger(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("create logger: %w", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("starting producer",
		zap.Duration("flush_interval", cfg.FlushInterval),
		zap.String("log_level", cfg.LogLevel),
	)

	reader, err := acc_shm.NewReader(ctx)
	if err != nil {
		return fmt.Errorf("create shared memory reader: %w", err)
	}
	defer closeReader(logger, reader)

	normalizerService := normalizer.New()
	serializer, err := avro.NewSerializer()
	if err != nil {
		return fmt.Errorf("create serializer: %w", err)
	}

	identity := normalizer.Identity{
		UsuarioID: userID,
		Username:  username,
	}

	err = runCaptureLoop(ctx, cfg.FlushInterval, logger, reader, normalizerService, serializer, identity, schemaVersion)
	if err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("run capture loop: %w", err)
	}

	logger.Info("producer stopped cleanly")
	return nil
}

func newLogger(levelText string) (*zap.Logger, error) {
	loggerConfig := zap.NewProductionConfig()
	if err := loggerConfig.Level.UnmarshalText([]byte(strings.TrimSpace(levelText))); err != nil {
		return nil, fmt.Errorf("parse PRODUCER_LOG_LEVEL: %w", err)
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("build logger: %w", err)
	}

	return logger, nil
}

func closeReader(logger *zap.Logger, reader acc_shm.Reader) {
	if err := reader.Close(); err != nil {
		logger.Warn("close shared memory reader", zap.Error(err))
		return
	}

	logger.Info("shared memory reader closed")
}

func runCaptureLoop(
	ctx context.Context,
	flushInterval time.Duration,
	logger *zap.Logger,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) error {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	logger.Info("capture loop started")

	for {
		if err := captureCycle(ctx, logger, reader, normalizerService, serializer, identity, schemaVersion); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return fmt.Errorf("capture cycle: %w", err)
		}

		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received")
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func captureCycle(
	ctx context.Context,
	logger *zap.Logger,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) error {
	cycleResult := struct {
		physicsBytes  int
		graphicsBytes int
		staticBytes   int
		failures      int
	}{}

	physicsBytes, err := capturePhysics(ctx, reader, normalizerService, serializer, identity, schemaVersion)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		cycleResult.failures++
		logger.Warn("capture physics failed", zap.Error(err))
	} else {
		cycleResult.physicsBytes = physicsBytes
	}

	graphicsBytes, err := captureGraphics(ctx, reader, normalizerService, serializer, identity, schemaVersion)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		cycleResult.failures++
		logger.Warn("capture graphics failed", zap.Error(err))
	} else {
		cycleResult.graphicsBytes = graphicsBytes
	}

	staticBytes, err := captureStatic(ctx, reader, normalizerService, serializer, identity, schemaVersion)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}
		cycleResult.failures++
		logger.Warn("capture static failed", zap.Error(err))
	} else {
		cycleResult.staticBytes = staticBytes
	}

	logger.Info("capture cycle completed",
		zap.Int("physics_bytes", cycleResult.physicsBytes),
		zap.Int("graphics_bytes", cycleResult.graphicsBytes),
		zap.Int("static_bytes", cycleResult.staticBytes),
		zap.Int("failures", cycleResult.failures),
	)

	return nil
}

func capturePhysics(
	ctx context.Context,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) (int, error) {
	page, err := reader.ReadPhysics(ctx)
	if err != nil {
		return 0, fmt.Errorf("read physics: %w", err)
	}

	event, err := normalizerService.NormalizePhysics(page, identity, schemaVersion)
	if err != nil {
		return 0, fmt.Errorf("normalize physics: %w", err)
	}

	payload, err := serializer.SerializePhysics(event)
	if err != nil {
		return 0, fmt.Errorf("serialize physics: %w", err)
	}

	return len(payload), nil
}

func captureGraphics(
	ctx context.Context,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) (int, error) {
	page, err := reader.ReadGraphics(ctx)
	if err != nil {
		return 0, fmt.Errorf("read graphics: %w", err)
	}

	event, err := normalizerService.NormalizeGraphics(page, identity, schemaVersion)
	if err != nil {
		return 0, fmt.Errorf("normalize graphics: %w", err)
	}

	payload, err := serializer.SerializeGraphics(event)
	if err != nil {
		return 0, fmt.Errorf("serialize graphics: %w", err)
	}

	return len(payload), nil
}

func captureStatic(
	ctx context.Context,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	identity normalizer.Identity,
	schemaVersion int32,
) (int, error) {
	page, err := reader.ReadStatic(ctx)
	if err != nil {
		return 0, fmt.Errorf("read static: %w", err)
	}

	event, err := normalizerService.NormalizeStatic(page, identity, schemaVersion)
	if err != nil {
		return 0, fmt.Errorf("normalize static: %w", err)
	}

	payload, err := serializer.SerializeStatic(event)
	if err != nil {
		return 0, fmt.Errorf("serialize static: %w", err)
	}

	return len(payload), nil
}
