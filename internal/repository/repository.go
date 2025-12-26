package repository

import (
	"context"
	"errors"

	"github.com/Oleg2210/goshortener/internal/entities"
)

var ErrAlreadyExists = errors.New("id already exists")

type URLRepository interface {
	Save(ctx context.Context, id string, url string, userID string, isDeleted bool) (string, error)
	BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error
	Get(ctx context.Context, id string) (entities.URLRecord, bool)
	Ping(ctx context.Context) bool
	GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error)
	MarkDelete(ctx context.Context, short []string, userID string) error
}
