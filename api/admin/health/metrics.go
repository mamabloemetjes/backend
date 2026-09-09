package health

import "github.com/prometheus/client_golang/prometheus"

var (
	HttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "api",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request latency",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	HttpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "api",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)
