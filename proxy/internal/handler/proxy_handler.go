package handler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/models"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/service"
	"time"
)

type ProxyHandler struct {
	proxyService service.ProxyService
	logger       *slog.Logger
}

func NewProxyHandler(proxyService service.ProxyService, logger *slog.Logger) *ProxyHandler {
	return &ProxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}
}

func (h *ProxyHandler) HandleGeoRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			h.logger.Info("Geo request processed",
				slog.Duration("duration", time.Since(start)),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
		}()

		h.proxyRequest(c, "geo")
	}
}

func (h *ProxyHandler) HandleAuthRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			h.logger.Info("Auth request processed",
				slog.Duration("duration", time.Since(start)),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
		}()

		h.proxyRequest(c, "auth")
	}
}

func (h *ProxyHandler) HandleUserRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			h.logger.Info("User request processed",
				slog.Duration("duration", time.Since(start)),
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
			)
		}()

		h.proxyRequest(c, "user")
	}
}

func (h *ProxyHandler) proxyRequest(c *gin.Context, serviceName string) {
	// 1. Подготовка контекста с метаданными
	md := h.extractHeaders(c.Request)
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)

	// 2. Чтение данных запроса (разное для GET/POST)
	var reqBody []byte
	var err error

	if c.Request.Method == http.MethodGet {
		// Для GET-запросов собираем данные из query-параметров
		params := make(map[string]string)
		for k, v := range c.Request.URL.Query() {
			if len(v) > 0 {
				params[k] = v[0]
			}
		}
		reqBody, err = json.Marshal(params)
	} else {
		// Для POST/PUT читаем тело запроса
		reqBody, err = io.ReadAll(io.LimitReader(c.Request.Body, 10_000_000)) // 10 MB
		defer c.Request.Body.Close()
	}

	if err != nil {
		h.logger.Error("Failed to read request data", "error", err)
		h.writeGinError(c, http.StatusBadRequest, "Invalid request")
		return
	}

	// 3. Проксирование через интерфейс
	resp, err := h.proxyService.ForwardRequest(ctx, serviceName, c.Request.URL.Path, reqBody, md)
	if err != nil {
		h.handleGRPCError(c, err)
		return
	}

	// 4. Отправка ответа
	c.Data(http.StatusOK, "application/json", resp)
}

func (h *ProxyHandler) extractHeaders(r *http.Request) metadata.MD {
	md := make(metadata.MD)
	for k, v := range r.Header {
		if strings.HasPrefix(k, "Grpc-") || strings.HasPrefix(k, "X-") {
			md[strings.ToLower(k)] = v
		}
	}
	// Добавляем обязательные заголовки для gRPC
	if auth := r.Header.Get("Authorization"); auth != "" {
		md["authorization"] = []string{auth}
	}
	return md
}

func (h *ProxyHandler) handleGRPCError(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		h.logger.Error("Non-gRPC error: %v", err)
		h.writeGinError(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	switch st.Code() {
	case codes.InvalidArgument:
		h.logger.Info("Invalid argument", "message", st.Message())
		h.writeGinError(c, http.StatusBadRequest, st.Message())
	case codes.Unauthenticated:
		h.logger.Info("Unauthenticated", "message", st.Message())
		h.writeGinError(c, http.StatusUnauthorized, st.Message())
	case codes.PermissionDenied:
		h.logger.Info("Permission denied", "message", st.Message())
		h.writeGinError(c, http.StatusForbidden, st.Message())
	case codes.NotFound:
		h.logger.Info("Not found", "message", st.Message())
		h.writeGinError(c, http.StatusNotFound, st.Message())
	case codes.Unavailable:
		h.logger.Error("Service unavailable", "message", st.Message())
		h.writeGinError(c, http.StatusServiceUnavailable, st.Message())
	default:
		h.logger.Error("gRPC error", "error", err)
		h.writeGinError(c, http.StatusInternalServerError, "Internal server error")
	}
}

func (h *ProxyHandler) writeGinError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, models.ErrorResponse{
		Error: message,
		Code:  statusCode,
	})
}
