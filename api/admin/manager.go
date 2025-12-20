package admin

import (
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/services"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type AdminRoutesManager struct {
	logger         *gecho.Logger
	productService *services.ProductService
	mw             *middleware.Middleware
}

func NewAdminRoutesManager(
	logger *gecho.Logger,
	productService *services.ProductService,
	mw *middleware.Middleware,
) *AdminRoutesManager {
	return &AdminRoutesManager{
		logger:         logger,
		productService: productService,
		mw:             mw,
	}
}

func (ar *AdminRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(ar.mw.AdminAuthMiddleware)
		r.Get("/products", ar.ListAllProducts)

		// Protected routes behind CSRF
		r.Group(func(r chi.Router) {
			r.Use(ar.mw.CSRFMiddleware())
			r.Post("/products", ar.CreateProduct)
			r.Put("/products/{id}", ar.UpdateProducts)
			r.Put("/products/{id}/stock", ar.UpdateProductsStock)
			r.Delete("/products/{id}", ar.DeleteProduct)
		})
	})
}
