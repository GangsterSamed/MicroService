package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/errors"
	userProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

// UserHandler обрабатывает запросы к user сервису
type UserHandler struct {
	UserClient userProto.UserServiceClient
	logger     *slog.Logger
}

// NewUserHandler создает новый обработчик user запросов
func NewUserHandler(userClient userProto.UserServiceClient, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		UserClient: userClient,
		logger:     logger,
	}
}

// HandleUserRequest обрабатывает запросы к user сервису
func (h *UserHandler) HandleUserRequest(ctx context.Context, method string, body []byte) ([]byte, int, error) {
	h.logger.Info("Starting handleUserRequest", "method", method)

	switch method {
	case "/api/user/profile":
		h.logger.Info("Getting user profile")
		// Получаем user_id из метаданных
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "no metadata in context")
		}

		userIDs := md.Get("user_id")
		if len(userIDs) == 0 {
			return nil, http.StatusUnauthorized, status.Error(codes.Unauthenticated, "user_id is not provided")
		}

		resp, err := h.UserClient.GetUserProfile(ctx, &userProto.GetUserRequest{
			UserId: userIDs[0],
		})
		return errors.MarshalResponse(resp, err)

	case "/api/user/list":
		h.logger.Info("Getting user list")
		// Получаем параметры пагинации из метаданных
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "no metadata in context")
		}

		// Получаем limit и offset из query параметров
		limitStr := md.Get("limit")
		offsetStr := md.Get("offset")

		// Передаем limit/offset как есть (валидированы на уровне Gin)
		resp, err := h.UserClient.ListUsers(ctx, &userProto.ListUsersRequest{
			Limit:  parseInt32(limitStr, 10),
			Offset: parseInt32(offsetStr, 0),
		})
		return errors.MarshalResponse(resp, err)

	default:
		h.logger.Warn("Method not implemented", "method", method)
		return nil, http.StatusNotImplemented, status.Error(codes.Unimplemented, "method not implemented")
	}
}

// parseInt32 преобразует строку в int32 с дефолтом
func parseInt32(strs []string, def int32) int32 {
	if len(strs) > 0 {
		if v, err := strconv.ParseInt(strs[0], 10, 32); err == nil {
			return int32(v)
		}
	}
	return def
}
