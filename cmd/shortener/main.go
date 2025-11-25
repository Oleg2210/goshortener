package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
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

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lrw := &loggingResponseWriter{
				ResponseWriter: w,
				status:         200,
			}

			var body []byte
			if r.Method == http.MethodPost {
				body, _ = io.ReadAll(r.Body)
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			next.ServeHTTP(lrw, r)

			logger.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", lrw.status),
				zap.Duration("duration", time.Since(start)),
				zap.ByteString("body", body),
				zap.String("remote_ip", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
		})
	}
}

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
	logger, _ := zap.NewProduction()

	router.Use(LoggingMiddleware(logger))
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
