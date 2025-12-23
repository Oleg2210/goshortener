package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/Oleg2210/goshortener/internal/entities"
	"github.com/Oleg2210/goshortener/internal/repository"
)

var ErrOutOfCombinations = errors.New("possible combinations are running out")

var ErrIDDoesNotExists = errors.New("such id does not exist")

var ErrURLExists = errors.New("such url already exists")

var ErrAnonymousUser = errors.New("anonymous user")

const anonymousUserID = "anonymous"

type ShortenerService struct {
	repo      repository.URLRepository
	rnd       *rand.Rand
	letters   string
	minLength int
	maxLength int
	userIDKey any
}

func NewShortenerService(
	repo repository.URLRepository,
	minLength int,
	maxLength int,
	userIDKey any,
) *ShortenerService {
	return &ShortenerService{
		repo:      repo,
		rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
		letters:   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		minLength: minLength,
		maxLength: maxLength,
		userIDKey: userIDKey,
	}
}

func (service *ShortenerService) generateRandomID(letters string, size int) string {
	randomText := make([]byte, size)
	for i := range randomText {
		randomIndex := service.rnd.Intn(len(letters))
		randomText[i] = letters[randomIndex]
	}
	return string(randomText)
}

func (service *ShortenerService) getUserID(ctx context.Context) string {
	userID, ok := ctx.Value(service.userIDKey).(string)

	if !ok {
		userID = anonymousUserID
	}

	return userID
}

func (service *ShortenerService) Shorten(
	ctx context.Context,
	url string,
) (string, error) {
	for i := service.minLength; i < service.maxLength; i++ {
		id := service.generateRandomID(
			service.letters,
			i,
		)

		userID := service.getUserID(ctx)

		short, err := service.repo.Save(
			ctx,
			id,
			url,
			userID,
		)

		if err == nil {
			if short != id {
				return short, ErrURLExists
			}
			return id, nil
		}
	}

	return "", ErrOutOfCombinations
}

func (service *ShortenerService) BatchShorten(
	ctx context.Context,
	records []entities.URLRecord,
) error {
	userID := service.getUserID(ctx)
	return service.repo.BatchSave(ctx, records, userID)
}

func (service *ShortenerService) GetURL(ctx context.Context, id string) (string, error) {
	url, exists := service.repo.Get(ctx, id)
	if !exists {
		return "", ErrIDDoesNotExists
	}

	return url, nil
}

func (service *ShortenerService) Ping(ctx context.Context) bool {
	return service.repo.Ping(ctx)
}

func (service *ShortenerService) GetUserShortens(ctx context.Context) ([]entities.URLRecord, error) {
	userID := service.getUserID(ctx)

	if userID == anonymousUserID {
		return []entities.URLRecord{}, ErrAnonymousUser
	}

	return service.repo.GetUserShortens(ctx, userID)
}
