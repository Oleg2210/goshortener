package service

import (
	"context"
	"testing"

	"github.com/Oleg2210/goshortener/internal/repository"
)

func BenchmarkShorten(b *testing.B) {
	repo := repository.NewMemoryRepository()
	svc := NewShortenerService(repo, 5, 10)
	url := "https://example.com/long/path/to/resource"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = svc.Shorten(context.Background(), url, "user1")
	}
}
