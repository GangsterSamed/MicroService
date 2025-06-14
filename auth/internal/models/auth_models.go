package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// User - модель пользователя для регистрации/авторизации
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email" validate:"required,email"`
	PasswordHash string    `json:"-"` // Хеш пароля (не возвращаем в JSON)
	CreatedAt    time.Time `json:"created_at"`
}

// RegisterRequest - запрос на регистрацию
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest - запрос на вход
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse - ответ с токеном
type AuthResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"` // Время в формате ISO 8601
	UserID    string `json:"user_id"`
}

// ErrorResponse - обертка для ошибок API
type ErrorResponse struct {
	Error string `json:"error"`
}

type JwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type HealthResponse struct {
	Status string `json:"status"`
}
