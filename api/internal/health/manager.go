package health

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type HealthRoutesManager struct {
	logger        *gecho.Logger
	db            *database.DB
	healthService *services.HealthService
}

func NewHealthRoutesManager(logger *gecho.Logger, db *database.DB) *HealthRoutesManager {
	return &HealthRoutesManager{
		logger:        logger,
		db:            db,
		healthService: services.NewHealthService(logger, db),
	}
}

func (hrm *HealthRoutesManager) RegisterRoutes(r chi.Router) {
	r.Get("/health/server", hrm.GetServerHealth)
	r.Get("/health/database", hrm.GetDatabaseHealth)
}
