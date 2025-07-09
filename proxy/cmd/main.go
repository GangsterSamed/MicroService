package main

import (
	"log/slog"
	"os"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	_ "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/docs"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/handlers"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/service"
)

// @title Proxy Service API
// @version 1.0
// @description API Gateway for microservices
// @host localhost:8080
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @tag.name auth
// @tag.description Authentication operations
// @tag.name user
// @tag.description User management operations
// @tag.name geo
// @tag.description Geocoding operations
func main() {
	// Загрузка конфигурации
	cfg, err := config.LoadProxyConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// 2. Инициализация логгера
	logger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "proxy",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(logger)

	// 3. Инициализация сервиса
	var proxyService handlers.ProxyServiceInterface
	proxyService, err = service.NewProxyService(
		cfg.GeoServiceAddr,
		cfg.AuthServiceAddr,
		cfg.UserServiceAddr,
		logger,
		cfg,
	)
	if err != nil {
		logger.Error("Failed to create proxy service", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := proxyService.Close(); err != nil {
			logger.Error("Failed to close proxy service", "error", err)
		}
	}()

	// 4. Инициализация HTTP сервера
	httpServer := startHTTPServer(cfg, proxyService, logger)
	if httpServer == nil {
		logger.Error("HTTP server initialization failed")
		os.Exit(1)
	}

	// Graceful shutdown
	waitForShutdown(httpServer, cfg.ShutdownTimeout, logger)
}
