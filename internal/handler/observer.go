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

type AuditEvent struct {
	TS     int64  `json:"ts"`
	Action string `json:"action"` // shorten | follow
	UserID string `json:"user_id,omitempty"`
	URL    string `json:"url"`
}

type AuditObserver interface {
	Notify(ctx context.Context, event AuditEvent) error
}

type AuditPublisher struct {
	observers []AuditObserver
}

func NewAuditPublisher() *AuditPublisher {
	return &AuditPublisher{}
}

func (p *AuditPublisher) Register(o AuditObserver) {
	p.observers = append(p.observers, o)
}

func (p *AuditPublisher) Publish(ctx context.Context, event AuditEvent) {
	for _, o := range p.observers {
		go func(obs AuditObserver) {
			_ = obs.Notify(ctx, event)
		}(o)
	}
}

type FileObserver struct {
	file *os.File
	mu   sync.Mutex
}

func NewFileObserver(path string) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return &FileObserver{file: f}, nil
}

func (o *FileObserver) Notify(ctx context.Context, event AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	o.mu.Lock()
	defer o.mu.Unlock()

	_, err = o.file.Write(append(data, '\n'))
	return err
}

type HTTPObserver struct {
	client  *http.Client
	url     string
	timeout time.Duration
}

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
