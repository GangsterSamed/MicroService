package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"google.golang.org/grpc"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc/backoff"
	authProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	geoProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/proto"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/cache"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/handlers"
	userProto "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

type proxyService struct {
	authHandler  *handlers.AuthHandler
	userHandler  *handlers.UserHandler
	geoHandler   *handlers.GeoHandler
	cacheManager *CacheManager
	logger       *slog.Logger
	geoConn      *grpc.ClientConn
	authConn     *grpc.ClientConn
	userConn     *grpc.ClientConn
	cfg          *config.ProxyConfig
}

// NewProxyService создает новый экземпляр прокси-сервиса
func NewProxyService(geoAddr string, authAddr string, userAddr string, logger *slog.Logger, cfg *config.ProxyConfig) (handlers.ProxyServiceInterface, error) {
	// Инициализируем кеш сервис
	cacheService, err := cache.NewCacheService(
		cfg.RedisAddr,
		cfg.RedisPassword,
		cfg.RedisPoolSize,
		cfg.RedisMinIdleConns,
		cfg.RedisMaxRetries,
		cfg.RedisDialTimeout,
		cfg.RedisReadTimeout,
		cfg.RedisWriteTimeout,
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache service: %v", err)
	}

	// Настройки gRPC connection pooling
	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: cfg.GRPCMinConnectTimeout,
			Backoff: backoff.Config{
				BaseDelay:  cfg.GRPCBackoffBaseDelay,
				Multiplier: cfg.GRPCBackoffMultiplier,
				MaxDelay:   cfg.GRPCBackoffMaxDelay,
			},
		}),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}

	// Создаем соединения с сервисами
	geoConn, err := grpc.Dial(geoAddr, dialOptions...)
	if err != nil {
		cacheService.Close()
		return nil, fmt.Errorf("failed to connect to geo service: %v", err)
	}

	authConn, err := grpc.Dial(authAddr, dialOptions...)
	if err != nil {
		cacheService.Close()
		geoConn.Close()
		return nil, fmt.Errorf("failed to connect to auth service: %v", err)
	}

	userConn, err := grpc.Dial(userAddr, dialOptions...)
	if err != nil {
		cacheService.Close()
		geoConn.Close()
		authConn.Close()
		return nil, fmt.Errorf("failed to connect to user service: %v", err)
	}

	// Создаем клиентов
	geoClient := geoProto.NewGeoServiceClient(geoConn)
	authClient := authProto.NewAuthServiceClient(authConn)
	userClient := userProto.NewUserServiceClient(userConn)

	// Создаем обработчики
	authHandler := handlers.NewAuthHandler(authClient, logger)
	userHandler := handlers.NewUserHandler(userClient, logger)
	geoHandler := handlers.NewGeoHandler(geoClient, logger)
	cacheManager := NewCacheManager(cacheService, logger, cfg)

	logger.Info("gRPC connections established with connection pooling",
		"geo_addr", geoAddr,
		"auth_addr", authAddr,
		"user_addr", userAddr,
		"min_connect_timeout", cfg.GRPCMinConnectTimeout,
		"backoff_base_delay", cfg.GRPCBackoffBaseDelay,
		"backoff_max_delay", cfg.GRPCBackoffMaxDelay,
		"backoff_multiplier", cfg.GRPCBackoffMultiplier,
		"cache_enabled", cfg.CacheEnabled,
	)

	return &proxyService{
		authHandler:  authHandler,
		userHandler:  userHandler,
		geoHandler:   geoHandler,
		cacheManager: cacheManager,
		logger:       logger,
		geoConn:      geoConn,
		authConn:     authConn,
		userConn:     userConn,
		cfg:          cfg,
	}, nil
}

func (s *proxyService) ForwardRequest(ctx context.Context, serviceName string, method string, reqBody []byte, headers metadata.MD) ([]byte, int, error) {
	// Добавляем timeout для запроса
	ctx, cancel := context.WithTimeout(ctx, s.cfg.RequestTimeout)
	defer cancel()

	// Проверяем кеш для кешируемых методов
	if s.cacheManager.ShouldCheckCache(serviceName, method, headers) {
		if cachedResponse, cachedStatus, err := s.cacheManager.GetCachedResponse(ctx, serviceName, method, reqBody, headers); err == nil {
			return cachedResponse, cachedStatus, nil
		}
	}

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

	var response []byte
	var statusCode int
	var err error

	switch serviceName {
	case "geo":
		response, statusCode, err = s.geoHandler.HandleGeoRequest(ctx, method, reqBody)
	case "auth":
		response, statusCode, err = s.authHandler.HandleAuthRequest(ctx, method, reqBody)
	case "user":
		response, statusCode, err = s.userHandler.HandleUserRequest(ctx, method, reqBody)
	default:
		return nil, http.StatusBadRequest, status.Error(codes.InvalidArgument, "invalid service name")
	}

	// Кешируем успешные ответы
	if err == nil && s.cacheManager.ShouldCacheResponse(serviceName, method, statusCode, headers) {
		s.cacheManager.CacheResponse(ctx, serviceName, method, reqBody, headers, response, statusCode)
	}

	return response, statusCode, err
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
		"metadata", md,
	)

	tokens := md.Get("authorization")
	s.logger.Info("Authorization tokens from metadata",
		"tokens", tokens,
	)
	if len(tokens) == 0 {
		s.logger.Warn("No authorization token in metadata")
		return nil, status.Error(codes.PermissionDenied, "no token provided")
	}

	// Извлекаем JWT токен
	jwtToken, err := s.authHandler.ExtractJWTToken(tokens[0])
	if err != nil {
		s.logger.Error("Failed to extract JWT token",
			"error", err.Error(),
			"token", tokens[0],
		)
		return nil, status.Error(codes.Unauthenticated, "invalid token format")
	}

	// Валидируем токен
	s.logger.Info("Validating token", "token", jwtToken)
	resp, err := s.authHandler.ValidateToken(ctx, jwtToken)
	if err != nil {
		s.logger.Error("ValidateToken error", "error", err)
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
			"existing_metadata", existingMD,
		)
		newMD = metadata.Join(existingMD, newMD)
	}

	s.logger.Info("Created new context with metadata",
		"new_metadata", newMD,
	)

	return metadata.NewOutgoingContext(ctx, newMD), nil
}

// Close закрывает все соединения
func (s *proxyService) Close() error {
	var errs []error

	// Закрываем кеш соединение
	if s.cacheManager != nil && s.cacheManager.cacheService != nil {
		if err := s.cacheManager.cacheService.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close cache connection: %w", err))
		}
	}

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

func (s *proxyService) PingService(ctx context.Context, serviceName string) error {
	switch serviceName {
	case "geo":
		// AddressSearch с пустым запросом
		_, err := s.geoHandler.GeoClient.AddressSearch(ctx, &geoProto.SearchRequest{Query: ""})
		return err
	case "auth":
		// ValidateToken с заведомо невалидным токеном
		_, err := s.authHandler.ValidateToken(ctx, "invalid-token-for-healthcheck")
		return err
	case "user":
		// ListUsers с limit=1
		_, err := s.userHandler.UserClient.ListUsers(ctx, &userProto.ListUsersRequest{Limit: 1, Offset: 0})
		return err
	default:
		return fmt.Errorf("unknown service: %s", serviceName)
	}
}
