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

	"acc-dp/producer/internal/batch"
	"acc-dp/producer/internal/broker/redpanda"
	"acc-dp/producer/internal/buffer/badger"
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
		zap.Duration("physics_interval", cfg.PhysicsInterval),
		zap.Duration("graphics_interval", cfg.GraphicsInterval),
		zap.Duration("static_interval", cfg.StaticInterval),
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

	eventSchemas, err := avro.BuildEventSchemas()
	if err != nil {
		return fmt.Errorf("build event schemas: %w", err)
	}

	publisher, err := redpanda.NewPublisher(redpanda.Config{
		Brokers:           cfg.Brokers,
		SchemaRegistryURL: cfg.SchemaRegistryURL,
		Topics: redpanda.Topics{
			Physics:  cfg.TopicPhysics,
			Graphics: cfg.TopicGraphics,
			Static:   cfg.TopicStatic,
		},
		Schemas: redpanda.Schemas{
			Physics:  eventSchemas.Physics,
			Graphics: eventSchemas.Graphics,
			Static:   eventSchemas.Static,
		},
	})
	if err != nil {
		return fmt.Errorf("create redpanda publisher: %w", err)
	}
	defer closePublisher(logger, publisher)

	localBuffer, err := badger.New(cfg.BadgerPath)
	if err != nil {
		return fmt.Errorf("create badger local buffer: %w", err)
	}
	defer func() {
		if err := localBuffer.Close(); err != nil {
			logger.Warn("close local buffer", zap.Error(err))
		} else {
			logger.Info("local buffer closed")
		}
	}()

	adapter := batch.NewPublisherAdapter(publisher)
	batcher, err := batch.New(adapter, batch.Config{
		FlushInterval: cfg.FlushInterval,
		Logger:        logger,
		LocalBuffer:   localBuffer,
	})
	if err != nil {
		return fmt.Errorf("create batcher: %w", err)
	}

	go batcher.Start(ctx)
	defer batcher.Stop()

	identity := normalizer.Identity{
		UsuarioID: userID,
		Username:  username,
	}

	err = runCaptureLoop(
		ctx,
		cfg,
		logger,
		reader,
		normalizerService,
		serializer,
		batcher,
		identity,
		schemaVersion,
	)
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

func closePublisher(logger *zap.Logger, publisher *redpanda.Publisher) {
	if publisher == nil {
		return
	}

	publisher.Close()
	logger.Info("redpanda publisher closed")
}

func runCaptureLoop(
	ctx context.Context,
	cfg config.Config,
	logger *zap.Logger,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	batcher *batch.Batcher,
	identity normalizer.Identity,
	schemaVersion int32,
) error {
	physicsTicker := time.NewTicker(cfg.PhysicsInterval)
	defer physicsTicker.Stop()

	graphicsTicker := time.NewTicker(cfg.GraphicsInterval)
	defer graphicsTicker.Stop()

	staticTicker := time.NewTicker(cfg.StaticInterval)
	defer staticTicker.Stop()

	capturePhysicsFn := func(runCtx context.Context) (int, error) {
		return capturePhysics(runCtx, reader, normalizerService, serializer, batcher, cfg.TopicPhysics, identity, schemaVersion)
	}

	captureGraphicsFn := func(runCtx context.Context) (int, error) {
		return captureGraphics(runCtx, reader, normalizerService, serializer, batcher, cfg.TopicGraphics, identity, schemaVersion)
	}

	captureStaticFn := func(runCtx context.Context) (int, error) {
		return captureStatic(runCtx, reader, normalizerService, serializer, batcher, cfg.TopicStatic, identity, schemaVersion)
	}

	logger.Info("capture loop started")

	if err := captureAndLog(ctx, logger, "physics", capturePhysicsFn); err != nil {
		return err
	}

	if err := captureAndLog(ctx, logger, "graphics", captureGraphicsFn); err != nil {
		return err
	}

	if err := captureAndLog(ctx, logger, "static", captureStaticFn); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received")
			return ctx.Err()
		case <-physicsTicker.C:
			if err := captureAndLog(ctx, logger, "physics", capturePhysicsFn); err != nil {
				return err
			}
		case <-graphicsTicker.C:
			if err := captureAndLog(ctx, logger, "graphics", captureGraphicsFn); err != nil {
				return err
			}
		case <-staticTicker.C:
			if err := captureAndLog(ctx, logger, "static", captureStaticFn); err != nil {
				return err
			}
		}
	}
}

func captureAndLog(
	ctx context.Context,
	logger *zap.Logger,
	source string,
	captureFn func(context.Context) (int, error),
) error {
	bytesRead, err := captureFn(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return err
		}

		logger.Warn("capture failed",
			zap.String("source", source),
			zap.Error(err),
		)
		return nil
	}

	logger.Info("capture completed",
		zap.String("source", source),
		zap.Int("bytes", bytesRead),
	)

	return nil
}

func capturePhysics(
	ctx context.Context,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	batcher *batch.Batcher,
	topic string,
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

	if err := batcher.Enqueue(batch.Message{
		Topic:   topic,
		Key:     event.EventID,
		Payload: payload,
	}); err != nil {
		return 0, fmt.Errorf("enqueue physics: %w", err)
	}

	return len(payload), nil
}

func captureGraphics(
	ctx context.Context,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	batcher *batch.Batcher,
	topic string,
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

	if err := batcher.Enqueue(batch.Message{
		Topic:   topic,
		Key:     event.EventID,
		Payload: payload,
	}); err != nil {
		return 0, fmt.Errorf("enqueue graphics: %w", err)
	}

	return len(payload), nil
}

func captureStatic(
	ctx context.Context,
	reader acc_shm.Reader,
	normalizerService *normalizer.Service,
	serializer *avro.Serializer,
	batcher *batch.Batcher,
	topic string,
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

	if err := batcher.Enqueue(batch.Message{
		Topic:   topic,
		Key:     event.EventID,
		Payload: payload,
	}); err != nil {
		return 0, fmt.Errorf("enqueue static: %w", err)
	}

	return len(payload), nil
}

