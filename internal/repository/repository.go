package repository

import (
	"context"
	"errors"

	"github.com/Oleg2210/goshortener/internal/entities"
)

// ErrAlreadyExists is returned when attempting to save a URL with an ID that already exists.
var ErrAlreadyExists = errors.New("id already exists")

// URLRepository defines the interface for storing and retrieving URL records.
type URLRepository interface {
	// Save stores a single URL record with a given ID and user.
	// Returns the ID used for storage or an error if it already exists.
	Save(ctx context.Context, id string, url string, userID string, isDeleted bool) (string, error)

	// BatchSave stores multiple URL records in a single operation.
	// Returns an error if any of the records already exist.
	BatchSave(ctx context.Context, records []entities.URLRecord, userID string) error

	// Get retrieves a URL record by its short ID.
	// Returns the record and true if it exists, or an empty record and false if not found.
	Get(ctx context.Context, id string) (entities.URLRecord, bool)

	// Ping checks the health or connectivity of the repository.
	// Returns true if the repository is reachable, false otherwise.
	Ping(ctx context.Context) bool

	// GetUserShortens returns all non-deleted URL records associated with a given user.
	GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error)

	// MarkDelete marks a list of short URLs as deleted for a specific user.
	MarkDelete(ctx context.Context, short []string, userID string) error
}
