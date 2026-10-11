package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/mamabloemetjes/backend/api/admin"
	"github.com/mamabloemetjes/backend/api/auth"
	"github.com/mamabloemetjes/backend/api/debug"
	"github.com/mamabloemetjes/backend/api/fs"
	"github.com/mamabloemetjes/backend/api/health"
	"github.com/mamabloemetjes/backend/api/orders"
	"github.com/mamabloemetjes/backend/api/products"
)

type routerManager struct {
	productRoutes    *products.ProductRoutesManager
	healthRoutes     *health.HealthRoutesManager
	authRoutes       *auth.AuthRoutesManager
	adminRoutes      *admin.AdminRoutesManager
	orderRoutes      *orders.OrderRoutesManager
	filesystemRoutes *fs.FileRoutesManager
	debugRoutes      *debug.DebugRoutesManager
}

func NewRouterManager(
	productRoutes *products.ProductRoutesManager,
	healthRoutes *health.HealthRoutesManager,
	authRoutes *auth.AuthRoutesManager,
	adminRoutes *admin.AdminRoutesManager,
	ordersRoutes *orders.OrderRoutesManager,
	filesystemRoutes *fs.FileRoutesManager,
	debugRoutes *debug.DebugRoutesManager,
) *routerManager {
	return &routerManager{
		productRoutes:    productRoutes,
		healthRoutes:     healthRoutes,
		authRoutes:       authRoutes,
		adminRoutes:      adminRoutes,
		filesystemRoutes: filesystemRoutes,
		debugRoutes:      debugRoutes,
		orderRoutes:      ordersRoutes,
	}
}

func (rm *routerManager) RegisterRoutes(r chi.Router) {
	rm.productRoutes.RegisterRoutes(r)
	rm.healthRoutes.RegisterRoutes(r)
	rm.authRoutes.RegisterRoutes(r)
	rm.adminRoutes.RegisterRoutes(r)
	rm.orderRoutes.RegisterRoutes(r)
	rm.debugRoutes.RegisterRoutes(r)
	rm.filesystemRoutes.RegisterRoutes(r)
}
