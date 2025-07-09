package config

import (
	"time"
)

// Config содержит все параметры конфигурации
type UserConfig struct {
	Environment     string        `env:"ENV" envDefault:"development"`
	HTTPPort        int           `env:"HTTP_PORT" envDefault:"8080"`
	GRPCPort        int           `env:"GRPC_PORT" envDefault:"50051"`
	DBHost          string        `env:"DB_HOST" envDefault:"postgres"`
	DBPort          string        `env:"DB_PORT" envDefault:"5432"`
	DBUser          string        `env:"DB_USER" envDefault:"user"`
	DBPassword      string        `env:"DB_PASSWORD" envDefault:"password"`
	DBName          string        `env:"DB_NAME" envDefault:"users"`
	DBSSLMode       string        `env:"DB_SSLMODE" envDefault:"disable"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat       string        `env:"LOG_FORMAT" envDefault:"json"`

	// Database connection pooling settings
	DBMaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	DBMaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
	DBConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	DBConnMaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"5m"`
}
