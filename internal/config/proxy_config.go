package config

import (
	"time"
)

type ProxyConfig struct {
	Environment     string        `env:"ENV" envDefault:"development"`
	HTTPPort        int           `env:"HTTP_PORT" envDefault:"8080"`
	GRPCPort        int           `env:"GRPC_PORT" envDefault:"50051"`
	GeoServiceAddr  string        `env:"GEO_SERVICE_ADDR" envDefault:"geo:50051"`
	AuthServiceAddr string        `env:"AUTH_SERVICE_ADDR" envDefault:"auth:50051"`
	UserServiceAddr string        `env:"USER_SERVICE_ADDR" envDefault:"user:50051"`
	JWTSecret       string        `env:"JWT_SECRET" envDefault:"your-secret-key"`
	ReadTimeout     time.Duration `env:"PROXY_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"PROXY_WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout     time.Duration `env:"PROXY_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat       string        `env:"LOG_FORMAT" envDefault:"json"`
}
