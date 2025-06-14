package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	pbUser "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email, password string) (*proto.RegisterResponse, error)
	LoginUser(ctx context.Context, email, password string) (*proto.AuthResponse, error)
	ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error)
}

type authService struct {
	userClient pbUser.UserServiceClient
	jwtSecret  string
	*slog.Logger
}

func NewAuthService(userClient pbUser.UserServiceClient, jwtSecret string, logger *slog.Logger) AuthService {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверяем подключение к user сервису
	_, err := userClient.GetUserByEmail(ctx, &pbUser.EmailRequest{Email: "test@test.com"})
	if err != nil {
		// Игнорируем ошибку "user not found", так как это ожидаемое поведение
		if !strings.Contains(err.Error(), "user not found") {
			logger.Warn("User service connection check failed", "error", err)
		} else {
			logger.Info("User service connection verified")
		}
	} else {
		logger.Info("User service connection verified")
	}

	return &authService{
		userClient: userClient,
		jwtSecret:  jwtSecret,
		Logger:     logger,
	}
}

// RegisterUser - регистрация нового пользователя.
func (s *authService) RegisterUser(ctx context.Context, email, password string) (*proto.RegisterResponse, error) {
	logger := s.Logger.With("method", "RegisterUser", "email", email)
	logger.Info("Starting user registration")

	// 1. Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("Failed to hash password", "error", err)
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 2. Создание пользователя
	logger.Info("Creating user in UserService")
	user, err := s.userClient.CreateUser(ctx, &pbUser.CreateUserRequest{
		Email:        email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		// Проверка на дубликат e-mail
		if strings.Contains(err.Error(), "email already exists") {
			logger.Warn("Registration failed: user already exists")
			return nil, errors.New("user already exists")
		}
		logger.Error("Failed to create user", "error", err)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info("User registered successfully", "user_id", user.Id)
	return &proto.RegisterResponse{
		UserId: user.Id,
	}, nil
}

// LoginUser - аутентификация пользователя.
func (s *authService) LoginUser(ctx context.Context, email, password string) (*proto.AuthResponse, error) {
	logger := s.Logger.With("method", "LoginUser", "email", email)
	logger.Info("Starting user login")

	// 1. Получение пользователя
	logger.Info("Getting user from UserService")
	user, err := s.userClient.GetUserByEmail(ctx, &pbUser.EmailRequest{Email: email})
	if err != nil {
		logger.Warn("Login failed: user not found")
		return nil, errors.New("invalid credentials")
	}

	// 2. Проверка пароля
	logger.Info("Verifying password")
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		logger.Warn("Login failed: invalid password")
		return nil, errors.New("invalid credentials")
	}

	// 3. Генерация токена
	logger.Info("Generating JWT token")
	token, expiresAt, err := s.generateJWT(user.Id)
	if err != nil {
		logger.Error("Failed to generate token", "error", err)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	logger.Info("Login successful", "user_id", user.Id)
	return &proto.AuthResponse{
		Token:     token,
		UserId:    user.Id,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateToken - проверка JWT токена через gRPC.
func (s *authService) ValidateToken(ctx context.Context, req *proto.TokenRequest) (*proto.TokenResponse, error) {
	logger := s.Logger.With("method", "ValidateToken")
	logger.Info("Starting token validation")

	claims, err := s.validateJWT(req.Token)
	if err != nil {
		logger.Warn("Token validation failed", "error", err)
		return nil, err
	}

	logger.Info("Token validated successfully", "user_id", claims.UserID)
	return &proto.TokenResponse{Valid: true, UserId: claims.UserID}, nil
}

// generateJWT - создает JWT токен.
func (s *authService) generateJWT(userID string) (string, string, error) {
	expiresAt := time.Now().Add(15 * time.Minute)
	claims := models.JwtClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	return signedToken, expiresAt.Format(time.RFC3339), err
}

// validateJWT - внутренний метод для проверки JWT токена.
func (s *authService) validateJWT(token string) (*models.JwtClaims, error) {
	logger := s.Logger.With("method", "validateJWT")
	logger.Info("Validating JWT token")

	// Убираем префикс Bearer если он есть
	token = strings.TrimPrefix(token, "Bearer ")

	// Проверяем, что токен имеет правильный формат JWT (три части, разделенные точками)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		logger.Warn("Invalid token format", "parts_count", len(parts))
		return nil, fmt.Errorf("token validation failed: token is malformed: token contains an invalid number of segments")
	}

	claims := &models.JwtClaims{}
	tokenObj, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		logger.Error("Token validation failed", "error", err)
		return nil, fmt.Errorf("token validation failed: %w", err)
	}
	if !tokenObj.Valid {
		logger.Warn("Token is not valid")
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
