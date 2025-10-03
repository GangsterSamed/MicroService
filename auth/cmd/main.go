package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/shutdown"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/service"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	pbUser "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

func main() {
	if err := run(); err != nil {
		slog.Error("auth service failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Загрузка конфигурации
	cfg, err := config.LoadAuthConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// 2. Инициализация логгера
	logger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "auth",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(logger)

	// 3. Инициализация клиента user-сервиса
	userConn, err := initUserClient(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to init user client: %w", err)
	}
	defer func() {
		if err := userConn.Close(); err != nil {
			logger.Error("Failed to close user connection", "error", err)
		}
	}()

	// 4. Создание сервисов
	authService := service.NewAuthService(
		pbUser.NewUserServiceClient(userConn),
		cfg.JWTSecret,
		logger,
	)

	// 5. Запуск серверов
	grpcServer, err := setupGRPCServer(authService, logger)
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
