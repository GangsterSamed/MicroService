package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	CacheHitsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"service", "method"},
	)
	CacheMissesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"service", "method"},
	)
)

func init() {
	prometheus.MustRegister(CacheHitsTotal)
	prometheus.MustRegister(CacheMissesTotal)
}

// RecordCacheMetrics записывает метрики кеширования
func RecordCacheMetrics(service, method string, hit bool) {
	labels := prometheus.Labels{
		"service": service,
		"method":  method,
	}

	if hit {
		CacheHitsTotal.With(labels).Inc()
	} else {
		CacheMissesTotal.With(labels).Inc()
	}
}
