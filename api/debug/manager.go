package debug

import (
	"github.com/go-chi/chi/v5"

	"github.com/mamabloemetjes/backend/config"
	"github.com/mamabloemetjes/backend/services"
)

type DebugRoutesManager struct {
	cacheService *services.CacheService
}

func NewDebugRoutesManager(cacheService *services.CacheService) *DebugRoutesManager {
	return &DebugRoutesManager{
		cacheService: cacheService,
	}
}

func (drm *DebugRoutesManager) RegisterRoutes(r chi.Router) {
	// Debug routes - only in non-production environments
	if !config.IsProduction() {
		r.Route("/debug", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Post("/cache/clear", drm.ClearCache)
			})
		})
	}
}
