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
