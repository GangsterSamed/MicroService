package models

import "net/http"

type GeoRequest struct {
	Query string `json:"query" example:"Москва"`
}

type GeoResponse struct {
	Addresses []Address `json:"addresses"`
}

type Address struct {
	City   string `json:"city" example:"Москва"`
	Street string `json:"street" example:"Арбат"`
}

type AuthRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"qwerty123"`
}

type AuthResponse struct {
	Token string `json:"token" example:"eyJhbGciOi..."`
}

type ErrorResponse struct {
	Error  string `json:"error"`            // Человекочитаемое описание
	Code   int    `json:"code"`             // HTTP-статус код
	Detail string `json:"detail,omitempty"` // Технические детали (для разработки)
}

func NewErrorResponse(statusCode int, err error) *ErrorResponse {
	return &ErrorResponse{
		Error:  http.StatusText(statusCode),
		Code:   statusCode,
		Detail: err.Error(),
	}
}
