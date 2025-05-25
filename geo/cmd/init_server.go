package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	proto2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/client"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/controller"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/service"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/provider"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"time"
)

func initAuthClient(cfg *config.GeoConfig, logger *slog.Logger) proto2.AuthServiceClient {
	conn, err := grpc.Dial(
		cfg.AuthServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Error("Failed to connect to auth service", "error", err)
		os.Exit(1)
	}
	return proto2.NewAuthServiceClient(conn)
}

func initGeoGRPCClient(cfg *config.GeoConfig, logger *slog.Logger) (*client.GRPCClient, error) {
	geoClient, err := client.NewGRPCClient("geo:"+strconv.Itoa(cfg.GRPCPort), 10*time.Second)
	if err != nil {
		logger.Error("Failed to create geo client", "error", err)
		return nil, fmt.Errorf("geo client init failed: %w", err)
	}
	return geoClient, nil
}

func initGeoProvider(cfg *config.GeoConfig) provider.GeoProvider {
	return provider.NewGeoProviderAdapter(
		provider.NewGeoService(
			cfg.DadataAPIKey,
			cfg.DadataSecretKey,
		),
	)
}

func startHTTPServer(cfg *config.GeoConfig, logger *slog.Logger) (*http.Server, error) {
	// Инициализация контроллера
	geoClient, err := initGeoGRPCClient(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init geo client: %w", err)
	}

	authService, err := service.NewAuthClient(cfg.AuthServiceAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to init auth client: %w", err)
	}
	addressCtrl := controller.NewAddressController(geoClient, authService)

	// Настройка маршрутов
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger2.GinLoggerMiddleware(logger))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

	// API routes
	api := router.Group("/api")
	{
		address := api.Group("/address")
		{
			address.POST("/search", addressCtrl.AddressSearchHandler)
			address.POST("/geocode", addressCtrl.AddressGeocodeHandler)
		}
	}

	// Запуск сервера
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.HTTPPort),
		Handler: router,
	}

	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	return srv, nil
}

func startGRPCServer(cfg *config.GeoConfig, redisClient *redis.Client, authClient proto2.AuthServiceClient, geoProvider provider.GeoProvider, logger *slog.Logger) *grpc.Server {
	// Создание gRPC сервера
	server := grpc.NewServer(
		grpc.UnaryInterceptor(logger2.GRPCLoggerInterceptor(logger)),
	)

	// Регистрация сервиса
	geoServer := &GeoGRPCServer{
		provider:    geoProvider,
		authClient:  authClient,
		redisClient: redisClient,
		logger:      logger,
	}
	proto.RegisterGeoServiceServer(server, geoServer)

	// Запуск сервера
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		logger.Error("Failed to listen gRPC", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("Starting gRPC server", "port", cfg.GRPCPort)
		if err := server.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	return server
}
