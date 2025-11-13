package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
	"github.com/go-chi/chi/v5"
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
	resolveAddress := config.ResolveAddress + "/%s"
	fmt.Fprintf(w, resolveAddress, id)
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
	config.ParseFlags()
	router := chi.NewRouter()
	router.Get("/{id}", handleGet)
	router.Post("/", handlePost)

	server := &http.Server{
		Addr:         config.PortAddres,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
