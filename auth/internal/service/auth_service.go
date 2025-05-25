package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/auth/proto"
	pbUser "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/user/proto"
	"time"
)

type AuthService interface {
	RegisterUser(ctx context.Context, email, password string) (*proto.AuthResponse, error)
	LoginUser(ctx context.Context, email, password string) (*proto.AuthResponse, error)
	ValidateToken(token string) (string, error)
}

type authService struct {
	userClient pbUser.UserServiceClient
	jwtSecret  string
	*slog.Logger
}

func NewAuthService(userClient pbUser.UserServiceClient, jwtSecret string, logger *slog.Logger) AuthService {
	return &authService{
		userClient: userClient,
		jwtSecret:  jwtSecret,
		Logger:     logger,
	}
}

// RegisterUser - регистрация нового пользователя.
func (s *authService) RegisterUser(ctx context.Context, email, password string) (*proto.AuthResponse, error) {
	// 1. Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 2. Создание пользователя
	s.Logger.Info("Calling UserService.CreateUser", "email", email)
	user, err := s.userClient.CreateUser(ctx, &pbUser.CreateUserRequest{
		Email:        email,
		PasswordHash: string(hashedPassword),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 3. Генерация JWT токена
	token, expiresAt, err := s.generateJWT(user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &proto.AuthResponse{
		Token:     token,
		UserId:    user.Id,
		ExpiresAt: expiresAt,
	}, nil
}

// LoginUser - аутентификация пользователя.
func (s *authService) LoginUser(ctx context.Context, email, password string) (*proto.AuthResponse, error) {
	// 1. Получение пользователя
	s.Logger.Info("Calling UserService.LoginUser", "email", email)
	user, err := s.userClient.GetUserByEmail(ctx, &pbUser.EmailRequest{Email: email})
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 3. Генерация токена
	token, expiresAt, err := s.generateJWT(user.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &proto.AuthResponse{
		Token:     token,
		UserId:    user.Id,
		ExpiresAt: expiresAt,
	}, nil
}

// ValidateToken - проверка JWT токена.
func (s *authService) ValidateToken(token string) (string, error) {
	claims, err := s.validateJWT(token)
	if err != nil {
		return "", errors.New("invalid token")
	}
	return claims.UserID, nil
}

// generateJWT - создает JWT токен.
func (s *authService) generateJWT(userID string) (string, int64, error) {
	expiresAt := time.Now().Add(24 * time.Hour).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     expiresAt,
	})
	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	return signedToken, expiresAt, err
}

// validateJWT - проверяет JWT токен.
func (s *authService) validateJWT(tokenString string) (*models.JwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.JwtClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
