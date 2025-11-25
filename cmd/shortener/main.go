package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/serializers"
	"github.com/Oleg2210/goshortener/internal/service"
	compres "github.com/Oleg2210/goshortener/pkg/middleware/compress"
	"github.com/Oleg2210/goshortener/pkg/middleware/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
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

func handlePostJson(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req serializers.Request

	if err := req.UnmarshalJSON(body); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	id, err := shortenerService.Shorten(req.URL)

	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	resp := serializers.Response{
		Result: config.ResolveAddress + "/" + id,
	}
	jsonBytes, _ := resp.MarshalJSON()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonBytes)
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
	logger, _ := zap.NewProduction()

	router.Use(logging.LoggingMiddleware(logger))
	router.Use(compres.GzipMiddleware)
	router.Get("/{id}", handleGet)
	router.Post("/", handlePost)
	router.Post("/api/shorten", handlePostJson)

	server := &http.Server{
		Addr:         config.PortAddres,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.ListenAndServe()
}
