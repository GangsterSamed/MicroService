package main

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log/slog"
	"net/http"
	"strconv"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/handler"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/service"
)

// Структура для хранения серверов
type servers struct {
	http *http.Server
}

func startHTTPServer(cfg *config.ProxyConfig, proxyService service.ProxyService, logger *slog.Logger) *http.Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger2.GinLoggerMiddleware(logger))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Инициализация обработчиков
	proxyHandler := handler.NewProxyHandler(proxyService, logger)

	// Роуты
	api := router.Group("/api")
	{
		api.POST("/address/search", proxyHandler.HandleGeoRequest())
		api.POST("/address/geocode", proxyHandler.HandleGeoRequest())
		api.POST("/auth/register", proxyHandler.HandleAuthRequest())
		api.POST("/auth/login", proxyHandler.HandleAuthRequest())
		api.GET("/user/profile", proxyHandler.HandleUserRequest())
		api.GET("/user/list", proxyHandler.HandleUserRequest())
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	})

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
