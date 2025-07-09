package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/handlers"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/middleware"
)

func healthCheckHandler(proxyService handlers.ProxyServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		statuses := map[string]string{
			"proxy": "OK",
			"geo":   "UNKNOWN",
			"auth":  "UNKNOWN",
			"user":  "UNKNOWN",
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		// Проверяем geo
		if err := proxyService.PingService(ctx, "geo"); err != nil {
			statuses["geo"] = "FAIL"
		} else {
			statuses["geo"] = "OK"
		}

		// Проверяем auth
		if err := proxyService.PingService(ctx, "auth"); err != nil {
			statuses["auth"] = "FAIL"
		} else {
			statuses["auth"] = "OK"
		}

		// Проверяем user
		if err := proxyService.PingService(ctx, "user"); err != nil {
			statuses["user"] = "FAIL"
		} else {
			statuses["user"] = "OK"
		}

		// Если хотя бы один сервис не OK, возвращаем 500
		statusCode := http.StatusOK
		for _, name := range []string{"geo", "auth", "user"} {
			if statuses[name] != "OK" {
				statusCode = http.StatusInternalServerError
				break
			}
		}

		c.JSON(statusCode, statuses)
	}
}

func startHTTPServer(cfg *config.ProxyConfig, proxyService handlers.ProxyServiceInterface, logger *slog.Logger) *http.Server {
	router := gin.New()

	// Используем middleware из handler
	router.Use(middleware.RequestIDMiddleware())
	router.Use(gin.Recovery())
	router.Use(logger2.GinLoggerMiddleware(logger))
	router.Use(middleware.PrometheusMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(60, time.Minute))
	router.Use(middleware.CacheMiddleware())

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// Initialize proxy handler
	proxyHandler, err := handlers.NewProxyHandler(proxyService, logger)
	if err != nil {
		logger.Error("Failed to initialize proxy handler", "error", err)
		return nil
	}

	// Роуты
	// Группа для API
	api := router.Group("/api")
	{
		// Auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/register", proxyHandler.HandleRegisterRequest())
			auth.POST("/login", proxyHandler.HandleLoginRequest())
		}

		// User endpoints
		user := api.Group("/user")
		{
			user.GET("/profile", proxyHandler.HandleProfileRequest())
			user.GET("/list", proxyHandler.HandleListRequest())
		}

		// Geo endpoints
		geo := api.Group("/address")
		{
			geo.POST("/search", proxyHandler.HandleSearchRequest())
			geo.POST("/geocode", proxyHandler.HandleGeocodeRequest())
		}
	}

	// Health check
	router.GET("/health", healthCheckHandler(proxyService))

	srv := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.HTTPPort),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		logger.Info("Starting HTTP server", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	return srv
}
