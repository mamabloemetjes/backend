package services

import (
	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/database"
	"github.com/mamabloemetjes/backend/structs"
)

type ServiceManager struct {
	AuthService       *AuthService
	EmailService      *EmailService
	CacheService      *CacheService
	HealthService     *HealthService
	ProductService    *ProductService
	FilesystemService *FileService
	OrderService      *OrderService
}

func NewServiceManager(logger *gecho.Logger, cfg *structs.Config, db *database.DB) *ServiceManager {
	authService := NewAuthService(cfg, logger, db)
	cacheService := NewCacheService(logger, cfg)
	emailService := NewEmailService(logger, cfg, db)
	healthService := NewHealthService(logger, db)
	filesystemService := NewFileService(logger, cfg)
	productService := NewProductService(logger, db, cacheService, filesystemService)
	orderService := NewOrderService(logger, cfg, db, productService, emailService)

	return &ServiceManager{
		AuthService:       authService,
		EmailService:      emailService,
		CacheService:      cacheService,
		HealthService:     healthService,
		ProductService:    productService,
		FilesystemService: filesystemService,
		OrderService:      orderService,
	}
}
