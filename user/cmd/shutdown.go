package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func waitForShutdown(srv *servers, shutTime time.Duration, logger *slog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), shutTime)
	defer cancel()

	if err := srv.http.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	srv.grpc.GracefulStop()
	logger.Info("Servers stopped gracefully")
}
