package logger2

import (
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"log/slog"
	"os"
	"time"
)

// SetupLogger создает и возвращает логгер, общедоступный для всех слоёв
func SetupLogger(logLevel string, logFormat string) *slog.Logger {
	level := parseLogLevel(logLevel)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug,
	}

	var handler slog.Handler
	if logFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	return logger
}

// GinLoggerMiddleware возвращает interceptor для логирования HTTP-запросов
func GinLoggerMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		logger.Info("HTTP request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
			"client_ip", c.ClientIP(),
		)
	}
}

// GRPCLoggerInterceptor возвращает interceptor для логирования gRPC-запросов
func GRPCLoggerInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)

		logger.Info("gRPC request",
			"method", info.FullMethod,
			"duration", time.Since(start),
			"error", err,
		)

		return resp, err
	}
}

// parseLogLevel парсит уровень логирования из строки
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
