package repository

import (
	"context"

	"github.com/Oleg2210/goshortener/internal/entities"
)

type MemoryRecord struct {
	OriginalURL string
	UserID      string
}

type MemoryRepository struct {
	data     map[string]MemoryRecord
	userData map[string]map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		data:     make(map[string]MemoryRecord),
		userData: make(map[string]map[string]string),
	}

	return repo
}

func (repo *MemoryRepository) Save(ctx context.Context, id string, url string, userID string) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	if _, exists := repo.data[id]; exists {
		return "", ErrAlreadyExists
	}

	repo.data[id] = MemoryRecord{OriginalURL: url, UserID: userID}
	if repo.userData[userID] == nil {
		repo.userData[userID] = make(map[string]string)
	}
	repo.userData[userID][id] = url
	return id, nil
}

func (repo *MemoryRepository) BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error {
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
		repo.data[r.Short] = MemoryRecord{OriginalURL: r.OriginalURL, UserID: userID}

		if repo.userData[userID] == nil {
			repo.userData[userID] = make(map[string]string)
		}
		repo.userData[userID][r.Short] = r.OriginalURL
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
	return url.OriginalURL, exists
}

func (repo *MemoryRepository) Ping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return false
}

func (repo *MemoryRepository) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	select {
	case <-ctx.Done():
		return []entities.URLRecord{}, nil
	default:
	}

	if repo.userData[userID] == nil {
		return []entities.URLRecord{}, nil
	}

	shortens := make([]entities.URLRecord, len(repo.userData[userID]))

	for k, v := range repo.userData[userID] {
		shortens = append(shortens, entities.URLRecord{OriginalURL: v, Short: k})
	}

	return shortens, nil
}
