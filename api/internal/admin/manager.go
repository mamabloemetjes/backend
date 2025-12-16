package admin

import (
	"mamabloemetjes_server/api/middleware"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

type AdminRoutesManager struct {
	logger         *gecho.Logger
	db             *database.DB
	authService    *services.AuthService
	productService *services.ProductService
	cfg            *structs.Config
	mw             *middleware.Middleware
}

func NewAuthRoutesManager(cfg *structs.Config, logger *gecho.Logger, db *database.DB, mw *middleware.Middleware) *AdminRoutesManager {
	return &AdminRoutesManager{
		logger:         logger,
		db:             db,
		cfg:            cfg,
		mw:             mw,
		productService: services.NewProductService(logger, db),
		authService:    services.NewAuthService(cfg, logger, db),
	}
}

func (ar *AdminRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(ar.mw.AdminAuthMiddleware)
		r.Get("/products", ar.ListAllProducts)
		r.Post("/products", ar.CreateProduct)
		r.Put("/products/{id}", ar.UpdateProducts)
		r.Put("/products/{id}/stock", ar.UpdateProductsStock)
		r.Delete("/products/{id}", ar.DeleteProduct)
	})
}
