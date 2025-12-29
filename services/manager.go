package services

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
)

type ServiceManager struct {
	AuthService    *AuthService
	EmailService   *EmailService
	CacheService   *CacheService
	HealthService  *HealthService
	ProductService *ProductService
	OrderService   *OrderService
}

func NewServiceManager(logger *gecho.Logger, cfg *structs.Config, db *database.DB) *ServiceManager {
	authService := NewAuthService(cfg, logger, db)
	cacheService := NewCacheService(logger, cfg)
	emailService := NewEmailService(logger, cfg, db)
	healthService := NewHealthService(logger, db)
	productService := NewProductService(logger, db, cacheService)
	orderService := NewOrderService(logger, cfg, db, productService)

	return &ServiceManager{
		AuthService:    authService,
		EmailService:   emailService,
		CacheService:   cacheService,
		HealthService:  healthService,
		ProductService: productService,
		OrderService:   orderService,
	}
}
