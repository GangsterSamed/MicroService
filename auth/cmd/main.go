package main

import (
	"log/slog"
	"os"
	_ "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/docs"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/service"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	pbUser "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

// @title Auth Service API
// @version 1.0
// @description API for user authentication and authorization
// @host localhost:8080
// @BasePath /api/auth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.LoadAuthConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Инициализация логгера
	makeLogger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "auth",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(makeLogger)

	// 3. Инициализация соединений
	userConn, err := initUserClient(cfg, makeLogger)
	if err != nil {
		slog.Error("Failed to initialize user client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := userConn.Close(); err != nil {
			makeLogger.Error("Failed to close user connection", "error", err)
		}
	}()

	// 4. Создание сервисов
	authService := service.NewAuthService(
		pbUser.NewUserServiceClient(userConn),
		cfg.JWTSecret,
		makeLogger,
	)

	// 5. Запуск серверов
	httpServer := startHTTPServer(authService, cfg, makeLogger)
	grpcServer := startGRPCServer(authService, cfg, makeLogger)

	// 6. Graceful shutdown
	waitForShutdown(httpServer, grpcServer, cfg.ShutdownTimeout, makeLogger)
}
