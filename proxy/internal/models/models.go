package models

import (
	"time"
)

// GeoSearchRequest описывает запрос на поиск адреса
type GeoSearchRequest struct {
	Query string `json:"query" binding:"required" example:"Москва"`
}

// GeoGeocodeRequest описывает запрос на геокодирование
type GeoGeocodeRequest struct {
	Lat string `json:"lat" binding:"required,numeric,gte=-90,lte=90" example:"55.7558"`
	Lng string `json:"lng" binding:"required,numeric,gte=-180,lte=180" example:"37.6173"`
}

// GeoResponse представляет ответ от geo сервиса
type GeoResponse struct {
	Addresses []Address `json:"addresses"`
}

// Address представляет структуру адреса
type Address struct {
	Value string  `json:"value"`
	Lat   float64 `json:"lat" binding:"gte=-90,lte=90"`
	Lng   float64 `json:"lng" binding:"gte=-180,lte=180"`
}

// AuthRequest описывает запрос на auth сервис
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"qwerty123"`
}

// AuthResponse представляет ответ от auth сервиса
type AuthResponse struct {
	UserID    string `json:"user_id"`
	Token     string `json:"token,omitempty"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

// UserResponse представляет ответ от user сервиса
type UserResponse struct {
	ID        string    `json:"id" binding:"required"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code,omitempty"`
}

// ListUsersResponse представляет ответ со списком пользователей
type ListUsersResponse struct {
	Users      []UserResponse `json:"users"`
	TotalCount int            `json:"total_count"`
	Limit      int            `json:"limit,omitempty" binding:"min=1,max=100"`
	Offset     int            `json:"offset,omitempty" binding:"min=0"`
}
