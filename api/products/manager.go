package products

import (
	"mamabloemetjes_server/services"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type ProductRoutesManager struct {
	logger         *gecho.Logger
	productService *services.ProductService
}

func NewProductRoutesManager(
	logger *gecho.Logger,
	productService *services.ProductService,
) *ProductRoutesManager {
	return &ProductRoutesManager{
		logger:         logger,
		productService: productService,
	}
}

func (prm *ProductRoutesManager) RegisterRoutes(r chi.Router) {
	// Register product-related routes here
	r.Get("/products", prm.FetchAllProducts)
	r.Get("/products/{id}", prm.FetchProductByID)
	r.Get("/products/active", prm.FetchActiveProducts)
	r.Get("/products/count", prm.GetProductCount)
}
