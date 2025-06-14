package shutdown

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(ctx context.Context, server *grpc.Server, listener net.Listener, timeout time.Duration, logger *slog.Logger) error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("Received shutdown signal", "signal", sig)
	case <-ctx.Done():
		logger.Info("Context cancelled, initiating shutdown")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 1. Сначала закрываем listener
	if err := listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		logger.Error("Failed to close listener", "error", err)
	}

	// 2. Graceful shutdown сервера
	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Info("Server stopped gracefully")
		return nil
	case <-shutdownCtx.Done():
		server.Stop()
		return fmt.Errorf("graceful shutdown timed out after %v", timeout)
	}
}
