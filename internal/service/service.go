package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/Oleg2210/goshortener/internal/entities"
	"github.com/Oleg2210/goshortener/internal/repository"
)

// ErrOutOfCombinations is returned when all possible short ID combinations are exhausted.
var ErrOutOfCombinations = errors.New("possible combinations are running out")

// ErrIDDoesNotExists is returned when the requested short URL ID does not exist.
var ErrIDDoesNotExists = errors.New("such id does not exist")

// ErrURLExists is returned when the URL already exists in the repository.
var ErrURLExists = errors.New("such url already exists")

// ShortenerService provides methods to shorten URLs and manage short URL records.
type ShortenerService struct {
	repo      repository.URLRepository
	rnd       *rand.Rand
	letters   string
	minLength int
	maxLength int
}

// NewShortenerService creates a new ShortenerService with a given repository and ID length constraints.
func NewShortenerService(
	repo repository.URLRepository,
	minLength int,
	maxLength int,
) *ShortenerService {
	return &ShortenerService{
		repo:      repo,
		rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
		letters:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		minLength: minLength,
		maxLength: maxLength,
	}
}

// generateRandomID creates a random string of given size using the allowed letters.
func (service *ShortenerService) generateRandomID(letters string, size int) string {
	randomText := make([]byte, size)
	for i := range randomText {
		randomIndex := service.rnd.Intn(len(letters))
		randomText[i] = letters[randomIndex]
	}
	return string(randomText)
}

// Shorten generates a short ID for the given URL and stores it in the repository.
// Returns the short ID or an error if unable to generate a unique ID.
func (service *ShortenerService) Shorten(
	ctx context.Context,
	url string,
	userID string,
) (string, error) {
	for i := service.minLength; i < service.maxLength; i++ {
		id := service.generateRandomID(service.letters, i)

		short, err := service.repo.Save(ctx, id, url, userID, false)
		if err == nil {
			if short != id {
				return short, ErrURLExists
			}
			return id, nil
		}
	}

	return "", ErrOutOfCombinations
}

// BatchShorten stores multiple URL records in a single operation.
func (service *ShortenerService) BatchShorten(
	ctx context.Context,
	records []entities.URLRecord,
	userID string,
) error {
	return service.repo.BatchSave(ctx, records, userID)
}

// GetURL retrieves a URL record by its short ID.
// Returns ErrIDDoesNotExists if the ID is not found.
func (service *ShortenerService) GetURL(ctx context.Context, id string) (entities.URLRecord, error) {
	url, exists := service.repo.Get(ctx, id)
	if !exists {
		return entities.URLRecord{}, ErrIDDoesNotExists
	}
	return url, nil
}

// Ping checks the health or connectivity of the underlying repository.
func (service *ShortenerService) Ping(ctx context.Context) bool {
	return service.repo.Ping(ctx)
}

// GetUserShortens returns all non-deleted URL records for a specific user.
func (service *ShortenerService) GetUserShortens(ctx context.Context, userID string) ([]entities.URLRecord, error) {
	return service.repo.GetUserShortens(ctx, userID)
}

// MarkDelete marks the specified short IDs as deleted for a given user.
func (service *ShortenerService) MarkDelete(ctx context.Context, shorts []string, userID string) error {
	return service.repo.MarkDelete(ctx, shorts, userID)
}

// GetInternalStatistic returns all users and url counts
func (service *ShortenerService) GetInternalStatistic(ctx context.Context) (int, int, error) {
	return service.repo.GetStatistic(ctx)
}
