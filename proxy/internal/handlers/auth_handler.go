package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	authProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
)

// AuthHandler обрабатывает запросы к auth сервису
type AuthHandler struct {
	AuthClient authProto.AuthServiceClient
	logger     *slog.Logger
}

// NewAuthHandler создает новый обработчик auth запросов
func NewAuthHandler(authClient authProto.AuthServiceClient, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{
		AuthClient: authClient,
		logger:     logger,
	}
}

// HandleAuthRequest обрабатывает запросы к auth сервису
func (h *AuthHandler) HandleAuthRequest(ctx context.Context, method string, body []byte) ([]byte, int, error) {
	switch method {
	case "/api/auth/register":
		var req authProto.RegisterRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := h.AuthClient.Register(ctx, &req)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		// Сериализуем ответ в JSON
		jsonResp, err := protojson.Marshal(resp)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to marshal response: %w", err)
		}
		return jsonResp, http.StatusCreated, nil

	case "/api/auth/login":
		var req authProto.LoginRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := h.AuthClient.Login(ctx, &req)
		if err != nil {
			return nil, http.StatusInternalServerError, err
		}
		// Сериализуем ответ в JSON
		jsonResp, err := protojson.Marshal(resp)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("failed to marshal response: %w", err)
		}
		return jsonResp, http.StatusOK, nil

	default:
		return nil, http.StatusNotImplemented, status.Error(codes.Unimplemented, "method not implemented")
	}
}

// ValidateToken валидирует JWT токен
func (h *AuthHandler) ValidateToken(ctx context.Context, token string) (*authProto.TokenResponse, error) {
	return h.AuthClient.ValidateToken(ctx, &authProto.TokenRequest{Token: token})
}

// ExtractJWTToken извлекает JWT токен из строки, обрабатывая base64 и JSON форматы
func (h *AuthHandler) ExtractJWTToken(token string) (string, error) {
	h.logger.Info("Extracting JWT token", "input", token)

	// Убираем префикс Bearer если он есть
	token = strings.TrimPrefix(token, "Bearer ")
	h.logger.Info("Token after removing Bearer prefix", "token", token)

	// Пытаемся декодировать base64
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		h.logger.Info("Token is not base64, using as is", "token", token)
		return token, nil
	}
	h.logger.Info("Successfully decoded base64", "decoded", string(decoded))

	// Пытаемся распарсить JSON
	var tokenData struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(decoded, &tokenData); err != nil {
		h.logger.Error("Failed to parse token JSON", "error", err)
		return "", fmt.Errorf("invalid token format: %w", err)
	}
	h.logger.Info("Successfully parsed JSON", "tokenData", tokenData)

	// Если нашли токен в JSON, возвращаем его
	if tokenData.Token != "" {
		h.logger.Info("Found token in JSON, returning it", "token", tokenData.Token)
		return tokenData.Token, nil
	}

	// Если токен не найден в JSON, возвращаем ошибку
	h.logger.Error("No token found in JSON")
	return "", fmt.Errorf("no token found in JSON")
}
