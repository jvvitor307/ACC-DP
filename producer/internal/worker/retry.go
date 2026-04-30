package worker

import (
	"context"
	"time"

	"go.uber.org/zap"

	"acc-dp/producer/internal/batch"
	"acc-dp/producer/internal/buffer/badger"
	"acc-dp/producer/internal/metrics"
)

// RetryWorker is responsible for asynchronously draining the persistent local buffer
// and attempting to republish the messages to the broker.
type RetryWorker struct {
	publisher batch.Publisher
	buffer    *badger.Buffer
	logger    *zap.Logger
	metrics   *metrics.Collector
}

func NewRetryWorker(pub batch.Publisher, buf *badger.Buffer, logger *zap.Logger, m *metrics.Collector) *RetryWorker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &RetryWorker{
		publisher: pub,
		buffer:    buf,
		logger:    logger,
		metrics:   m,
	}
}

// Start begins the asynchronous polling loop.
// It will periodically review the backlogged content in BadgerDB and try publishing it.
func (w *RetryWorker) Start(ctx context.Context) {
	// A static 5-second backoff allows us to prevent I/O saturation and CPU blocking.
	backoff := 5 * time.Second
	ticker := time.NewTicker(backoff)
	defer ticker.Stop()

	w.logger.Info("retry worker started", zap.Duration("backoff", backoff))

	for {
		select {
			
		case <-ctx.Done():
			w.logger.Info("retry worker stopped")
			return
		case <-ticker.C:
			w.processIter(ctx)
		}
	}
}

func (w *RetryWorker) processIter(ctx context.Context) {
	if w.buffer == nil {
		w.logger.Warn("retry worker storage buffer is nil, skipping iteration")
		return
	}

	pending, err := w.buffer.ReadPending()
	if err != nil {
		w.logger.Error("failed to read pending messages from local buffer", zap.Error(err))
		return
	}

	if len(pending) == 0 {
		// Nothing to process, backlog is empty.
		return
	}

	w.logger.Debug("found pending messages to retry", zap.Int("total_pending", len(pending)))

	// Group messages by topic so we can batch publish them efficiently.
	grouped := make(map[string][]badger.BufferedMessage)
	for _, msg := range pending {
		grouped[msg.Topic] = append(grouped[msg.Topic], msg)
	}

	for topic, msgs := range grouped {
		var batchMsgs []batch.Message
		for _, m := range msgs {
			batchMsgs = append(batchMsgs, batch.Message{
				Topic:   m.Topic,
				Key:     m.Key,
				Payload: m.Payload,
			})
		}

		err := w.publisher.PublishBatch(ctx, topic, batchMsgs)
		if err != nil {
			w.logger.Warn("retry backlog publish failed, keeping them in storage",
				zap.String("topic", topic),
				zap.Int("count", len(batchMsgs)),
				zap.Error(err),
			)

			if w.metrics != nil {
				w.metrics.IncRetryErrors()
			}

			continue
		}

		successCount := 0
		for _, m := range msgs {
			if ackErr := w.buffer.Ack(m.ID); ackErr != nil {
				w.logger.Error("failed to ack/remove message after successful retry",
					zap.String("id", m.ID),
					zap.Error(ackErr),
				)
			} else {
				successCount++
			}
		}

		if w.metrics != nil {
			w.metrics.AddRetrySuccess(uint64(successCount))
			w.metrics.AddDequeued(uint64(successCount))
		}

		w.logger.Info("successfully retried and removed backlog messages",
			zap.String("topic", topic),
			zap.Int("successful_acks", successCount),
		)
	}
}
