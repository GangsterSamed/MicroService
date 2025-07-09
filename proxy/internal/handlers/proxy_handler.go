package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/errors"
)

// ProxyServiceInterface определяет интерфейс для прокси-сервиса
type ProxyServiceInterface interface {
	ForwardRequest(ctx context.Context, serviceName, path string, body []byte, md metadata.MD) ([]byte, int, error)
	Close() error
	PingService(ctx context.Context, serviceName string) error
}

type ProxyHandler interface {
	HandleRegisterRequest() gin.HandlerFunc
	HandleLoginRequest() gin.HandlerFunc
	HandleSearchRequest() gin.HandlerFunc
	HandleGeocodeRequest() gin.HandlerFunc
	HandleProfileRequest() gin.HandlerFunc
	HandleListRequest() gin.HandlerFunc
}

type proxyHandler struct {
	proxyService ProxyServiceInterface
	logger       *slog.Logger
}

func NewProxyHandler(proxyService ProxyServiceInterface, logger *slog.Logger) (ProxyHandler, error) {
	return &proxyHandler{
		proxyService: proxyService,
		logger:       logger,
	}, nil
}

// @Summary Регистрация нового пользователя
// @Description Регистрирует нового пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.AuthRequest true "Данные для регистрации (email и password минимум 8 символов)"
// @Success 201 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/register [post]
func (h *proxyHandler) HandleRegisterRequest() gin.HandlerFunc {
	return h.withLogging("Register", h.handleAuthRequest)
}

// @Summary Авторизация пользователя
// @Description Авторизует пользователя в системе
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.AuthRequest true "Данные для авторизации (email и password)"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/auth/login [post]
func (h *proxyHandler) HandleLoginRequest() gin.HandlerFunc {
	return h.withLogging("Login", h.handleAuthRequest)
}

// @Summary Поиск адресов
// @Description Поиск адресов по поисковому запросу
// @Tags geo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.GeoSearchRequest true "Поисковый запрос" SchemaExample({"query": "Москва"})
// @Success 200 {object} models.GeoResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/address/search [post]
func (h *proxyHandler) HandleSearchRequest() gin.HandlerFunc {
	return h.withLogging("Search", h.handleGeoRequest)
}

// @Summary Геокодирование адреса
// @Description Получение адреса по координатам
// @Tags geo
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.GeoGeocodeRequest true "Координаты для геокодирования" SchemaExample({"lat": "55.7558", "lng": "37.6173"})
// @Success 200 {object} models.GeoResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/address/geocode [post]
func (h *proxyHandler) HandleGeocodeRequest() gin.HandlerFunc {
	return h.withLogging("Geocode", h.handleGeoRequest)
}

// @Summary Получение профиля пользователя
// @Description Возвращает профиль текущего авторизованного пользователя
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.UserResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/user/profile [get]
func (h *proxyHandler) HandleProfileRequest() gin.HandlerFunc {
	return h.withLogging("Profile", h.handleUserRequest)
}

// @Summary Получение списка пользователей
// @Description Возвращает список пользователей с пагинацией
// @Tags user
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param limit query int false "Лимит" default(10)
// @Param offset query int false "Смещение" default(0)
// @Success 200 {object} models.ListUsersResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/user/list [get]
func (h *proxyHandler) HandleListRequest() gin.HandlerFunc {
	return h.withLogging("List", h.handleUserRequest)
}

func (h *proxyHandler) handleAuthRequest(c *gin.Context) {
	h.handleRequestWithBody(c, "auth")
}

func (h *proxyHandler) handleGeoRequest(c *gin.Context) {
	h.handleRequestWithBody(c, "geo")
}

// handleRequestWithBody обрабатывает запросы с телом (POST)
func (h *proxyHandler) handleRequestWithBody(c *gin.Context, serviceName string) {
	requestData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		errors.WriteError(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	headers := h.extractHeaders(c.Request)
	ctx := c.Request.Context()
	if serviceName != "auth" {
		ctx = metadata.NewOutgoingContext(ctx, headers)
	}
	response, statusCode, err := h.proxyService.ForwardRequest(ctx, serviceName, c.Request.URL.Path, requestData, headers)

	h.handleResponse(c, response, statusCode, err)
}

func (h *proxyHandler) handleUserRequest(c *gin.Context) {
	headers := h.extractHeaders(c.Request)

	// Добавляем параметры пагинации в метаданные
	if c.Request.URL.Path == "/api/user/list" {
		limit := c.DefaultQuery("limit", "10")
		offset := c.DefaultQuery("offset", "0")
		headers.Set("limit", limit)
		headers.Set("offset", offset)
	}

	ctx := metadata.NewOutgoingContext(c.Request.Context(), headers)
	response, statusCode, err := h.proxyService.ForwardRequest(ctx, "user", c.Request.URL.Path, nil, headers)

	h.handleResponse(c, response, statusCode, err)
}

func (h *proxyHandler) extractHeaders(r *http.Request) metadata.MD {
	md := metadata.MD{}

	// Логируем все заголовки запроса
	h.logger.Info("Received headers",
		"headers", r.Header,
	)

	// Добавим все заголовки, которые начинаются с X- или Grpc- (без учёта регистра)
	for k, v := range r.Header {
		lowerKey := strings.ToLower(k)
		if strings.HasPrefix(lowerKey, "x-") || strings.HasPrefix(lowerKey, "grpc-") {
			md[lowerKey] = v
		}
	}

	// Обрабатываем authorization, если он есть
	if auth := r.Header.Get("authorization"); auth != "" {
		h.logger.Info("Found authorization header",
			"auth", auth,
		)
		md["authorization"] = []string{auth}
	} else {
		h.logger.Warn("No authorization header found")
	}

	h.logger.Info("Created metadata",
		"metadata", md,
	)

	return md
}

func (h *proxyHandler) handleResponse(c *gin.Context, response []byte, statusCode int, err error) {
	if err != nil {
		errors.HandleError(c, err)
		return
	}

	var responseMap map[string]interface{}
	if err := json.Unmarshal(response, &responseMap); err != nil {
		errors.WriteError(c, http.StatusInternalServerError, "failed to parse response")
		return
	}

	c.JSON(statusCode, responseMap)
}

func (h *proxyHandler) withLogging(name string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			h.logger.Info(name+" request processed",
				"duration", time.Since(start),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
		}()
		handler(c)
	}
}
