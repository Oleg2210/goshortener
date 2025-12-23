package repository

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/Oleg2210/goshortener/internal/entities"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

type FileRepository struct {
	memoryRepo *MemoryRepository
	path       string
	mu         sync.Mutex
}

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

func (repo *FileRepository) Get(ctx context.Context, id string) (entities.URLRecord, bool) {
	select {
	case <-ctx.Done():
		return entities.URLRecord{}, false
	default:
	}

	return repo.memoryRepo.Get(ctx, id)
}

func (repo *FileRepository) Ping(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	return true
}

func (repo *FileRepository) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	select {
	case <-ctx.Done():
		return []entities.URLRecord{}, nil
	default:
	}

	return repo.memoryRepo.GetUserShortens(ctx, userID)
}

func (repo *FileRepository) MarkDelete(ctx context.Context, short string, userID string) error {
	select {
	case <-ctx.Done():
		return nil
	default:
	}

	err := repo.memoryRepo.MarkDelete(ctx, short, userID)

	if err != nil {
		return err
	}

	return repo.saveToFile()
}
