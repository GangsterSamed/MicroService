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
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" envDefault:"30s"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"info"`
	LogFormat       string        `env:"LOG_FORMAT" envDefault:"json"`

	// gRPC connection pooling settings
	GRPCDialTimeout       time.Duration `env:"GRPC_DIAL_TIMEOUT" envDefault:"30s"`
	GRPCMinConnectTimeout time.Duration `env:"GRPC_MIN_CONNECT_TIMEOUT" envDefault:"30s"`
	GRPCBackoffBaseDelay  time.Duration `env:"GRPC_BACKOFF_BASE_DELAY" envDefault:"2s"`
	GRPCBackoffMaxDelay   time.Duration `env:"GRPC_BACKOFF_MAX_DELAY" envDefault:"60s"`
	GRPCBackoffMultiplier float64       `env:"GRPC_BACKOFF_MULTIPLIER" envDefault:"1.6"`
	GRPCMaxRetries        int           `env:"GRPC_MAX_RETRIES" envDefault:"3"`

	// Redis caching settings
	RedisAddr         string        `env:"REDIS_ADDR" envDefault:"redis:6379"`
	RedisPassword     string        `env:"REDIS_PASSWORD" envDefault:""`
	RedisPoolSize     int           `env:"REDIS_POOL_SIZE" envDefault:"10"`
	RedisMinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" envDefault:"5"`
	RedisMaxRetries   int           `env:"REDIS_MAX_RETRIES" envDefault:"3"`
	RedisDialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" envDefault:"5s"`
	RedisReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" envDefault:"3s"`
	RedisWriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" envDefault:"3s"`

	// Cache settings
	CacheEnabled bool          `env:"CACHE_ENABLED" envDefault:"true"`
	CacheTTL     time.Duration `env:"CACHE_TTL" envDefault:"5m"`
	CacheMaxSize int           `env:"CACHE_MAX_SIZE" envDefault:"1000"`
}
