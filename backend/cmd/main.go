package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"backend/config"
	"backend/internal/app"
	"backend/internal/db"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("configuration loading error", zap.Error(err))
	}

	database, err := db.Connect(cfg)
	if err != nil {
		logger.Fatal("error connecting to postgresql", zap.Error(err))
	}
	defer database.Close()

	server := app.NewServer(cfg, logger, database)

	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal("server startup error", zap.Error(err))
		}
	}()

	logger.Info("server is running", zap.String("port", cfg.ServerPort))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("termination signal received, stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("error while stopping server", zap.Error(err))
	}
}
