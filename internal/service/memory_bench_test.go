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
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = svc.Shorten(context.Background(), url, "user1")
		}
	})
}
