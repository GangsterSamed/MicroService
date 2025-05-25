package client

import (
	"context"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/geo/internal/models"
	"time"
)

// GeoClient - общий интерфейс для всех клиентов
type GeoClient interface {
	AddressSearch(ctx context.Context, query string) ([]*models.Address, error)
	GeoCode(ctx context.Context, lat, lng string) ([]*models.Address, error)
	Close() error
}

// BaseClient - общая структура для клиентов
type BaseClient struct {
	timeout time.Duration
}

func (c *BaseClient) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.timeout > 0 {
		return context.WithTimeout(ctx, c.timeout)
	}
	return ctx, func() {}
