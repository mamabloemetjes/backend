package api

import (
	"mamabloemetjes_server/api/admin"
	"mamabloemetjes_server/api/auth"
	"mamabloemetjes_server/api/debug"
	"mamabloemetjes_server/api/health"
	"mamabloemetjes_server/api/orders"
	"mamabloemetjes_server/api/products"

	"github.com/go-chi/chi/v5"
)

type routerManager struct {
	productRoutes *products.ProductRoutesManager
	healthRoutes  *health.HealthRoutesManager
	authRoutes    *auth.AuthRoutesManager
	adminRoutes   *admin.AdminRoutesManager
	orderRoutes   *orders.OrderRoutesManager
	debugRoutes   *debug.DebugRoutesManager
}

func NewRouterManager(
	productRoutes *products.ProductRoutesManager,
	healthRoutes *health.HealthRoutesManager,
	authRoutes *auth.AuthRoutesManager,
	adminRoutes *admin.AdminRoutesManager,
	ordersRoutes *orders.OrderRoutesManager,
	debugRoutes *debug.DebugRoutesManager,
) *routerManager {
	return &routerManager{
		productRoutes: productRoutes,
		healthRoutes:  healthRoutes,
		authRoutes:    authRoutes,
		adminRoutes:   adminRoutes,
		debugRoutes:   debugRoutes,
		orderRoutes:   ordersRoutes,
	}
}

func (rm *routerManager) RegisterRoutes(r chi.Router) {
	rm.productRoutes.RegisterRoutes(r)
	rm.healthRoutes.RegisterRoutes(r)
	rm.authRoutes.RegisterRoutes(r)
	rm.adminRoutes.RegisterRoutes(r)
	rm.orderRoutes.RegisterRoutes(r)
	rm.debugRoutes.RegisterRoutes(r)
}
