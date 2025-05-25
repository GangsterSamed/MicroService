package main

import (
	"context"
	"google.golang.org/grpc"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func waitForShutdown(httpServer *http.Server, grpcServer *grpc.Server, shutTime time.Duration, logger *slog.Logger) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down servers...", "timeout", shutTime)
	ctx, cancel := context.WithTimeout(context.Background(), shutTime)
	defer cancel()

	// Остановка HTTP сервера
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	// Остановка gRPC сервера
	grpcServer.GracefulStop()
	logger.Info("Servers stopped gracefully")
}
