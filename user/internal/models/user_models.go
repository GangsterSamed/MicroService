package models

import (
	"time"
)

// User - основная модель пользователя
type User struct {
	ID           string    `db:"id" json:"id"`
	Email        string    `db:"email" json:"email" validate:"required,email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

// UpdateUserRequest - модель для обновления данных пользователя
type UpdateUserRequest struct {
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=8"`
}

// UserProfileResponse - модель ответа с профилем пользователя
type UserProfileResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListUsersResponse struct {
	Users      []UserProfile `json:"users"`
	TotalCount int           `json:"total_count"`
	Limit      int           `json:"limit,omitempty"`
	Offset     int           `json:"offset,omitempty"`
}

type UserProfile struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Status    string    `json:"status,omitempty"` // active/banned etc
}

// ErrorResponse - модель для ошибок API
type ErrorResponse struct {
	Error string `json:"error"`
}
