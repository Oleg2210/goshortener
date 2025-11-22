package repository

type MemoryRepository struct {
	data map[string]string
}

func NewMemoryRepository(fileStoragePath string) *MemoryRepository {
	return &MemoryRepository{
		data: make(map[string]string),
	}
}

func (repo *MemoryRepository) Save(id string, url string) error {
	_, exists := repo.data[id]
	if exists {
		return ErrAlreadyExists
	}

	repo.data[id] = url
	return nil
}

func (repo *MemoryRepository) Get(id string) (string, bool) {
	url, exists := repo.data[id]
	return url, exists
}
