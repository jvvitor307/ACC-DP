package batch

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"acc-dp/producer/internal/metrics"
)

// Message represents a single message to be batched and published.
type Message struct {
	Topic   string
	Key     string
	Payload []byte
}

// Publisher defines the interface for publishing a batch of messages.
type Publisher interface {
	PublishBatch(ctx context.Context, topic string, messages []Message) error
}

// LocalBuffer defines the interface for persisting messages on publish failure.
type LocalBuffer interface {
	Append(topic string, key string, payload []byte) error
}

// Batcher accumulates messages grouped by topic and flushes them
// at a configurable interval. It is safe for concurrent use.
type Batcher struct {
	publisher     Publisher
	localBuffer   LocalBuffer
	flushInterval time.Duration
	logger        *zap.Logger
	metrics       *metrics.Collector

	mu      sync.Mutex
	pending map[string][]Message

	done   chan struct{}
	closed chan struct{}
}

type Config struct {
	FlushInterval time.Duration
	Logger        *zap.Logger
	LocalBuffer   LocalBuffer
	Metrics       *metrics.Collector
}

// New creates a new Batcher. Call Start to begin the flush loop.
func New(publisher Publisher, cfg Config) (*Batcher, error) {
	if publisher == nil {
		return nil, fmt.Errorf("new batcher: publisher is required")
	}

	if cfg.FlushInterval <= 0 {
		return nil, fmt.Errorf("new batcher: flush interval must be greater than zero")
	}

	logger := cfg.Logger
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Batcher{
		publisher:     publisher,
		localBuffer:   cfg.LocalBuffer,
		flushInterval: cfg.FlushInterval,
		logger:        logger,
		metrics:       cfg.Metrics,
		pending:       make(map[string][]Message),
		done:          make(chan struct{}),
		closed:        make(chan struct{}),
	}, nil
}

// Start begins the periodic flush loop. It blocks until the context
// is cancelled or Stop is called. Typically called in a goroutine.
func (b *Batcher) Start(ctx context.Context) {
	ticker := time.NewTicker(b.flushInterval)
	defer ticker.Stop()
	defer close(b.closed)

	b.logger.Info("batcher started",
		zap.Duration("flush_interval", b.flushInterval),
	)

	for {
		select {
		case <-ctx.Done():
			b.flush(context.Background())
			b.logger.Info("batcher stopped: context cancelled")
			return
		case <-b.done:
			b.flush(context.Background())
			b.logger.Info("batcher stopped")
			return
		case <-ticker.C:
			b.flush(ctx)
		}
	}
}

// Stop signals the batcher to perform a final flush and stop.
// It blocks until the flush loop exits.
func (b *Batcher) Stop() {
	select {
	case <-b.done:
		// already stopped
	default:
		close(b.done)
	}
	<-b.closed
}

// Enqueue adds a message to the batch for its topic.
// The message will be published on the next flush cycle.
func (b *Batcher) Enqueue(msg Message) error {
	if msg.Topic == "" {
		return fmt.Errorf("enqueue: topic is required")
	}

	if msg.Key == "" {
		return fmt.Errorf("enqueue: key is required")
	}

	if len(msg.Payload) == 0 {
		return fmt.Errorf("enqueue: payload is empty")
	}

	b.mu.Lock()
	b.pending[msg.Topic] = append(b.pending[msg.Topic], msg)
	b.mu.Unlock()

	if b.metrics != nil {
		b.metrics.IncEnqueued()
	}

	return nil
}

// PendingCount returns the total number of messages waiting to be flushed.
func (b *Batcher) PendingCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()

	count := 0
	for _, msgs := range b.pending {
		count += len(msgs)
	}
	return count
}

// flush drains all pending messages and publishes them per topic.
func (b *Batcher) flush(ctx context.Context) {
	b.mu.Lock()
	snapshot := b.pending
	b.pending = make(map[string][]Message)
	b.mu.Unlock()

	if len(snapshot) == 0 {
		return
	}

	totalMessages := 0
	for topic, messages := range snapshot {
		totalMessages += len(messages)

		if err := b.publisher.PublishBatch(ctx, topic, messages); err != nil {
			b.logger.Error("batch publish failed",
				zap.String("topic", topic),
				zap.Int("count", len(messages)),
				zap.Error(err),
			)

			if b.metrics != nil {
				b.metrics.IncPublishErrors()
			}

			if b.localBuffer != nil {
				for _, m := range messages {
					if storeErr := b.localBuffer.Append(m.Topic, m.Key, m.Payload); storeErr != nil {
						b.logger.Error("failed to persist message to local buffer",
							zap.String("topic", m.Topic),
							zap.String("key", m.Key),
							zap.Error(storeErr),
						)

						if b.metrics != nil {
							b.metrics.IncBufferErrors()
						}
					}
				}
			}

			continue
		}

		if b.metrics != nil {
			b.metrics.AddDequeued(uint64(len(messages)))
		}

		b.logger.Debug("batch flushed",
			zap.String("topic", topic),
			zap.Int("count", len(messages)),
		)
	}

	b.logger.Info("flush completed",
		zap.Int("topics", len(snapshot)),
		zap.Int("messages", totalMessages),
	)
}
