package repository

import (
	"encoding/json"
	"os"
)

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type MemoryRepository struct {
	data map[string]string
	path string
}

func NewMemoryRepository(fileStoragePath string) *MemoryRepository {
	repo := &MemoryRepository{
		data: make(map[string]string),
	}

	repo.loadDataFromFile()
	return repo
}

func (repo *MemoryRepository) loadDataFromFile() {
	bytes, err := os.ReadFile(repo.path)
	if err != nil {
		return
	}

	var records []record
	if err := json.Unmarshal(bytes, &records); err != nil {
		return
	}

	for _, r := range records {
		repo.data[r.ShortURL] = r.OriginalURL
	}
}

func (repo *MemoryRepository) saveToFile() error {
	records := make([]record, 0, len(repo.data))
	for short, original := range repo.data {
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

func (repo *MemoryRepository) Save(id string, url string) error {
	_, exists := repo.data[id]
	if exists {
		return ErrAlreadyExists
	}

	repo.data[id] = url
	repo.saveToFile()
	return nil
}

func (repo *MemoryRepository) Get(id string) (string, bool) {
	url, exists := repo.data[id]
	return url, exists
}
