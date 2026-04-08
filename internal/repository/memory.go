package repository

import (
	"context"
	"sync"

	"github.com/Oleg2210/goshortener/internal/entities"
)

// generate:reset
type MemoryRecord struct {
	OriginalURL string
	UserID      string
	IsDeleted   bool
}

// MemoryRepository is an in-memory implementation of a URL repository.
// It allows fast storage and retrieval of URLs without using persistent storage.
type MemoryRepository struct {
	mu       sync.RWMutex
	data     map[string]MemoryRecord
	userData map[string]map[string]string
}

// expectedURLs is the initial map capacity for memory optimization.
const expectedURLs = 10000

// NewMemoryRepository creates a new MemoryRepository instance.
// It preallocates maps to reduce memory allocations during runtime.
func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		data:     make(map[string]MemoryRecord, expectedURLs),
		userData: make(map[string]map[string]string, expectedURLs),
	}

	return repo
}

// Save stores a new URL in the repository.
// Returns ErrAlreadyExists if the id already exists.

func (repo *MemoryRepository) Save(ctx context.Context, id string, url string, userID string, isDeleted bool) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if _, exists := repo.data[id]; exists {
		return "", ErrAlreadyExists
	}

	repo.data[id] = MemoryRecord{OriginalURL: url, UserID: userID, IsDeleted: isDeleted}
	if repo.userData[userID] == nil {
		repo.userData[userID] = make(map[string]string)
	}
	repo.userData[userID][id] = url
	return id, nil
}

// BatchSave stores multiple URLs at once.
// Returns ErrAlreadyExists if any of the IDs already exist.
func (repo *MemoryRepository) BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, r := range records {
		if _, exists := repo.data[r.Short]; exists {
			return ErrAlreadyExists
		}
	}

	for _, r := range records {
		repo.data[r.Short] = MemoryRecord{OriginalURL: r.OriginalURL, UserID: userID, IsDeleted: false}

		if repo.userData[userID] == nil {
			repo.userData[userID] = make(map[string]string)
		}
		repo.userData[userID][r.Short] = r.OriginalURL
	}

	return nil
}

// Get retrieves a URL record by id.
// Returns false if the record does not exist.
func (repo *MemoryRepository) Get(ctx context.Context, id string) (entities.URLRecord, bool) {
	select {
	case <-ctx.Done():
		return entities.URLRecord{}, false
	default:
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	url, exists := repo.data[id]
	return entities.URLRecord{OriginalURL: url.OriginalURL, Short: id, IsDeleted: url.IsDeleted}, exists
}

// Ping checks the repository availability.
// Always returns false for in-memory storage.
func (repo *MemoryRepository) Ping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return false
}

// GetUserShortens returns all non-deleted URLs for a specific user.
func (repo *MemoryRepository) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	select {
	case <-ctx.Done():
		return []entities.URLRecord{}, nil
	default:
	}

	repo.mu.RLock()
	defer repo.mu.RUnlock()

	if repo.userData[userID] == nil {
		return []entities.URLRecord{}, nil
	}

	shortens := make([]entities.URLRecord, 0, len(repo.userData[userID]))
	for k, v := range repo.userData[userID] {
		if !repo.data[k].IsDeleted {
			shortens = append(shortens, entities.URLRecord{OriginalURL: v, Short: k})
		}
	}

	return shortens, nil
}

// MarkDelete marks a list of URLs as deleted for a given user.
func (repo *MemoryRepository) MarkDelete(ctx context.Context, shorts []string, userID string) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	for _, short := range shorts {
		url, ok := repo.data[short]
		if !ok {
			continue
		}

		if url.UserID == userID {
			repo.data[short] = MemoryRecord{OriginalURL: url.OriginalURL, UserID: userID, IsDeleted: true}
		}
	}

	return nil
}

// GetStatistic returns all users and url counts
func (repo *MemoryRepository) GetStatistic(ctx context.Context) (int, int, error) {
	select {
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	default:
	}

	return len(repo.data), len(repo.userData), nil
}
