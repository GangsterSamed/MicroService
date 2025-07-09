package cache

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GenerateKey(service, method string, data interface{}) string
	Close() error
}

type cacheService struct {
	redisClient *redis.Client
	logger      *slog.Logger
	enabled     bool
}

func NewCacheService(redisAddr, redisPassword string, poolSize, minIdleConns, maxRetries int,
	dialTimeout, readTimeout, writeTimeout time.Duration, logger *slog.Logger) (CacheService, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,

		// Connection pooling settings
		PoolSize:     poolSize,
		MinIdleConns: minIdleConns,
		MaxRetries:   maxRetries,

		// Timeout settings
		DialTimeout:  dialTimeout,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	})

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Warn("Failed to connect to Redis, caching will be disabled", "error", err)
		return &cacheService{
			redisClient: nil,
			logger:      logger,
			enabled:     false,
		}, nil
	}

	logger.Info("Redis cache service initialized",
		"addr", redisAddr,
		"pool_size", poolSize,
		"min_idle_conns", minIdleConns,
		"max_retries", maxRetries,
	)

	return &cacheService{
		redisClient: client,
		logger:      logger,
		enabled:     true,
	}, nil
}

func (c *cacheService) Get(ctx context.Context, key string) ([]byte, error) {
	if !c.enabled || c.redisClient == nil {
		return nil, fmt.Errorf("cache is disabled")
	}

	data, err := c.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			c.logger.Debug("Cache miss", "key", key)
			return nil, fmt.Errorf("key not found")
		}
		c.logger.Error("Failed to get from cache", "key", key, "error", err)
		return nil, err
	}

	c.logger.Debug("Cache hit", "key", key)
	return data, nil
}

func (c *cacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if !c.enabled || c.redisClient == nil {
		return nil // Игнорируем если кеш отключен
	}

	err := c.redisClient.Set(ctx, key, value, ttl).Err()
	if err != nil {
		c.logger.Error("Failed to set cache", "key", key, "error", err)
		return err
	}

	c.logger.Debug("Cache set", "key", key, "ttl", ttl)
	return nil
}

func (c *cacheService) Delete(ctx context.Context, key string) error {
	if !c.enabled || c.redisClient == nil {
		return nil
	}

	err := c.redisClient.Del(ctx, key).Err()
	if err != nil {
		c.logger.Error("Failed to delete from cache", "key", key, "error", err)
		return err
	}

	c.logger.Debug("Cache deleted", "key", key)
	return nil
}

func (c *cacheService) GenerateKey(service, method string, data interface{}) string {
	// Создаем уникальный ключ на основе сервиса, метода и данных
	keyData := fmt.Sprintf("%s:%s:%v", service, method, data)

	// Хешируем для получения фиксированной длины
	hash := md5.Sum([]byte(keyData))
	return fmt.Sprintf("proxy:%s", hex.EncodeToString(hash[:]))
}

func (c *cacheService) Close() error {
	if c.redisClient != nil {
		return c.redisClient.Close()
	}
	return nil
}

// CacheResponse представляет кешированный ответ
type CacheResponse struct {
	Data      []byte `json:"data"`
	Status    int    `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

// MarshalResponse сериализует ответ для кеширования
func MarshalResponse(data []byte, status int) ([]byte, error) {
	resp := CacheResponse{
		Data:      data,
		Status:    status,
		Timestamp: time.Now().Unix(),
	}
	return json.Marshal(resp)
}

// UnmarshalResponse десериализует ответ из кеша
func UnmarshalResponse(cachedData []byte) ([]byte, int, error) {
	var resp CacheResponse
	if err := json.Unmarshal(cachedData, &resp); err != nil {
		return nil, 0, err
	}
	return resp.Data, resp.Status, nil
}
