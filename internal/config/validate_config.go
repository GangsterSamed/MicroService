package config

import (
	"fmt"
	"strings"
)

// --- Хелпер-функции для валидации ---
func checkPort(val int, name string, errors *[]string) {
	if val <= 0 || val > 65535 {
		*errors = append(*errors, fmt.Sprintf("%s must be between 1 and 65535", name))
	}
}

func checkPositiveInt(val int, name string, errors *[]string) {
	if val <= 0 {
		*errors = append(*errors, fmt.Sprintf("%s must be positive", name))
	}
}

func checkPositiveDuration(val interface{}, name string, errors *[]string) {
	switch v := val.(type) {
	case int:
		if v <= 0 {
			*errors = append(*errors, fmt.Sprintf("%s must be positive", name))
		}
	case int64:
		if v <= 0 {
			*errors = append(*errors, fmt.Sprintf("%s must be positive", name))
		}
	case float64:
		if v <= 0 {
			*errors = append(*errors, fmt.Sprintf("%s must be positive", name))
		}
	case string:
		// не проверяем
	default:
		// не проверяем
	}
}

func checkRequired(val, name string, errors *[]string) {
	if val == "" {
		*errors = append(*errors, fmt.Sprintf("%s is required", name))
	}
}

func checkMinLen(val, name string, min int, errors *[]string) {
	if len(val) < min {
		*errors = append(*errors, fmt.Sprintf("%s must be at least %d characters long", name, min))
	}
}

func checkLogLevel(val string, errors *[]string) {
	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[strings.ToLower(val)] {
		*errors = append(*errors, "LOG_LEVEL must be one of: debug, info, warn, error")
	}
}

func checkLogFormat(val string, errors *[]string) {
	validLogFormats := map[string]bool{"json": true, "text": true}
	if !validLogFormats[strings.ToLower(val)] {
		*errors = append(*errors, "LOG_FORMAT must be one of: json, text")
	}
}

// --- Компактные функции валидации ---
func validateProxyConfig(cfg *ProxyConfig) error {
	errors := []string{}
	checkPort(cfg.HTTPPort, "HTTP_PORT", &errors)
	checkPort(cfg.GRPCPort, "GRPC_PORT", &errors)
	checkRequired(cfg.GeoServiceAddr, "GEO_SERVICE_ADDR", &errors)
	checkRequired(cfg.AuthServiceAddr, "AUTH_SERVICE_ADDR", &errors)
	checkRequired(cfg.UserServiceAddr, "USER_SERVICE_ADDR", &errors)
	if cfg.JWTSecret == "" || cfg.JWTSecret == "your-secret-key" {
		errors = append(errors, "JWT_SECRET must be set to a secure value")
	}
	checkPositiveDuration(cfg.ReadTimeout, "PROXY_READ_TIMEOUT", &errors)
	checkPositiveDuration(cfg.WriteTimeout, "PROXY_WRITE_TIMEOUT", &errors)
	checkPositiveDuration(cfg.IdleTimeout, "PROXY_IDLE_TIMEOUT", &errors)
	checkPositiveDuration(cfg.ShutdownTimeout, "SHUTDOWN_TIMEOUT", &errors)
	checkPositiveDuration(cfg.RequestTimeout, "REQUEST_TIMEOUT", &errors)
	checkRequired(cfg.RedisAddr, "REDIS_ADDR", &errors)
	checkPositiveInt(cfg.RedisPoolSize, "REDIS_POOL_SIZE", &errors)
	if cfg.RedisMinIdleConns < 0 {
		errors = append(errors, "REDIS_MIN_IDLE_CONNS must be non-negative")
	}
	if cfg.RedisMinIdleConns > cfg.RedisPoolSize {
		errors = append(errors, "REDIS_MIN_IDLE_CONNS cannot be greater than REDIS_POOL_SIZE")
	}
	checkPositiveDuration(cfg.CacheTTL, "CACHE_TTL", &errors)
	checkPositiveInt(cfg.CacheMaxSize, "CACHE_MAX_SIZE", &errors)
	checkLogLevel(cfg.LogLevel, &errors)
	checkLogFormat(cfg.LogFormat, &errors)
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}
	return nil
}

func validateAuthConfig(cfg *AuthConfig) error {
	errors := []string{}
	checkPort(cfg.GRPCPort, "GRPC_PORT", &errors)
	checkRequired(cfg.UserServiceAddr, "USER_SERVICE_ADDR", &errors)
	if cfg.JWTSecret == "" || cfg.JWTSecret == "your-secret-key" {
		errors = append(errors, "JWT_SECRET must be set to a secure value")
	}
	checkMinLen(cfg.JWTSecret, "JWT_SECRET", 32, &errors)
	checkPositiveDuration(cfg.ReadTimeout, "AUTH_READ_TIMEOUT", &errors)
	checkPositiveDuration(cfg.WriteTimeout, "AUTH_WRITE_TIMEOUT", &errors)
	checkPositiveDuration(cfg.IdleTimeout, "AUTH_IDLE_TIMEOUT", &errors)
	checkPositiveDuration(cfg.ShutdownTimeout, "SHUTDOWN_TIMEOUT", &errors)
	checkPositiveDuration(cfg.GRPCDialTimeout, "GRPC_DIAL_TIMEOUT", &errors)
	checkPositiveDuration(cfg.GRPCMinConnectTimeout, "GRPC_MIN_CONNECT_TIMEOUT", &errors)
	checkPositiveDuration(cfg.GRPCBackoffBaseDelay, "GRPC_BACKOFF_BASE_DELAY", &errors)
	checkPositiveDuration(cfg.GRPCBackoffMaxDelay, "GRPC_BACKOFF_MAX_DELAY", &errors)
	if cfg.GRPCBackoffMultiplier <= 0 {
		errors = append(errors, "GRPC_BACKOFF_MULTIPLIER must be positive")
	}
	if cfg.GRPCMaxRetries < 0 {
		errors = append(errors, "GRPC_MAX_RETRIES must be non-negative")
	}
	checkLogLevel(cfg.LogLevel, &errors)
	checkLogFormat(cfg.LogFormat, &errors)
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}
	return nil
}

func validateGeoConfig(cfg *GeoConfig) error {
	errors := []string{}
	checkPort(cfg.GRPCPort, "GRPC_PORT", &errors)
	checkRequired(cfg.DadataAPIKey, "DADATA_API_KEY", &errors)
	checkRequired(cfg.DadataSecretKey, "DADATA_SECRET_KEY", &errors)
	checkRequired(cfg.RedisAddr, "REDIS_ADDR", &errors)
	checkPositiveDuration(cfg.ShutdownTimeout, "SHUTDOWN_TIMEOUT", &errors)
	checkPositiveDuration(cfg.CacheTTL, "CACHE_TTL", &errors)
	checkPositiveInt(cfg.MaxAddressResults, "MAX_ADDRESS_RESULTS", &errors)
	checkPositiveInt(cfg.RedisPoolSize, "REDIS_POOL_SIZE", &errors)
	if cfg.RedisMinIdleConns < 0 {
		errors = append(errors, "REDIS_MIN_IDLE_CONNS must be non-negative")
	}
	if cfg.RedisMinIdleConns > cfg.RedisPoolSize {
		errors = append(errors, "REDIS_MIN_IDLE_CONNS cannot be greater than REDIS_POOL_SIZE")
	}
	if cfg.RedisMaxRetries < 0 {
		errors = append(errors, "REDIS_MAX_RETRIES must be non-negative")
	}
	checkPositiveDuration(cfg.RedisDialTimeout, "REDIS_DIAL_TIMEOUT", &errors)
	checkPositiveDuration(cfg.RedisReadTimeout, "REDIS_READ_TIMEOUT", &errors)
	checkPositiveDuration(cfg.RedisWriteTimeout, "REDIS_WRITE_TIMEOUT", &errors)
	checkLogLevel(cfg.LogLevel, &errors)
	checkLogFormat(cfg.LogFormat, &errors)
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}
	return nil
}

func validateUserConfig(cfg *UserConfig) error {
	errors := []string{}
	checkPort(cfg.GRPCPort, "GRPC_PORT", &errors)
	checkRequired(cfg.DBHost, "DB_HOST", &errors)
	checkRequired(cfg.DBPort, "DB_PORT", &errors)
	checkRequired(cfg.DBUser, "DB_USER", &errors)
	checkRequired(cfg.DBPassword, "DB_PASSWORD", &errors)
	checkRequired(cfg.DBName, "DB_NAME", &errors)
	checkPositiveDuration(cfg.ReadTimeout, "HTTP_READ_TIMEOUT", &errors)
	checkPositiveDuration(cfg.WriteTimeout, "HTTP_WRITE_TIMEOUT", &errors)
	checkPositiveDuration(cfg.IdleTimeout, "HTTP_IDLE_TIMEOUT", &errors)
	checkPositiveDuration(cfg.ShutdownTimeout, "SHUTDOWN_TIMEOUT", &errors)
	checkPositiveInt(cfg.DBMaxOpenConns, "DB_MAX_OPEN_CONNS", &errors)
	checkPositiveInt(cfg.DBMaxIdleConns, "DB_MAX_IDLE_CONNS", &errors)
	if cfg.DBMaxIdleConns > cfg.DBMaxOpenConns {
		errors = append(errors, "DB_MAX_IDLE_CONNS cannot be greater than DB_MAX_OPEN_CONNS")
	}
	checkPositiveDuration(cfg.DBConnMaxLifetime, "DB_CONN_MAX_LIFETIME", &errors)
	checkPositiveDuration(cfg.DBConnMaxIdleTime, "DB_CONN_MAX_IDLE_TIME", &errors)
	checkLogLevel(cfg.LogLevel, &errors)
	checkLogFormat(cfg.LogFormat, &errors)
	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}
	return nil
}
