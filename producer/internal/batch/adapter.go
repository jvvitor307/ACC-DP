package batch

import (
	"context"

	"acc-dp/producer/internal/broker/redpanda"
)

// PublisherAdapter adapts a *redpanda.Publisher to the batch.Publisher interface.
type PublisherAdapter struct {
	publisher *redpanda.Publisher
}

// NewPublisherAdapter creates a new adapter wrapping the given redpanda publisher.
func NewPublisherAdapter(publisher *redpanda.Publisher) *PublisherAdapter {
	return &PublisherAdapter{publisher: publisher}
}

// PublishBatch converts batch messages to redpanda batch messages and delegates
// to the underlying publisher.
func (a *PublisherAdapter) PublishBatch(ctx context.Context, topic string, messages []Message) error {
	rpMessages := make([]redpanda.BatchMessage, len(messages))
	for i, msg := range messages {
		rpMessages[i] = redpanda.BatchMessage{
			Topic:   msg.Topic,
			Key:     msg.Key,
			Payload: msg.Payload,
		}
	}

	return a.publisher.PublishBatch(ctx, topic, rpMessages)
}
