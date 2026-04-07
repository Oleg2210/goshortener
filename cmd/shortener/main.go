package main

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Oleg2210/goshortener/internal/config"
	"github.com/Oleg2210/goshortener/internal/handler"
	"github.com/Oleg2210/goshortener/internal/repository"
	"github.com/Oleg2210/goshortener/internal/service"
	compres "github.com/Oleg2210/goshortener/pkg/middleware/compress"
	"github.com/Oleg2210/goshortener/pkg/middleware/cookies"
	"github.com/Oleg2210/goshortener/pkg/middleware/logging"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func printBuildVars() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func chooseStorage(ctx context.Context, logger *zap.Logger) repository.URLRepository {
	if config.DatabaseInfo != "" {
		repo, err := repository.NewDBRepository(config.DatabaseInfo)

		if err == nil {
			return repo
		}

		logger.Error("failed to create db repo", zap.Error(err))
	}

	if config.FileStoragePath != "" {
		repo, err := repository.NewFileRepository(ctx, config.FileStoragePath)

		if err == nil {
			return repo
		}

		logger.Error("failed to create file repo", zap.Error(err))
	}

	return repository.NewMemoryRepository()
}

func makePublisher(logger *zap.Logger) (*handler.AuditPublisher, error) {
	publisher := handler.NewAuditPublisher(100, 100, logger)

	if config.AuditFile != "" {
		fileObs, err := handler.NewFileObserver(config.AuditFile, logger)
		if err != nil {
			logger.Error("failed to create file observer", zap.Error(err))
			return nil, err
		}
		publisher.Register(fileObs)
	}

	if config.AuditURL != "" {
		httpObs := handler.NewHTTPObserver(config.AuditURL, logger)
		publisher.Register(httpObs)
	}

	return publisher, nil
}

func main() {
	printBuildVars()

	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse config: %v\n", err)
		os.Exit(1)
	}

	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init zap logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer stop()

	var wg sync.WaitGroup

	repo := chooseStorage(ctx, logger)

	shortenerService := service.NewShortenerService(
		repo,
		config.MinLength,
		config.MaxLength,
	)

	deleter := handler.NewDeleter(ctx, &wg, logger, shortenerService, 1)

	publisher, err := makePublisher(logger)
	if err != nil {
		logger.Fatal("failed to init publisher", zap.Error(err))
	}

	app := handler.App{
		ShortenerService: shortenerService,
		Logger:           logger,
		Deleter:          deleter,
		Publisher:        publisher,
	}

	router := chi.NewRouter()
	router.Use(logging.LoggingMiddleware(logger))
	router.Use(cookies.AuthMiddleware([]byte(config.AuthSecret)))
	router.Use(compres.GzipMiddleware)

	router.Get("/{id}", app.HandleGet)
	router.Post("/", app.HandlePost)
	router.Post("/api/shorten", app.HandlePostJSON)
	router.Post("/api/shorten/batch", app.HandlePostBatchJSON)
	router.Get("/ping", app.HandlePing)
	router.Get("/api/user/urls", app.HandleGetAllUserUrls)
	router.Delete("/api/user/urls", app.HandleMarkDelete)
	router.Get("/api/internal/stats", app.HandleGetStatistic)

	mainServer := &http.Server{
		Addr:         config.PortAddres,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 45 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	pprofServer := &http.Server{
		Addr:    "localhost:6060",
		Handler: nil,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("main server started", zap.String("addr", config.PortAddres))

		var err error
		if config.EnableHTTPS {
			err = mainServer.ListenAndServeTLS(config.CertHTTPS, config.KeyHTTPS)
		} else {
			err = mainServer.ListenAndServe()
		}

		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("main server error", zap.Error(err))
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("pprof started", zap.String("addr", "localhost:6060"))

		if err := pprofServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("pprof server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := mainServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("main server shutdown error", zap.Error(err))
	}

	if err := pprofServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("pprof shutdown error", zap.Error(err))
	}

	wg.Wait()
	logger.Info("servers stopped")
}
