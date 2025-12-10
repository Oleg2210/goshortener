package repository

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/Oleg2210/goshortener/internal/entities"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileRepository struct {
	memoryRepo *MemoryRepository
	path       string
	mu         sync.Mutex
}

func NewFileRepository(fileStoragePath string) (*FileRepository, error) {
	repo := &FileRepository{
		memoryRepo: NewMemoryRepository(),
		path:       fileStoragePath,
	}

	err := repo.loadDataFromFile()
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func (repo *FileRepository) loadDataFromFile() error {
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
		repo.memoryRepo.Save(r.ShortURL, r.OriginalURL)
	}
	return nil
}

func (repo *FileRepository) saveToFile() error {
	records := make([]record, 0, len(repo.memoryRepo.data))
	for short, original := range repo.memoryRepo.data {
		records = append(records, record{
			UUID:        short,
			ShortURL:    short,
			OriginalURL: original,
		})
	}

	bytes, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(repo.path, bytes, 0644)
}

func (repo *FileRepository) Save(id string, url string) (string, error) {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	_, exists := repo.memoryRepo.Get(id)
	if exists {
		return "", ErrAlreadyExists
	}

	id, err := repo.memoryRepo.Save(id, url)
	if err != nil {
		return id, err
	}

	return id, repo.saveToFile()
}

func (repo *FileRepository) BatchSave(
	records []entities.URLRecord,
) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	err := repo.memoryRepo.BatchSave(records)
	if err != nil {
		return err
	}

	return repo.saveToFile()
}

func (repo *FileRepository) Get(id string) (string, bool) {
	return repo.memoryRepo.Get(id)
}

func (repo *FileRepository) Ping() bool {
	return false
}
