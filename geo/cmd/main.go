package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/shutdown"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/provider"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("geo service failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Инициализация конфигурации
	cfg, err := config.LoadGeoConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Инициализация логгера
	logger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "geo",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(logger)

	// 3. Инициализация Redis
	redisClient := initRedis(cfg, logger)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("failed to close redis", "error", err)
		}
	}()

	// 4. Инициализация провайдера геоданных
	geoProvider := provider.NewGeoProviderAdapter(
		provider.NewGeoService(cfg.DadataAPIKey, cfg.DadataSecretKey),
	)

	// 5. Настройка и запуск gRPC сервера
	grpcServer, err := setupGRPCServer(geoProvider, redisClient, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to setup gRPC server: %w", err)
	}
	listener, err := startGRPCServer(ctx, grpcServer, cfg.GRPCPort, logger)
	if err != nil {
		return fmt.Errorf("failed to start gRPC server: %w", err)
	}
	defer func() {
		if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
			logger.Error("Failed to close listener", "error", err)
		}
	}()

	// 6. Ожидание сигналов завершения
	if err := shutdown.WaitForShutdown(ctx, grpcServer, listener, cfg.ShutdownTimeout, logger); err != nil {
		return fmt.Errorf("shutdown failed: %w", err)
	}

	return nil
}
