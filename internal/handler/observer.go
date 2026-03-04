package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"go.uber.org/zap"
)

// AuditEvent represents an event for auditing purposes.
// TS is the Unix timestamp, Action is the type of event ("shorten" | "follow"),
// UserID is optional, and URL is the affected URL.
type AuditEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"` // shorten | follow
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

// AuditObserver defines the interface for receiving audit events.
type AuditObserver interface {
	// Notify is called when an audit event occurs.
	Notify(ctx context.Context, event AuditEvent) error
}

// AuditPublisher is responsible for publishing audit events to registered observers.
// It maintains a buffered channel for events, runs a pool of worker goroutines,
// and logs errors using zap.Logger.
type AuditPublisher struct {
	mu        sync.RWMutex
	observers map[AuditObserver]struct{}

	events chan AuditEvent
	wg     sync.WaitGroup

	logger *zap.Logger
}

// NewAuditPublisher creates a new AuditPublisher instance with a fixed number of workers
// and a buffered channel of specified size. Optionally accepts a zap.Logger for logging.
func NewAuditPublisher(workers int, buffer int, logger *zap.Logger) *AuditPublisher {
	p := &AuditPublisher{
		observers: make(map[AuditObserver]struct{}),
		events:    make(chan AuditEvent, buffer),
		logger:    logger,
	}

	// Start worker goroutines to process events asynchronously
	for i := 0; i < workers; i++ {
		p.wg.Add(1)
		go p.worker()
	}

	return p
}

// worker runs in a goroutine and processes events from the channel.
// It calls Notify on all registered observers for each event and logs any errors.
func (p *AuditPublisher) worker() {
	defer p.wg.Done()

	for event := range p.events {
		p.mu.RLock()
		for o := range p.observers {
			err := o.Notify(context.Background(), event)
			if err != nil && p.logger != nil {
				p.logger.Error("Failed to notify observer", zap.Any("event", event), zap.Error(err))
			}
		}
		p.mu.RUnlock()
	}
}

// Publish sends an audit event to the publisher's queue.
// If the queue is full, it logs an error via zap.Logger and drops the event.
func (p *AuditPublisher) Publish(ctx context.Context, event AuditEvent) {
	select {
	case p.events <- event:
		// Successfully queued the event
	default:
		if p.logger != nil {
			p.logger.Error("AuditPublisher queue full, dropping event", zap.Any("event", event))
		}
	}
}

// Register adds a new observer to the publisher. Thread-safe.
func (p *AuditPublisher) Register(o AuditObserver) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.observers[o] = struct{}{}
}

// Unregister removes an observer from the publisher. Thread-safe.
func (p *AuditPublisher) Unregister(o AuditObserver) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.observers, o)
}

// Close gracefully shuts down the publisher, closing the event channel
// and waiting for all worker goroutines to finish processing.
func (p *AuditPublisher) Close() {
	close(p.events)
	p.wg.Wait()
}

// FileObserver writes audit events to a file in JSON format.
type FileObserver struct {
	file   *os.File
	mu     sync.Mutex
	enc    *json.Encoder
	logger *zap.Logger
}

// NewFileObserver creates a new FileObserver for the given file path.
// It opens the file for appending and prepares a JSON encoder.
func NewFileObserver(path string, logger *zap.Logger) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileObserver{
		file:   f,
		enc:    json.NewEncoder(f),
		logger: logger,
	}, nil
}

// Notify writes the audit event to the file in JSON format.
// Thread-safe using a mutex. Logs any errors via zap.Logger.
func (o *FileObserver) Notify(ctx context.Context, event AuditEvent) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	err := o.enc.Encode(event)
	if err != nil && o.logger != nil {
		o.logger.Error("Failed to write audit event to file", zap.Any("event", event), zap.Error(err))
	}
	return err
}

// Close closes the underlying file. Should be called when the observer is no longer needed.
func (o *FileObserver) Close() error {
	return o.file.Close()
}

// HTTPObserver sends audit events via HTTP POST to a configured endpoint using retryablehttp.
type HTTPObserver struct {
	client *retryablehttp.Client
	url    string
	logger *zap.Logger
}

// NewHTTPObserver creates a new HTTPObserver for the given endpoint URL.
// Uses retryablehttp with default retry/backoff settings.
func NewHTTPObserver(url string, logger *zap.Logger) *HTTPObserver {
	client := retryablehttp.NewClient()
	client.RetryMax = 3
	client.RetryWaitMin = 100 * time.Millisecond
	client.RetryWaitMax = 2 * time.Second
	client.Logger = nil // disable default retryablehttp logging

	return &HTTPObserver{
		client: client,
		url:    url,
		logger: logger,
	}
}

// Notify sends the audit event to the HTTP endpoint as JSON.
// Logs any errors via zap.Logger.
func (o *HTTPObserver) Notify(ctx context.Context, event AuditEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		if o.logger != nil {
			o.logger.Error("Failed to marshal audit event", zap.Any("event", event), zap.Error(err))
		}
		return err
	}

	req, err := retryablehttp.NewRequest(
		http.MethodPost,
		o.url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		if o.logger != nil {
			o.logger.Error("Failed to create HTTP request", zap.Any("event", event), zap.Error(err))
		}
		return err
	}

	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		if o.logger != nil {
			o.logger.Error("Failed to send HTTP audit event", zap.Any("event", event), zap.Error(err))
		}
		return err
	}
	defer resp.Body.Close()

	return nil
}
