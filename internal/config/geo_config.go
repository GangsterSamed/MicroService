package config

import (
	"time"
)

type GeoConfig struct {
	Environment       string        `env:"ENV" envDefault:"development"`
	GRPCPort          int           `env:"GRPC_PORT" envDefault:"50051"`
	HTTPPort          int           `env:"HTTP_PORT" envDefault:"8081"`
	AuthServiceAddr   string        `env:"AUTH_SERVICE_ADDR" envDefault:"auth:50051"`
	RedisAddr         string        `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword     string        `env:"REDIS_PASSWORD" envDefault:""`
	CacheTTL          time.Duration `env:"CACHE_TTL" envDefault:"12h"`
	DadataAPIKey      string        `env:"DADATA_API_KEY,required"`
	DadataSecretKey   string        `env:"DADATA_SECRET_KEY,required"`
	LogLevel          string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat         string        `env:"LOG_FORMAT" envDefault:"json"`
	ShutdownTimeout   time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	MaxAddressResults int           `env:"MAX_ADDRESS_RESULTS" envDefault:"10"`
}
