package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"net/http"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/controller"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/service"
)

// Структура для хранения серверов
type servers struct {
	grpc *grpc.Server
	http *http.Server
}

func createGRPCServer(userServer *service.UserServer, logger *slog.Logger) *grpc.Server {
	return grpc.NewServer(
		grpc.UnaryInterceptor(logger2.GRPCLoggerInterceptor(logger)),
	)
}

func createHTTPServer(userCtrl controller.UserController, cfg *config.UserConfig, logger *slog.Logger) *http.Server {
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
		user := api.Group("/user")
		user.Use(userCtrl.AuthMiddleware())
		{
			user.GET("/profile", userCtrl.GetUserProfileHandler)
			user.GET("/list", userCtrl.ListUsersHandler)
		}
	}

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
}

func startServers(grpcServer *grpc.Server, httpServer *http.Server, cfg *config.UserConfig, logger *slog.Logger) *servers {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Error("Failed to listen gRPC", "error", err)
		os.Exit(1)
	}

	go func() {
		logger.Info("Starting gRPC server", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
			os.Exit(1)
		}
	}()

	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	return &servers{
		grpc: grpcServer,
		http: httpServer,
	}
}
