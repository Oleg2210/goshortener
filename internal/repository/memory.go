package repository

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

func (repo *MemoryRepository) Get(id string) (string, bool) {
	url, exists := repo.data[id]
	return url, exists
}
