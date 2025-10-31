package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
)

var repo = repository.NewMemoryRepository()
var shortenerService = service.NewShortenerService(
	repo,
	config.Letters,
	config.MinLength,
	config.MaxLength,
)

func handlePost(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	url := string(body)

	id, err := shortenerService.Shorten(url)

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "http://localhost:8080/%s", id)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[1:]
	url, err := shortenerService.GetUrl(id)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			handlePost(w, r)
			return
		}
		if r.Method == http.MethodGet {
			handleGet(w, r)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
	})

	http.ListenAndServe(":8080", nil)
}
