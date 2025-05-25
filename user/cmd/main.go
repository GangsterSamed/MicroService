package main

import (
	"log/slog"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/controller"

	_ "github.com/lib/pq"
	_ "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/docs"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/repository"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/service"
)

// @title User Service API
// @version 1.0
// @description Микросервис управления пользователями
// @contact.name Roman Malcev
// @contact.email romanmalcev89665@gmail.com
// @host localhost:8081
// @BasePath /api/user
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	// Инициализация конфигурации
	cfg, err := config.LoadUserConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Инициализация логгера
	makeLogger := logger2.SetupLogger(cfg.LogLevel, cfg.LogFormat).With(
		"service", "user",
		"version", "1.0",
		"env", cfg.Environment,
	)
	slog.SetDefault(makeLogger)

	// Подключение к DB
	db, err := initDatabase(cfg, makeLogger)
	if err != nil {
		makeLogger.Error("Database initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Инициализация слоев
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userCtrl, err := controller.NewUserController(userService)
	if err != nil {
		makeLogger.Error("Failed to create controller", "error", err)
		os.Exit(1)
	}

	userGRPCServer, err := service.NewUserServer(userService)
	if err != nil {
		makeLogger.Error("Failed to create user gRPC server", "error", err)
		os.Exit(1)
	}

	// Создание серверов
	grpcServer := createGRPCServer(userGRPCServer, makeLogger)
	httpServer := createHTTPServer(userCtrl, cfg, makeLogger)

	// Запуск серверов
	servers := startServers(grpcServer, httpServer, cfg, makeLogger)

	// Ожидание сигналов завершения
	waitForShutdown(servers, cfg.ShutdownTimeout, makeLogger)
}
