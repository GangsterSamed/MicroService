package config

import (
	"time"
)

type AuthConfig struct {
	// Основные настройки
	Environment     string        `env:"ENV" envDefault:"development"`
	HTTPPort        int           `env:"HTTP_PORT" envDefault:"8080"`
	GRPCPort        int           `env:"GRPC_PORT" envDefault:"50051"`
	UserServiceAddr string        `env:"USER_SERVICE_ADDR" envDefault:"user:50051"`
	JWTSecret       string        `env:"JWT_SECRET" envDefault:"your-secret-key"`
	ReadTimeout     time.Duration `env:"AUTH_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"AUTH_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"AUTH_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	// Настройки логирования
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`

	// Настройки базы данных
	DBHost     string `env:"DB_HOST" envDefault:"postgres"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"user"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"password"`
	DBName     string `env:"DB_NAME" envDefault:"auth"`

	// gRPC connection pooling settings
	GRPCDialTimeout       time.Duration `env:"GRPC_DIAL_TIMEOUT" envDefault:"30s"`
	GRPCMinConnectTimeout time.Duration `env:"GRPC_MIN_CONNECT_TIMEOUT" envDefault:"30s"`
	GRPCBackoffBaseDelay  time.Duration `env:"GRPC_BACKOFF_BASE_DELAY" envDefault:"2s"`
	GRPCBackoffMaxDelay   time.Duration `env:"GRPC_BACKOFF_MAX_DELAY" envDefault:"60s"`
	GRPCBackoffMultiplier float64       `env:"GRPC_BACKOFF_MULTIPLIER" envDefault:"1.6"`
	GRPCMaxRetries        int           `env:"GRPC_MAX_RETRIES" envDefault:"3"`
}
