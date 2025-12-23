package repository

import (
	"context"
	"errors"

	"github.com/Oleg2210/goshortener/internal/entities"
)

var ErrAlreadyExists = errors.New("id already exists")

type URLRepository interface {
	Save(ctx context.Context, id string, url string, userID string) (string, error)
	BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error
	Get(ctx context.Context, id string) (string, bool)
	Ping(ctx context.Context) bool
	GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error)
}
