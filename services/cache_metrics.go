package services

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	cacheMetricsOnce sync.Once
	CacheOperations  = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mamabloemetjes",
			Subsystem: "cache",
			Name:      "operations_total",
			Help:      "Total Redis cache operations by operation and outcome.",
		},
		[]string{"operation", "outcome"},
	)
	CacheOperationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "mamabloemetjes",
			Subsystem: "cache",
			Name:      "operation_duration_seconds",
			Help:      "Redis cache operation duration.",
		},
		[]string{"operation"},
	)
	CacheInvalidations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "mamabloemetjes",
			Subsystem: "cache",
			Name:      "invalidations_total",
			Help:      "Total cache invalidations by outcome.",
		},
		[]string{"outcome"},
	)
)

func RegisterCacheMetrics() {
	cacheMetricsOnce.Do(func() {
		prometheus.MustRegister(CacheOperations, CacheOperationDuration, CacheInvalidations)
	})
}

func observeCacheOperation(operation string, err error, started time.Time) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	CacheOperations.WithLabelValues(operation, outcome).Inc()
	CacheOperationDuration.WithLabelValues(operation).Observe(time.Since(started).Seconds())
}
