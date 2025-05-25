package config

import (
	"time"
)

type AuthConfig struct {
	// Основные настройки
	Environment     string        `env:"ENV" envDefault:"development"`
	HTTPPort        int           `env:"HTTP_PORT" envDefault:"8081"`
	GRPCPort        int           `env:"GRPC_PORT" envDefault:"50051"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`

	// Настройки сервисов
	UserServiceAddr string `env:"USER_SERVICE_ADDR" envDefault:"user:50051"`

	// Настройки аутентификации
	JWTSecret      string `env:"JWT_SECRET,required"`
	JWTExpireHours int    `env:"JWT_EXPIRE_HOURS" envDefault:"24"`

	// Настройки логирования
	LogLevel  string `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat string `env:"LOG_FORMAT" envDefault:"json"`

	// Настройки базы данных
	DBHost     string `env:"DB_HOST" envDefault:"postgres"`
	DBPort     int    `env:"DB_PORT" envDefault:"5432"`
	DBUser     string `env:"DB_USER" envDefault:"user"`
	DBPassword string `env:"DB_PASSWORD" envDefault:"password"`
	DBName     string `env:"DB_NAME" envDefault:"auth"`
}
