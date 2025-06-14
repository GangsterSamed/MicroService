package main

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/provider"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/service"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
)

func setupGRPCServer(provider provider.GeoProvider, redis *redis.Client, cfg *config.GeoConfig, logger *slog.Logger) (*grpc.Server, error) {
	// Проверка соединения с Redis
	if _, err := redis.Ping(context.Background()).Result(); err != nil {
		logger.Error("Redis connection check failed", "error", err)
		os.Exit(1)
	}

	// Инициализация слоев
	geoService := service.NewGeoGRPCServer(provider, redis, cfg, logger)

	// Создание gRPC сервера
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger2.GRPCLoggerInterceptor(logger)),
	)
	pb.RegisterGeoServiceServer(grpcServer, geoService)

	return grpcServer, nil
}

func startGRPCServer(ctx context.Context, server *grpc.Server, port int, logger *slog.Logger) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	go func() {
		logger.Info("starting gRPC server", "port", port)
		if err := server.Serve(listener); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	return listener, nil
}
