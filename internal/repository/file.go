package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/Oleg2210/goshortener/internal/entities"
)

// record represents a URL record stored in the file repository.
type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// FileRepository is a persistent URL repository that stores records in a JSON file.
// It wraps a MemoryRepository and synchronizes reads/writes to disk.
type FileRepository struct {
	memoryRepo *MemoryRepository
	path       string
	mu         sync.Mutex
}

// NewFileRepository creates a new FileRepository.
// It loads existing records from the file if it exists.
func NewFileRepository(ctx context.Context, fileStoragePath string) (*FileRepository, error) {
	repo := &FileRepository{
		memoryRepo: NewMemoryRepository(),
		path:       fileStoragePath,
	}

	err := repo.loadDataFromFile(ctx)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// loadDataFromFile reads records from the JSON file and populates the memory repository.
func (repo *FileRepository) loadDataFromFile(ctx context.Context) error {
	bytes, err := os.ReadFile(repo.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var records []record
	if err := json.Unmarshal(bytes, &records); err != nil {
		return err
	}

	for _, r := range records {
		repo.memoryRepo.Save(ctx, r.ShortURL, r.OriginalURL, r.UserID, r.IsDeleted)
	}
	return nil
}

// saveToFile writes all in-memory records to the JSON file.
func (repo *FileRepository) saveToFile() error {
	records := make([]record, 0, len(repo.memoryRepo.data))
	for short, url := range repo.memoryRepo.data {
		records = append(records, record{
			UUID:        short,
			ShortURL:    short,
			OriginalURL: url.OriginalURL,
			UserID:      url.UserID,
			IsDeleted:   url.IsDeleted,
		})
	}

	bytes, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(repo.path, bytes, 0644)
}

// Save stores a new URL record and persists it to the file.
// Returns ErrAlreadyExists if the id already exists.
func (repo *FileRepository) Save(ctx context.Context, id string, url string, userID string, isDeleted bool) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	_, exists := repo.memoryRepo.Get(ctx, id)
	if exists {
		return "", ErrAlreadyExists
	}

	id, err := repo.memoryRepo.Save(ctx, id, url, userID, isDeleted)
	if err != nil {
		return id, err
	}

	return id, repo.saveToFile()
}

// BatchSave stores multiple URL records and persists them to the file.
// Returns ErrAlreadyExists if any of the IDs already exist.
func (repo *FileRepository) BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if err := repo.memoryRepo.BatchSave(ctx, records, userID); err != nil {
		return err
	}

	return repo.saveToFile()
}

// Get retrieves a URL record by id.
func (repo *FileRepository) Get(ctx context.Context, id string) (entities.URLRecord, bool) {
	select {
	case <-ctx.Done():
		return entities.URLRecord{}, false
	default:
	}

	return repo.memoryRepo.Get(ctx, id)
}

// Ping checks the repository availability. Always returns true.
func (repo *FileRepository) Ping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return true
}

// GetUserShortens returns all non-deleted URLs for a specific user.
func (repo *FileRepository) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	select {
	case <-ctx.Done():
		return []entities.URLRecord{}, nil
	default:
	}

	return repo.memoryRepo.GetUserShortens(ctx, userID)
}

// MarkDelete marks a list of URLs as deleted for a given user and persists the change to the file.
func (repo *FileRepository) MarkDelete(ctx context.Context, shorts []string, userID string) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	err := repo.memoryRepo.MarkDelete(ctx, shorts, userID)
	if err != nil {
		return err
	}

	return repo.saveToFile()
}
