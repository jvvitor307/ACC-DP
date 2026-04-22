package redpanda

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	DefaultPublishTimeout        = 5 * time.Second
	DefaultSchemaRegistryTimeout = 5 * time.Second

	confluentMagicByte        = byte(0)
	schemaRegistryContentType = "application/vnd.schemaregistry.v1+json"
)

type Topics struct {
	Physics  string
	Graphics string
	Static   string
}

type Schemas struct {
	Physics  string
	Graphics string
	Static   string
}

type SchemaIDs struct {
	Physics  int
	Graphics int
	Static   int
}

type Config struct {
	Brokers               []string
	Topics                Topics
	Schemas               Schemas
	SchemaIDs             SchemaIDs
	SchemaRegistryURL     string
	PublishTimeout        time.Duration
	SchemaRegistryTimeout time.Duration
}

// BatchMessage represents a single message within a batch publish call.
type BatchMessage struct {
	Topic   string
	Key     string
	Payload []byte
}

type publisherClient interface {
	Produce(ctx context.Context, topic string, key []byte, payload []byte) error
	ProduceBatch(ctx context.Context, records []*kgo.Record) error
	Close()
}

type schemaRegistryClient interface {
	RegisterSchema(ctx context.Context, subject string, schema string) (int, error)
	SetCompatibility(ctx context.Context, subject string, level string) error
}

type franzClient struct {
	client *kgo.Client
}

type schemaRegistryHTTPClient struct {
	baseURL string
	client  *http.Client
}

type Publisher struct {
	client         publisherClient
	topics         Topics
	schemaIDs      map[string]int
	publishTimeout time.Duration
}

func NewPublisher(cfg Config) (*Publisher, error) {
	cfg = cfg.withDefaults()

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("new publisher: %w", err)
	}

	client, err := newFranzClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("new publisher: %w", err)
	}

	schemaIDs := cfg.SchemaIDs
	if !schemaIDs.isSet() {
		registryClient := newSchemaRegistryHTTPClient(cfg.SchemaRegistryURL, cfg.SchemaRegistryTimeout)

		resolvedSchemaIDs, resolveErr := resolveSchemaIDs(context.Background(), registryClient, cfg.Topics, cfg.Schemas)
		if resolveErr != nil {
			client.Close()
			return nil, fmt.Errorf("new publisher: %w", resolveErr)
		}

		schemaIDs = resolvedSchemaIDs
	}

	cfg.SchemaIDs = schemaIDs

	publisher, err := newPublisherWithClient(cfg, client)
	if err != nil {
		client.Close()
		return nil, err
	}

	return publisher, nil
}

func newPublisherWithClient(cfg Config, client publisherClient) (*Publisher, error) {
	cfg = cfg.withDefaults()

	if err := cfg.Topics.validate(); err != nil {
		return nil, fmt.Errorf("new publisher: %w", err)
	}

	if err := cfg.SchemaIDs.validate(); err != nil {
		return nil, fmt.Errorf("new publisher: %w", err)
	}

	if cfg.PublishTimeout <= 0 {
		return nil, fmt.Errorf("new publisher: publish timeout must be greater than zero")
	}

	if client == nil {
		return nil, fmt.Errorf("new publisher: client is required")
	}

	return &Publisher{
		client: client,
		topics: cfg.Topics,
		schemaIDs: map[string]int{
			cfg.Topics.Physics:  cfg.SchemaIDs.Physics,
			cfg.Topics.Graphics: cfg.SchemaIDs.Graphics,
			cfg.Topics.Static:   cfg.SchemaIDs.Static,
		},
		publishTimeout: cfg.PublishTimeout,
	}, nil
}

func (p *Publisher) PublishPhysics(ctx context.Context, key string, payload []byte) error {
	return p.publish(ctx, p.topics.Physics, key, payload)
}

func (p *Publisher) PublishGraphics(ctx context.Context, key string, payload []byte) error {
	return p.publish(ctx, p.topics.Graphics, key, payload)
}

func (p *Publisher) PublishStatic(ctx context.Context, key string, payload []byte) error {
	return p.publish(ctx, p.topics.Static, key, payload)
}

// PublishBatch publishes a slice of messages to a single topic in one
// synchronous produce call. Each message payload is framed with the
// topic's schema ID before sending.
func (p *Publisher) PublishBatch(ctx context.Context, topic string, messages []BatchMessage) error {
	if p == nil {
		return fmt.Errorf("publish batch: publisher is nil")
	}

	if len(messages) == 0 {
		return nil
	}

	schemaID, exists := p.schemaIDs[topic]
	if !exists {
		return fmt.Errorf("publish batch topic %s: schema id not configured", topic)
	}

	records := make([]*kgo.Record, 0, len(messages))
	for _, msg := range messages {
		framedPayload, err := framePayload(schemaID, msg.Payload)
		if err != nil {
			return fmt.Errorf("publish batch topic %s: %w", topic, err)
		}

		records = append(records, &kgo.Record{
			Topic: topic,
			Key:   []byte(msg.Key),
			Value: framedPayload,
		})
	}

	publishCtx, cancel := context.WithTimeout(ctx, p.publishTimeout)
	defer cancel()

	if err := p.client.ProduceBatch(publishCtx, records); err != nil {
		return fmt.Errorf("publish batch topic %s: %w", topic, err)
	}

	return nil
}

func (p *Publisher) Close() {
	if p == nil || p.client == nil {
		return
	}

	p.client.Close()
}

func (p *Publisher) publish(ctx context.Context, topic, key string, payload []byte) error {
	if p == nil {
		return fmt.Errorf("publish: publisher is nil")
	}

	if strings.TrimSpace(topic) == "" {
		return fmt.Errorf("publish: topic is required")
	}

	trimmedKey := strings.TrimSpace(key)
	if trimmedKey == "" {
		return fmt.Errorf("publish topic %s: key is required", topic)
	}

	if len(payload) == 0 {
		return fmt.Errorf("publish topic %s: payload is empty", topic)
	}

	schemaID, exists := p.schemaIDs[topic]
	if !exists {
		return fmt.Errorf("publish topic %s: schema id not configured", topic)
	}

	framedPayload, err := framePayload(schemaID, payload)
	if err != nil {
		return fmt.Errorf("publish topic %s: %w", topic, err)
	}

	publishCtx, cancel := context.WithTimeout(ctx, p.publishTimeout)
	defer cancel()

	if err := p.client.Produce(publishCtx, topic, []byte(trimmedKey), framedPayload); err != nil {
		return fmt.Errorf("publish topic %s: %w", topic, err)
	}

	return nil
}

func framePayload(schemaID int, payload []byte) ([]byte, error) {
	if schemaID <= 0 {
		return nil, fmt.Errorf("schema id must be greater than zero")
	}

	if len(payload) == 0 {
		return nil, fmt.Errorf("payload is empty")
	}

	framed := make([]byte, 5+len(payload))
	framed[0] = confluentMagicByte
	binary.BigEndian.PutUint32(framed[1:5], uint32(schemaID))
	copy(framed[5:], payload)

	return framed, nil
}

func newFranzClient(cfg Config) (publisherClient, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.RecordDeliveryTimeout(cfg.PublishTimeout),
	)
	if err != nil {
		return nil, fmt.Errorf("create franz-go client: %w", err)
	}

	return &franzClient{client: client}, nil
}

func (c *franzClient) Produce(ctx context.Context, topic string, key []byte, payload []byte) error {
	results := c.client.ProduceSync(ctx, &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: payload,
	})

	if err := results.FirstErr(); err != nil {
		return err
	}

	return nil
}

func (c *franzClient) ProduceBatch(ctx context.Context, records []*kgo.Record) error {
	results := c.client.ProduceSync(ctx, records...)

	if err := results.FirstErr(); err != nil {
		return err
	}

	return nil
}

func (c *franzClient) Close() {
	c.client.Close()
}

func (cfg Config) withDefaults() Config {
	if cfg.PublishTimeout <= 0 {
		cfg.PublishTimeout = DefaultPublishTimeout
	}

	if cfg.SchemaRegistryTimeout <= 0 {
		cfg.SchemaRegistryTimeout = DefaultSchemaRegistryTimeout
	}

	return cfg
}

func (cfg Config) validate() error {
	if err := validateBrokers(cfg.Brokers); err != nil {
		return err
	}

	if err := cfg.Topics.validate(); err != nil {
		return err
	}

	if cfg.SchemaIDs.isSet() {
		return cfg.SchemaIDs.validate()
	}

	if strings.TrimSpace(cfg.SchemaRegistryURL) == "" {
		return fmt.Errorf("schema registry url is required when schema ids are not preloaded")
	}

	if err := cfg.Schemas.validate(); err != nil {
		return err
	}

	return nil
}

func validateBrokers(brokers []string) error {
	if len(brokers) == 0 {
		return fmt.Errorf("brokers must contain at least one entry")
	}

	for _, broker := range brokers {
		if strings.TrimSpace(broker) == "" {
			return fmt.Errorf("brokers must not contain empty entries")
		}
	}

	return nil
}

func (t Topics) validate() error {
	if strings.TrimSpace(t.Physics) == "" {
		return fmt.Errorf("physics topic is required")
	}

	if strings.TrimSpace(t.Graphics) == "" {
		return fmt.Errorf("graphics topic is required")
	}

	if strings.TrimSpace(t.Static) == "" {
		return fmt.Errorf("static topic is required")
	}

	return nil
}

func (s Schemas) validate() error {
	if strings.TrimSpace(s.Physics) == "" {
		return fmt.Errorf("physics schema is required")
	}

	if strings.TrimSpace(s.Graphics) == "" {
		return fmt.Errorf("graphics schema is required")
	}

	if strings.TrimSpace(s.Static) == "" {
		return fmt.Errorf("static schema is required")
	}

	return nil
}

func (s SchemaIDs) isSet() bool {
	return s.Physics > 0 || s.Graphics > 0 || s.Static > 0
}

func (s SchemaIDs) validate() error {
	if s.Physics <= 0 {
		return fmt.Errorf("physics schema id must be greater than zero")
	}

	if s.Graphics <= 0 {
		return fmt.Errorf("graphics schema id must be greater than zero")
	}

	if s.Static <= 0 {
		return fmt.Errorf("static schema id must be greater than zero")
	}

	return nil
}

func resolveSchemaIDs(ctx context.Context, registryClient schemaRegistryClient, topics Topics, schemas Schemas) (SchemaIDs, error) {
	if registryClient == nil {
		return SchemaIDs{}, fmt.Errorf("schema registry client is required")
	}

	physicsID, err := registerTopicSchema(ctx, registryClient, topics.Physics, schemas.Physics)
	if err != nil {
		return SchemaIDs{}, fmt.Errorf("resolve physics schema id: %w", err)
	}

	graphicsID, err := registerTopicSchema(ctx, registryClient, topics.Graphics, schemas.Graphics)
	if err != nil {
		return SchemaIDs{}, fmt.Errorf("resolve graphics schema id: %w", err)
	}

	staticID, err := registerTopicSchema(ctx, registryClient, topics.Static, schemas.Static)
	if err != nil {
		return SchemaIDs{}, fmt.Errorf("resolve static schema id: %w", err)
	}

	return SchemaIDs{
		Physics:  physicsID,
		Graphics: graphicsID,
		Static:   staticID,
	}, nil
}

func registerTopicSchema(ctx context.Context, registryClient schemaRegistryClient, topic string, schema string) (int, error) {
	subject := fmt.Sprintf("%s-value", strings.TrimSpace(topic))

	schemaID, err := registryClient.RegisterSchema(ctx, subject, schema)
	if err != nil {
		return 0, fmt.Errorf("register subject %s: %w", subject, err)
	}

	if err := registryClient.SetCompatibility(ctx, subject, "BACKWARD"); err != nil {
		return 0, fmt.Errorf("set compatibility for subject %s: %w", subject, err)
	}

	return schemaID, nil
}

func newSchemaRegistryHTTPClient(baseURL string, timeout time.Duration) schemaRegistryClient {
	return &schemaRegistryHTTPClient{
		baseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

type registerSchemaRequest struct {
	Schema string `json:"schema"`
}

type registerSchemaResponse struct {
	ID int `json:"id"`
}

func (c *schemaRegistryHTTPClient) RegisterSchema(ctx context.Context, subject string, schema string) (int, error) {
	if strings.TrimSpace(subject) == "" {
		return 0, fmt.Errorf("subject is required")
	}

	if strings.TrimSpace(schema) == "" {
		return 0, fmt.Errorf("schema is required")
	}

	requestBody, err := json.Marshal(registerSchemaRequest{Schema: schema})
	if err != nil {
		return 0, fmt.Errorf("marshal register schema request: %w", err)
	}

	requestURL := fmt.Sprintf("%s/subjects/%s/versions", c.baseURL, url.PathEscape(subject))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(requestBody))
	if err != nil {
		return 0, fmt.Errorf("create register schema request: %w", err)
	}

	request.Header.Set("Content-Type", schemaRegistryContentType)
	request.Header.Set("Accept", schemaRegistryContentType)

	response, err := c.client.Do(request)
	if err != nil {
		return 0, fmt.Errorf("execute register schema request: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return 0, fmt.Errorf("read register schema response: %w", err)
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return 0, fmt.Errorf("register schema returned status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var parsedResponse registerSchemaResponse
	if err := json.Unmarshal(responseBody, &parsedResponse); err != nil {
		return 0, fmt.Errorf("decode register schema response: %w", err)
	}

	if parsedResponse.ID <= 0 {
		return 0, fmt.Errorf("register schema response has invalid id")
	}

	return parsedResponse.ID, nil
}

func (c *schemaRegistryHTTPClient) SetCompatibility(ctx context.Context, subject string, level string) error {
	if strings.TrimSpace(subject) == "" {
		return fmt.Errorf("subject is required")
	}

	if strings.TrimSpace(level) == "" {
		return fmt.Errorf("compatibility level is required")
	}

	requestBody := []byte(fmt.Sprintf(`{"compatibility": "%s"}`, level))
	requestURL := fmt.Sprintf("%s/config/%s", c.baseURL, url.PathEscape(subject))
	
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, requestURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("create set compatibility request: %w", err)
	}

	request.Header.Set("Content-Type", schemaRegistryContentType)
	request.Header.Set("Accept", schemaRegistryContentType)

	response, err := c.client.Do(request)
	if err != nil {
		return fmt.Errorf("execute set compatibility request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		responseBody, _ := io.ReadAll(io.LimitReader(response.Body, 1<<20))
		return fmt.Errorf("set compatibility returned status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	return nil
}
