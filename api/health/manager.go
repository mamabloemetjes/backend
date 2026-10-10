package health

import (
	"crypto/subtle"
	"mamabloemetjes_server/config"
	"mamabloemetjes_server/services"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HealthRoutesManager struct {
	healthService *services.HealthService
}

func NewHealthRoutesManager(healthService *services.HealthService) *HealthRoutesManager {
	return &HealthRoutesManager{
		healthService: healthService,
	}
}

func (hrm *HealthRoutesManager) RegisterRoutes(r chi.Router) {
	r.Get("/health/server", hrm.GetServerHealth)
	r.Get("/health/database", hrm.internalOnly(hrm.GetDatabaseHealth))

	// Prometheus metrics endpoint
	r.Get("/metrics", hrm.internalOnly(promhttp.Handler().ServeHTTP))
	// Register Prometheus metrics
	prometheus.MustRegister(HttpDuration, HttpRequests)
	services.RegisterCacheMetrics()
}

func (hrm *HealthRoutesManager) internalOnly(next http.HandlerFunc) http.HandlerFunc {
	token := config.GetConfig().Server.MonitoringToken
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthorizedMonitoringRequest(r.Header.Get("Authorization"), token) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func isAuthorizedMonitoringRequest(authorization, token string) bool {
	provided := strings.TrimPrefix(authorization, "Bearer ")
	return token != "" && provided != authorization &&
		len(provided) == len(token) &&
		subtle.ConstantTimeCompare([]byte(provided), []byte(token)) == 1
}
