package products

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type ProductRoutesManager struct {
	logger         *gecho.Logger
	db             *database.DB
	productService *services.ProductService
	cacheService   *services.CacheService
	cfg            *structs.Config
}

func NewProductRoutesManager(logger *gecho.Logger, db *database.DB, cfg *structs.Config) *ProductRoutesManager {
	cacheService := services.NewCacheService(logger, cfg)
	return &ProductRoutesManager{
		logger:         logger,
		db:             db,
		cfg:            cfg,
		cacheService:   cacheService,
		productService: services.NewProductService(logger, db, cacheService),
	}
}

func (prm *ProductRoutesManager) RegisterRoutes(r chi.Router) {
	// Register product-related routes here
	r.Get("/products", prm.FetchAllProducts)
	r.Get("/products/{id}", prm.FetchProductByID)
	r.Get("/products/active", prm.FetchActiveProducts)
	r.Get("/products/count", prm.GetProductCount)
}
