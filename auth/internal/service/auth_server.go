package service

import (
	"context"
	"log/slog"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
)

type AuthServer struct {
	proto.UnimplementedAuthServiceServer
	service AuthService
	logger  *slog.Logger
}

func NewAuthServer(authService AuthService, logger *slog.Logger) *AuthServer {
	return &AuthServer{
		service: authService,
		logger:  logger,
	}
}
func (s *AuthServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	logger := s.logger.With(
		"method", "Register",
		"email", req.Email,
	)
	logger.Info("Handling gRPC register request")

	resp, err := s.service.RegisterUser(ctx, req.Email, req.Password)
	if err != nil {
		logger.Error("Register failed", "error", err)
		return nil, err
	}

	logger.Info("Register successful", "user_id", resp.UserId)
	return resp, nil
}

func (s *AuthServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.AuthResponse, error) {
	logger := s.logger.With(
		"method", "Login",
		"email", req.Email,
	)
	logger.Info("Handling gRPC login request")

	resp, err := s.service.LoginUser(ctx, req.Email, req.Password)
	if err != nil {
		logger.Warn("Login failed", "error", err)
		return nil, err
	}

	logger.Info("Login successful", "user_id", resp.UserId)
	return resp, nil
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	logger := s.logger.With(
		"method", "ValidateToken",
	)
	logger.Debug("Validating token")

	resp, err := s.service.ValidateToken(ctx, req)
	if err != nil {
		logger.Warn("Token validation failed", "error", err)
		return &proto.TokenResponse{Valid: false}, nil
	}

	logger.Info("Token validated successfully", "user_id", resp.UserId)
	return resp, nil
}
