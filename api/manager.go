package api

import (
	"mamabloemetjes_server/api/internal/admin"
	"mamabloemetjes_server/api/internal/auth"
	"mamabloemetjes_server/api/internal/health"
	"mamabloemetjes_server/api/internal/products"
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type routerManager struct {
	productRoutes *products.ProductRoutesManager
	healthRoutes  *health.HealthRoutesManager
	authRoutes    *auth.AuthRoutesManager
	adminRoutes   *admin.AdminRoutesManager
}

func NewRouterManager(
	logger *gecho.Logger,
	db *database.DB,
	cfg *structs.Config,
	mw *middleware.Middleware,
) *routerManager {
	return &routerManager{
		productRoutes: products.NewProductRoutesManager(logger, db),
		healthRoutes:  health.NewHealthRoutesManager(logger, db),
		authRoutes:    auth.NewAuthRoutesManager(cfg, logger, db, mw),
		adminRoutes:   admin.NewAuthRoutesManager(cfg, logger, db, mw),
	}
}

func (rm *routerManager) RegisterRoutes(r chi.Router) {
	rm.productRoutes.RegisterRoutes(r)
	rm.healthRoutes.RegisterRoutes(r)
	rm.authRoutes.RegisterRoutes(r)
	rm.adminRoutes.RegisterRoutes(r)
}
