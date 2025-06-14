package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"

	authProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	geoProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/errors"
	userProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

type ProxyService interface {
	ForwardRequest(ctx context.Context, serviceName, path string, body []byte, md metadata.MD) ([]byte, int, error)
	Close() error
}

type proxyService struct {
	geoClient  geoProto.GeoServiceClient
	authClient authProto.AuthServiceClient
	userClient userProto.UserServiceClient
	logger     *slog.Logger
	geoConn    *grpc.ClientConn
	authConn   *grpc.ClientConn
	userConn   *grpc.ClientConn
}

// NewProxyService создает новый экземпляр прокси-сервиса
func NewProxyService(geoAddr string, authAddr string, userAddr string, logger *slog.Logger) (ProxyService, error) {
	// Создаем соединения с сервисами
	geoConn, err := grpc.Dial(geoAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to geo service: %v", err)
	}

	authConn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	if err != nil {
		geoConn.Close()
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	userConn, err := grpc.Dial(userAddr, grpc.WithInsecure())
	if err != nil {
		geoConn.Close()
		authConn.Close()
		return nil, fmt.Errorf("failed to connect to user service: %v", err)
	}

	// Создаем клиентов
	geoClient := geoProto.NewGeoServiceClient(geoConn)
	authClient := authProto.NewAuthServiceClient(authConn)
	userClient := userProto.NewUserServiceClient(userConn)

	return &proxyService{
		geoClient:  geoClient,
		authClient: authClient,
		userClient: userClient,
		logger:     logger,
		geoConn:    geoConn,
		authConn:   authConn,
		userConn:   userConn,
	}, nil
}

func (s *proxyService) ForwardRequest(ctx context.Context, serviceName string, method string, reqBody []byte, headers metadata.MD) ([]byte, int, error) {
	// Добавляем заголовки в контекст всегда
	if len(headers) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, headers)
	}

	// Для всех сервисов, кроме auth, валидируем токен и подготавливаем контекст
	if serviceName != "auth" {
		var err error
		ctx, err = s.validateAndPrepareContext(ctx)
		if err != nil {
			return nil, http.StatusUnauthorized, err
		}
	}

	switch serviceName {
	case "geo":
		return s.handleGeoRequest(ctx, method, reqBody)
	case "auth":
		return s.handleAuthRequest(ctx, method, reqBody)
	case "user":
		return s.handleUserRequest(ctx, method, reqBody)
	default:
		return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid service name")
	}
}

// extractJWTToken извлекает JWT токен из строки, обрабатывая base64 и JSON форматы
func (s *proxyService) extractJWTToken(token string) (string, error) {
	s.logger.Info("Extracting JWT token", "input", token)

	// Убираем префикс Bearer если он есть
	token = strings.TrimPrefix(token, "Bearer ")
	s.logger.Info("Token after removing Bearer prefix", "token", token)

	// Пытаемся декодировать base64
	decoded, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		s.logger.Info("Token is not base64, using as is", "token", token)
		return token, nil
	}
	s.logger.Info("Successfully decoded base64", "decoded", string(decoded))

	// Пытаемся распарсить JSON
	var tokenData struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(decoded, &tokenData); err != nil {
		s.logger.Error("Failed to parse token JSON", "error", err)
		return "", fmt.Errorf("invalid token format: %w", err)
	}
	s.logger.Info("Successfully parsed JSON", "tokenData", tokenData)

	// Если нашли токен в JSON, возвращаем его
	if tokenData.Token != "" {
		s.logger.Info("Found token in JSON, returning it", "token", tokenData.Token)
		return tokenData.Token, nil
	}

	// Если токен не найден в JSON, возвращаем ошибку
	s.logger.Error("No token found in JSON")
	return "", fmt.Errorf("no token found in JSON")
}

// validateAndPrepareContext validates the token and prepares the context with proper metadata
func (s *proxyService) validateAndPrepareContext(ctx context.Context) (context.Context, error) {
	s.logger.Info("Starting token validation")

	// Получаем токен из заголовков
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		s.logger.Warn("No metadata in outgoing context")
		return nil, status.Error(codes.PermissionDenied, "no token provided")
	}

	s.logger.Info("Received metadata in context",
		slog.Any("metadata", md),
	)

	tokens := md.Get("authorization")
	s.logger.Info("Authorization tokens from metadata",
		slog.Any("tokens", tokens),
	)
	if len(tokens) == 0 {
		s.logger.Warn("No authorization token in metadata")
		return nil, status.Error(codes.PermissionDenied, "no token provided")
	}

	// Извлекаем JWT токен
	jwtToken, err := s.extractJWTToken(tokens[0])
	if err != nil {
		s.logger.Error("Failed to extract JWT token",
			slog.String("error", err.Error()),
			slog.String("token", tokens[0]),
		)
		return nil, status.Error(codes.Unauthenticated, "invalid token format")
	}

	// Валидируем токен
	s.logger.Info("Validating token", "token", jwtToken)
	resp, err := s.authClient.ValidateToken(ctx, &authProto.TokenRequest{Token: jwtToken})
	if err != nil {
		s.logger.Error("ValidateToken error", "err", err)
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	if !resp.Valid {
		s.logger.Warn("Token is not valid")
		return nil, status.Error(codes.Unauthenticated, "invalid token")
	}
	s.logger.Info("Token validated successfully", "user_id", resp.UserId)

	// Создаем новые метаданные
	newMD := metadata.New(map[string]string{
		"user_id":       resp.UserId,
		"authorization": "Bearer " + jwtToken,
	})

	// Объединяем существующие метаданные с новыми
	if existingMD, ok := metadata.FromOutgoingContext(ctx); ok {
		s.logger.Info("Found existing outgoing metadata, merging with new metadata",
			slog.Any("existing_metadata", existingMD),
		)
		newMD = metadata.Join(existingMD, newMD)
	}

	s.logger.Info("Created new context with metadata",
		slog.Any("new_metadata", newMD),
	)

	return metadata.NewOutgoingContext(ctx, newMD), nil
}

func (s *proxyService) handleGeoRequest(ctx context.Context, method string, body []byte) ([]byte, int, error) {
	s.logger.Info("Starting handleGeoRequest", "method", method)

	switch method {
	case "/api/address/search":
		s.logger.Info("Searching address")
		var req geoProto.SearchRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.geoClient.AddressSearch(ctx, &req)
		return errors.MarshalResponse(resp, err)

	case "/api/address/geocode":
		s.logger.Info("Geocoding address")
		var req geoProto.GeoRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.geoClient.GeoCode(ctx, &req)
		return errors.MarshalResponse(resp, err)

	default:
		s.logger.Warn("Method not implemented", "method", method)
		return nil, http.StatusNotImplemented, status.Error(codes.Unimplemented, "method not implemented")
	}
}

func (s *proxyService) handleAuthRequest(ctx context.Context, method string, body []byte) ([]byte, int, error) {
	switch method {
	case "/api/auth/register":
		var req authProto.RegisterRequest
		if err := protojson.Unmarshal(body, &req); err != nil {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid request body")
		}
		resp, err := s.authClient.Register(ctx, &req)
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
		resp, err := s.authClient.Login(ctx, &req)
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

func (s *proxyService) handleUserRequest(ctx context.Context, method string, body []byte) ([]byte, int, error) {
	s.logger.Info("Starting handleUserRequest", "method", method)

	switch method {
	case "/api/user/profile":
		s.logger.Info("Getting user profile")
		resp, err := s.userClient.GetUserProfile(ctx, &userProto.GetUserRequest{})
		return errors.MarshalResponse(resp, err)

	case "/api/user/list":
		s.logger.Info("Getting user list")
		// Получаем параметры пагинации из метаданных
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "no metadata in context")
		}

		// Получаем limit и offset из query параметров
		limitStr := md.Get("limit")
		offsetStr := md.Get("offset")

		limit := int32(10) // значение по умолчанию
		offset := int32(0) // значение по умолчанию

		if len(limitStr) > 0 {
			if l, err := strconv.ParseInt(limitStr[0], 10, 32); err == nil {
				limit = int32(l)
			}
		}
		if len(offsetStr) > 0 {
			if o, err := strconv.ParseInt(offsetStr[0], 10, 32); err == nil {
				offset = int32(o)
			}
		}

		s.logger.Info("Parsed pagination parameters",
			slog.Int("limit", int(limit)),
			slog.Int("offset", int(offset)),
		)

		resp, err := s.userClient.ListUsers(ctx, &userProto.ListUsersRequest{
			Limit:  limit,
			Offset: offset,
		})
		return errors.MarshalResponse(resp, err)

	default:
		s.logger.Warn("Method not implemented", "method", method)
		return nil, http.StatusNotImplemented, status.Error(codes.Unimplemented, "method not implemented")
	}
}

// Close закрывает все соединения
func (s *proxyService) Close() error {
	var errs []error

	// Закрываем соединения с сервисами
	if s.authConn != nil {
		if err := s.authConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close auth connection: %w", err))
		}
	}

	if s.userConn != nil {
		if err := s.userConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close user connection: %w", err))
		}
	}

	if s.geoConn != nil {
		if err := s.geoConn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close geo connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during close: %v", errs)
	}
	return nil
}
