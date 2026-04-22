package badger

import (
	"encoding/json"
	"fmt"
	"time"

	badgerdb "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// Buffer is a local persistent buffer powered by BadgerDB.
type Buffer struct {
	db *badgerdb.DB
}

// BufferedMessage represents a message stored in the local buffer.
type BufferedMessage struct {
	ID        string    `json:"id"`
	Topic     string    `json:"topic"`
	Key       string    `json:"key"`
	Payload   []byte    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

// New initializes and opens a new BadgerDB buffer.
func New(path string) (*Buffer, error) {
	if path == "" {
		return nil, fmt.Errorf("badger path is required")
	}

	opts := badgerdb.DefaultOptions(path).WithLoggingLevel(badgerdb.WARNING)
	db, err := badgerdb.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("open badger db: %w", err)
	}

	return &Buffer{db: db}, nil
}

// Append generates a new UUID and persists the message to the store.
func (b *Buffer) Append(topic string, key string, payload []byte) error {
	id := uuid.New().String()
	msg := BufferedMessage{
		ID:        id,
		Topic:     topic,
		Key:       key,
		Payload:   payload,
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal buffered message: %w", err)
	}

	return b.db.Update(func(txn *badgerdb.Txn) error {
		return txn.Set([]byte(id), data)
	})
}

// ReadPending returns all pending messages from the store.
func (b *Buffer) ReadPending() ([]BufferedMessage, error) {
	var messages []BufferedMessage

	err := b.db.View(func(txn *badgerdb.Txn) error {
		opts := badgerdb.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var msg BufferedMessage
				if err := json.Unmarshal(val, &msg); err != nil {
					return err
				}
				messages = append(messages, msg)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("iterate badger db: %w", err)
	}

	return messages, nil
}

// Ack removes a securely published message from the store.
func (b *Buffer) Ack(id string) error {
	if id == "" {
		return fmt.Errorf("id is required for ack")
	}

	return b.db.Update(func(txn *badgerdb.Txn) error {
		return txn.Delete([]byte(id))
	})
}

// Close gracefully stops the BadgerDB engine.
func (b *Buffer) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}
