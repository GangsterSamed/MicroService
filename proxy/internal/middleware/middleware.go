package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
}

// PrometheusMiddleware записывает метрики Prometheus
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		labels := prometheus.Labels{
			"method": c.Request.Method,
			"path":   c.FullPath(),
			"status": fmt.Sprintf("%d", status),
		}
		httpRequestsTotal.With(labels).Inc()
		httpRequestDuration.With(labels).Observe(duration)
	}
}

// RequestIDMiddleware генерирует и прокидывает Request ID
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// CORSMiddleware настраивает CORS заголовки
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// RateLimitMiddleware ограничивает количество запросов
func RateLimitMiddleware(maxRequests int, window time.Duration) gin.HandlerFunc {
	type clientData struct {
		count     int
		lastReset time.Time
	}
	var (
		clients = make(map[string]*clientData)
		mu      = &sync.Mutex{}
	)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		mu.Lock()
		data, exists := clients[ip]
		if !exists || time.Since(data.lastReset) > window {
			data = &clientData{count: 0, lastReset: time.Now()}
			clients[ip] = data
		}
		if data.count >= maxRequests {
			mu.Unlock()
			c.JSON(429, gin.H{"error": "Too many requests. Please try again later."})
			c.Abort()
			return
		}
		data.count++
		mu.Unlock()
		c.Next()
	}
}

// CacheMiddleware добавляет информацию о кешировании в заголовки ответа
func CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Добавляем заголовки для кеширования
		c.Header("Cache-Control", "public, max-age=300") // 5 минут
		c.Header("Vary", "Authorization, X-Request-ID")

		c.Next()

		// Добавляем заголовок с информацией о кешировании
		if c.Writer.Status() == 200 {
			c.Header("X-Cache-Status", "MISS")
		}
	}
}
