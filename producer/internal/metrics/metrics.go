package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type Snapshot struct {
	Enqueued      uint64
	Dequeued      uint64
	PublishErrors uint64
	BufferErrors  uint64
	RetrySuccess  uint64
	RetryErrors   uint64
	BacklogSize   uint64
	Timestamp     time.Time
}

type Collector struct {
	enqueued      atomic.Uint64
	dequeued      atomic.Uint64
	publishErrors atomic.Uint64
	bufferErrors  atomic.Uint64
	retrySuccess  atomic.Uint64
	retryErrors   atomic.Uint64
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) IncEnqueued() {
	c.enqueued.Add(1)
}

func (c *Collector) AddDequeued(n uint64) {
	c.dequeued.Add(n)
}

func (c *Collector) IncPublishErrors() {
	c.publishErrors.Add(1)
}

func (c *Collector) IncBufferErrors() {
	c.bufferErrors.Add(1)
}

func (c *Collector) AddRetrySuccess(n uint64) {
	c.retrySuccess.Add(n)
}

func (c *Collector) IncRetryErrors() {
	c.retryErrors.Add(1)
}

func (c *Collector) Snapshot(backlogSize uint64) Snapshot {
	return Snapshot{
		Enqueued:      c.enqueued.Load(),
		Dequeued:      c.dequeued.Load(),
		PublishErrors: c.publishErrors.Load(),
		BufferErrors:  c.bufferErrors.Load(),
		RetrySuccess:  c.retrySuccess.Load(),
		RetryErrors:   c.retryErrors.Load(),
		BacklogSize:   backlogSize,
		Timestamp:     time.Now().UTC(),
	}
}

type BacklogProvider interface {
	PendingCount() int
	BufferCount() int
}

type Reporter struct {
	collector       *Collector
	backlogProvider BacklogProvider
	logger          *zap.Logger
	interval        time.Duration
	wg              sync.WaitGroup
}

func NewReporter(collector *Collector, provider BacklogProvider, logger *zap.Logger, interval time.Duration) *Reporter {
	if logger == nil {
		logger = zap.NewNop()
	}

	if interval <= 0 {
		interval = 30 * time.Second
	}

	return &Reporter{
		collector:       collector,
		backlogProvider: provider,
		logger:          logger,
		interval:        interval,
	}
}

func (r *Reporter) Start(ctx interface {
	Done() <-chan struct{}
}) {
	r.wg.Add(1)
	defer r.wg.Done()

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.Info("metrics reporter stopped")
			return
		case <-ticker.C:
			r.report()
		}
	}
}

func (r *Reporter) Wait() {
	r.wg.Wait()
}

func (r *Reporter) report() {
	var backlogSize uint64
	if r.backlogProvider != nil {
		backlogSize = uint64(r.backlogProvider.PendingCount() + r.backlogProvider.BufferCount())
	}

	snap := r.collector.Snapshot(backlogSize)

	r.logger.Info("backlog metrics",
		zap.Uint64("enqueued", snap.Enqueued),
		zap.Uint64("dequeued", snap.Dequeued),
		zap.Uint64("publish_errors", snap.PublishErrors),
		zap.Uint64("buffer_errors", snap.BufferErrors),
		zap.Uint64("retry_success", snap.RetrySuccess),
		zap.Uint64("retry_errors", snap.RetryErrors),
		zap.Uint64("backlog_size", snap.BacklogSize),
	)
}
