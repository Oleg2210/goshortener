package repository

import (
	"errors"

	"github.com/Oleg2210/goshortener/internal/entities"
)

var ErrAlreadyExists = errors.New("id already exists")

type URLRepository interface {
	Save(id string, url string) error
	BatchSave(records []entities.URLRecord) error
	Get(id string) (string, bool)
	Ping() bool
}
