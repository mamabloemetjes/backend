package health

import (
	"mamabloemetjes_server/services"

	"github.com/go-chi/chi/v5"
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
	r.Get("/health/database", hrm.GetDatabaseHealth)
}
