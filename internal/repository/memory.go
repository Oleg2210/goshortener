package repository

import "github.com/Oleg2210/goshortener/internal/entities"

type MemoryRepository struct {
	data map[string]string
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		data: make(map[string]string),
	}

	return repo
}

func (repo *MemoryRepository) Save(id string, url string) error {
	_, exists := repo.data[id]
	if exists {
		return ErrAlreadyExists
	}

	repo.data[id] = url
	return nil
}

func (repo *MemoryRepository) BatchSave(
	records []entities.URLRecord,
) error {
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

func (repo *MemoryRepository) Get(id string) (string, bool) {
	url, exists := repo.data[id]
	return url, exists
}

func (repo *MemoryRepository) Ping() bool {
	return false
}
