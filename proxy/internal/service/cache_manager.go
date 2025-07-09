package service

import (
	"context"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/metadata"

	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/cache"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/metrics"
)

// Кешируемые методы
var cacheableMethods = map[string]bool{
	"/api/address/search":  true,
	"/api/address/geocode": true,
	"/api/user/list":       true,
}

// CacheManager управляет кешированием запросов
type CacheManager struct {
	cacheService cache.CacheService
	logger       *slog.Logger
	cfg          *config.ProxyConfig
}

// NewCacheManager создает новый менеджер кеширования
func NewCacheManager(cacheService cache.CacheService, logger *slog.Logger, cfg *config.ProxyConfig) *CacheManager {
	return &CacheManager{
		cacheService: cacheService,
		logger:       logger,
		cfg:          cfg,
	}
}

// ShouldCheckCache определяет, нужно ли проверять кеш для данного запроса
func (cm *CacheManager) ShouldCheckCache(serviceName, method string, headers metadata.MD) bool {
	if !cm.cfg.CacheEnabled {
		return false
	}
	if !cacheableMethods[method] {
		return false
	}
	// Только если есть user_id (т.е. пользователь авторизован)
	if userID := headers.Get("user_id"); len(userID) == 0 {
		return false
	}
	return true
}

// ShouldCacheResponse определяет, нужно ли кешировать ответ
func (cm *CacheManager) ShouldCacheResponse(serviceName, method string, statusCode int, headers metadata.MD) bool {
	if !cm.cfg.CacheEnabled {
		return false
	}
	if statusCode != http.StatusOK {
		return false
	}
	if !cacheableMethods[method] {
		return false
	}
	// Только если есть user_id (т.е. пользователь авторизован)
	if userID := headers.Get("user_id"); len(userID) == 0 {
		return false
	}
	return true
}

// GenerateCacheKey создает уникальный ключ для кеширования
func (cm *CacheManager) GenerateCacheKey(serviceName, method string, reqBody []byte, headers metadata.MD) string {
	// Создаем данные для ключа
	keyData := map[string]interface{}{
		"service": serviceName,
		"method":  method,
		"body":    string(reqBody),
	}

	// Добавляем важные заголовки
	if userID := headers.Get("user_id"); len(userID) > 0 {
		keyData["user_id"] = userID[0]
	}

	// Добавляем параметры пагинации
	if limit := headers.Get("limit"); len(limit) > 0 {
		keyData["limit"] = limit[0]
	}
	if offset := headers.Get("offset"); len(offset) > 0 {
		keyData["offset"] = offset[0]
	}

	return cm.cacheService.GenerateKey(serviceName, method, keyData)
}

// GetCachedResponse получает кешированный ответ
func (cm *CacheManager) GetCachedResponse(ctx context.Context, serviceName, method string, reqBody []byte, headers metadata.MD) ([]byte, int, error) {
	cacheKey := cm.GenerateCacheKey(serviceName, method, reqBody, headers)
	if cachedData, err := cm.cacheService.Get(ctx, cacheKey); err == nil {
		cm.logger.Info("Cache hit", "service", serviceName, "method", method)
		metrics.RecordCacheMetrics(serviceName, method, true)
		return cache.UnmarshalResponse(cachedData)
	} else {
		metrics.RecordCacheMetrics(serviceName, method, false)
		return nil, 0, err
	}
}

// CacheResponse кеширует ответ
func (cm *CacheManager) CacheResponse(ctx context.Context, serviceName, method string, reqBody []byte, headers metadata.MD, response []byte, statusCode int) {
	cacheKey := cm.GenerateCacheKey(serviceName, method, reqBody, headers)
	if cachedData, marshalErr := cache.MarshalResponse(response, statusCode); marshalErr == nil {
		if cacheErr := cm.cacheService.Set(ctx, cacheKey, cachedData, cm.cfg.CacheTTL); cacheErr != nil {
			cm.logger.Warn("Failed to cache response", "error", cacheErr)
		} else {
			cm.logger.Debug("Response cached", "service", serviceName, "method", method)
		}
	}
}
