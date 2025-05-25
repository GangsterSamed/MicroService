package main

import (
	"log/slog"
	"os"
	_ "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/docs"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
)

// @title Geo Service API
// @version 1.0
// @description Сервис геокодирования и поиска адресов
// @host localhost:8080
// @BasePath /api/address
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	cfg, err := config.LoadGeoConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	makeLogger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "geo",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(makeLogger)

	// Инициализация зависимостей
	redisClient := initRedis(cfg, makeLogger)
	authClient := initAuthClient(cfg, makeLogger)
	geoProvider := initGeoProvider(cfg)

	// Запуск серверов
	httpServer, err := startHTTPServer(cfg, makeLogger)
	if err != nil {
		makeLogger.Error("Failed to start HTTP server", "error", err)
		os.Exit(1)
	}
	grpcServer := startGRPCServer(cfg, redisClient, authClient, geoProvider, makeLogger)

	// Ожидание сигналов завершения
	waitForShutdown(httpServer, grpcServer, cfg.ShutdownTimeout, makeLogger)
}
