package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/service"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
)

func initUserClient(cfg *config.AuthConfig, logger *slog.Logger) (*grpc.ClientConn, error) {
	logger.Info("connecting to user service", "address", cfg.UserServiceAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: 30 * time.Second,
			Backoff: backoff.Config{
				BaseDelay:  2.0 * time.Second,
				Multiplier: 1.6,
				MaxDelay:   60 * time.Second,
			},
		}),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to user service: %w", err)
	}

	logger.Info("successfully connected to user service")
	return conn, nil
}

func setupGRPCServer(authService service.AuthService, logger *slog.Logger) (*grpc.Server, error) {
	// Создание gRPC сервера
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger2.GRPCLoggerInterceptor(logger)),
	)
	pb.RegisterAuthServiceServer(grpcServer, service.NewAuthServer(authService, logger))

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
