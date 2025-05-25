package main

import (
	"log/slog"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/service"
)

func main() {
	// 1. Загрузка конфигурации
	cfg, err := config.LoadProxyConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Инициализация логгера
	makeLogger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "proxy",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(makeLogger)

	// 3. Инициализация сервисов
	proxyService, err := service.NewProxyService(
		cfg.GeoServiceAddr,
		cfg.AuthServiceAddr,
		cfg.UserServiceAddr,
		makeLogger,
	)
	if err != nil {
		makeLogger.Error("Failed to create proxy service", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := proxyService.Close(); err != nil {
			makeLogger.Error("Failed to close proxy service", "error", err)
		}
	}()

	// 4. Запуск серверов
	httpServer := startHTTPServer(cfg, proxyService, makeLogger)

	// 5. Graceful shutdown
	waitForShutdown(httpServer, cfg.ShutdownTimeout, makeLogger)
}
