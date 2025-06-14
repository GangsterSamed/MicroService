package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/config"
	logger2 "studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/internal/logger"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/handler"
	"studentgit.kata.academy/romanmalcev89665_gmail.com/go-kata/new-repository/MicroService/proxy/internal/service"
)

func startHTTPServer(cfg *config.ProxyConfig, proxyService service.ProxyService, logger *slog.Logger) *http.Server {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(logger2.GinLoggerMiddleware(logger))

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger/doc.json")))

	// Initialize proxy handler
	proxyHandler, err := handler.NewProxyHandler(proxyService, logger)
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
