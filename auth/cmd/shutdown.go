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
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), shutTime)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	grpcServer.GracefulStop()
	logger.Info("Servers stopped")
}
