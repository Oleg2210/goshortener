package repository

import "errors"

var ErrAlreadyExists = errors.New("id already exists")

type URLRepository interface {
	Save(id string, url string) error
	Get(id string) (string, bool)
}
