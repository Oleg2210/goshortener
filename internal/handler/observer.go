package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"
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
type AuditPublisher struct {
	observers []AuditObserver
}

// NewAuditPublisher creates a new AuditPublisher instance.
func NewAuditPublisher() *AuditPublisher {
	return &AuditPublisher{}
}

// Register adds a new observer to the publisher.
func (p *AuditPublisher) Register(o AuditObserver) {
	p.observers = append(p.observers, o)
}

// Publish sends the audit event to all registered observers asynchronously.
func (p *AuditPublisher) Publish(ctx context.Context, event AuditEvent) {
	for _, o := range p.observers {
		go func(obs AuditObserver) {
			_ = obs.Notify(ctx, event)
		}(o)
	}
}

// FileObserver writes audit events to a file in JSON format.
type FileObserver struct {
	file *os.File
	mu   sync.Mutex
	enc  *json.Encoder
}

// NewFileObserver creates a new FileObserver for the given file path.
func NewFileObserver(path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileObserver{
		file: f,
		enc:  json.NewEncoder(f),
	}, nil
}

// Notify writes the audit event to the file.
func (o *FileObserver) Notify(ctx context.Context, event AuditEvent) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.enc.Encode(event)
}

// HTTPObserver sends audit events via HTTP POST to a configured endpoint.
type HTTPObserver struct {
	client  *http.Client
	url     string
	timeout time.Duration
}

// NewHTTPObserver creates a new HTTPObserver for the given endpoint URL.
func NewHTTPObserver(url string) *HTTPObserver {
	timeout := 5 * time.Second

	return &HTTPObserver{
		client: &http.Client{
			Timeout: timeout,
		},
		url:     url,
		timeout: timeout,
	}
}

// Notify sends the audit event to the HTTP endpoint as JSON.
func (o *HTTPObserver) Notify(ctx context.Context, event AuditEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		o.url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
