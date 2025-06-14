package main

import (
	"context"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"log/slog"
	"net"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/shutdown"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализация конфигурации
	cfg, err := config.LoadUserConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Инициализация логгера
	logger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "user",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(logger)

	// Подключение к DB
	db, err := initDatabase(cfg, logger)
	if err != nil {
		logger.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error("Failed to close database connection", "error", err)
		}
	}()

	// Создание и запуск gRPC сервера
	grpcServer, err := setupGRPCServer(db, logger)
	if err != nil {
		return fmt.Errorf("failed to setup gRPC server: %w", err)
	}
	listener, err := startGRPCServer(grpcServer, cfg.GRPCPort, logger)
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error("Failed to close listener", "error", err)
		}
	}()

	// Ожидание сигналов завершения
	if err := shutdown.WaitForShutdown(ctx, grpcServer, listener, cfg.ShutdownTimeout, logger); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}

	return nil
}
