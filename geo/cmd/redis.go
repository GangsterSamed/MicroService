package main

import (
	"context"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
)

func initRedis(cfg *config.GeoConfig, logger *slog.Logger) *redis.Client {
	clientR := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if _, err := clientR.Ping(context.Background()).Result(); err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	return clientR
}
