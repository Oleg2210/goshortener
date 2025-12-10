package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Oleg2210/goshortener/internal/repository"
)

var ErrOutOfCombinations = errors.New("possible combinations are running out")

var ErrIDDoesNotExists = errors.New("such id does not exist")

type ShortenerService struct {
	repo      repository.URLRepository
	rnd       *rand.Rand
	letters   string
	minLength int
	maxLength int
}

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

func (service *ShortenerService) generateRandomID(letters string, size int) string {
	randomText := make([]byte, size)
	for i := range randomText {
		randomIndex := service.rnd.Intn(len(letters))
		randomText[i] = letters[randomIndex]
	}
	return string(randomText)
}

func (service *ShortenerService) Shorten(
	url string,
) (string, error) {
	for i := service.minLength; i < service.maxLength; i++ {
		id := service.generateRandomID(
			service.letters,
			i,
		)
		err := service.repo.Save(
			id,
			url,
		)

		if err == nil {
			return id, nil
		}
	}

	return "", ErrOutOfCombinations
}

func (service *ShortenerService) GetURL(id string) (string, error) {
	url, exists := service.repo.Get(id)
	if !exists {
		return "", ErrIDDoesNotExists
	}

	return url, nil
}

func (service *ShortenerService) Ping() bool {
	return service.repo.Ping()
}
