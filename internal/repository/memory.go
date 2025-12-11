package repository

import (
	"context"

	"github.com/Oleg2210/goshortener/internal/entities"
)

type MemoryRepository struct {
	data map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		data: make(map[string]string),
	}

	return repo
}

func (repo *MemoryRepository) Save(ctx context.Context, id string, url string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if _, exists := repo.data[id]; exists {
		return "", ErrAlreadyExists
	}

	repo.data[id] = url
	return id, nil
}

func (repo *MemoryRepository) BatchSave(ctx context.Context, records []entities.URLRecord) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for _, r := range records {
		if _, exists := repo.data[r.Short]; exists {
			return ErrAlreadyExists
		}
	}

	for _, r := range records {
		repo.data[r.Short] = r.OriginalURL
	}

	return nil
}

func (repo *MemoryRepository) Get(ctx context.Context, id string) (string, bool) {
	select {
	case <-ctx.Done():
		return "", false
	default:
	}

	url, exists := repo.data[id]
	return url, exists
}

func (repo *MemoryRepository) Ping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return false
}
