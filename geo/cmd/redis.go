package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
)

func initRedis(cfg *config.GeoConfig, logger *slog.Logger) (*redis.Client, error) {
	clientR := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,

		// Connection pooling settings
		PoolSize:     cfg.RedisPoolSize,
		MinIdleConns: cfg.RedisMinIdleConns,
		MaxRetries:   cfg.RedisMaxRetries,

		// Timeout settings
		DialTimeout:  cfg.RedisDialTimeout,
		ReadTimeout:  cfg.RedisReadTimeout,
		WriteTimeout: cfg.RedisWriteTimeout,
	})

	if _, err := clientR.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established with connection pooling",
		"pool_size", cfg.RedisPoolSize,
		"min_idle_conns", cfg.RedisMinIdleConns,
		"max_retries", cfg.RedisMaxRetries,
		"dial_timeout", cfg.RedisDialTimeout,
		"read_timeout", cfg.RedisReadTimeout,
		"write_timeout", cfg.RedisWriteTimeout,
	)

	return clientR, nil
}
