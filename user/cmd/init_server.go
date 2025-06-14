package main

import (
	"database/sql"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/repository"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/internal/service"
	pb "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

// Создание grpc сервера
func setupGRPCServer(db *sql.DB, logger *slog.Logger) (*grpc.Server, error) {
	// Инициализация слоев
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userGRPCServer := service.NewUserServer(userService)

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger2.GRPCLoggerInterceptor(logger)),
	)
	pb.RegisterUserServiceServer(grpcServer, userGRPCServer)

	return grpcServer, nil
}

// Запуск grpc сервера
func startGRPCServer(server *grpc.Server, port int, logger *slog.Logger) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	go func() {
		logger.Info("Starting gRPC server", "port", port)
		if err := server.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	return listener, nil
}
