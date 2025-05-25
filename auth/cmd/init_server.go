package main

import (
	"context"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/controller"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/service"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"time"
)

func initUserClient(cfg *config.AuthConfig, logger *slog.Logger) (*grpc.ClientConn, error) {
	logger.Info("Connecting to user service", "address", cfg.UserServiceAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 10 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  1.0 * time.Second,
				Multiplier: 1.6,
				MaxDelay:   15 * time.Second,
			},
		}),
		grpc.WithBlock(),
	)

	if err != nil {
		return nil, err
	}

	logger.Info("Successfully connected to user service")
	return conn, nil
}

func startHTTPServer(authService service.AuthService, cfg *config.AuthConfig, logger *slog.Logger) *http.Server {
	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(logger2.GinLoggerMiddleware(logger))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Routes
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			ctrl := controller.NewAuthController(authService, logger)
			auth.POST("/register", ctrl.RegisterHandler)
			auth.POST("/login", ctrl.LoginHandler)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		status := http.StatusOK
		// Проверка соединения с БД и другими сервисами
		c.JSON(status, gin.H{
			"status":  "OK",
			"details": map[string]interface{}{},
		})
	})

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(cfg.HTTPPort),
		Handler: router,
	}

	go func() {
		slog.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
		}
	}()

	return srv
}

func startGRPCServer(authService service.AuthService, cfg *config.AuthConfig, logger *slog.Logger) *grpc.Server {
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(cfg.GRPCPort))
	if err != nil {
		slog.Error("Failed to listen GRPC", "port", cfg.GRPCPort, "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(logger2.GRPCLoggerInterceptor(logger)),
	)
	proto.RegisterAuthServiceServer(srv, service.NewAuthServer(authService, logger))

	go func() {
		slog.Info("Starting GRPC server", "port", cfg.GRPCPort)
		if err := srv.Serve(lis); err != nil {
			slog.Error("GRPC server failed", "error", err)
		}
	}()

	return srv
}
