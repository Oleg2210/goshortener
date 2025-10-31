package service

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Oleg2210/goshortener/internal/repository"
)

var ErrOutOfCombinations = errors.New("possible combinations are running out")

var ErrIdDoesNotExists = errors.New("such id does not exist")

type ShortenerService struct {
	repo      repository.URLRepository
	rnd       *rand.Rand
	letters   string
	minLength int
	maxLength int
}

func NewShortenerService(
	repo repository.URLRepository,
	letters string,
	minLength int,
	maxLength int,
) *ShortenerService {
	return &ShortenerService{
		repo:      repo,
		rnd:       rand.New(rand.NewSource(time.Now().UnixNano())),
		letters:   letters,
		minLength: minLength,
		maxLength: maxLength,
	}
}

func (service *ShortenerService) generateRandomID(letters string, size int) string {
	random_text := make([]byte, size)
	for i := range random_text {
		random_index := service.rnd.Intn(len(letters))
		random_text[i] = letters[random_index]
	}
	return string(random_text)
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

func (service *ShortenerService) GetUrl(id string) (string, error) {
	url, exists := service.repo.Get(id)
	if !exists {
		return "", ErrIdDoesNotExists
	}

	return url, nil
}
