package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nikndip/projecthub-demo/internal/app"
	"github.com/nikndip/projecthub-demo/internal/httpapi"
	"github.com/nikndip/projecthub-demo/internal/store/postgres"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	databaseURL := getenv("DATABASE_URL", "postgres://projecthub:projecthub@localhost:5432/projecthub?sslmode=disable")
	store, err := postgres.Open(ctx, databaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}
	defer store.Close()
	if err := store.Migrate(ctx); err != nil {
		logger.Error("migrate database", "error", err)
		os.Exit(1)
	}

	service := app.NewService(store)
	handler := httpapi.NewHandler(service, store, logger)
	server := &http.Server{
		Addr:              getenv("APP_ADDR", ":8080"),
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server started", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("serve", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown", "error", err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
