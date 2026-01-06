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
	orderService   *services.OrderService
	mw             *middleware.Middleware
}

func NewAdminRoutesManager(
	logger *gecho.Logger,
	productService *services.ProductService,
	orderService *services.OrderService,
	mw *middleware.Middleware,
) *AdminRoutesManager {
	return &AdminRoutesManager{
		logger:         logger,
		productService: productService,
		orderService:   orderService,
		mw:             mw,
	}
}

func (ar *AdminRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(ar.mw.UserAuthMiddleware)
		r.Use(ar.mw.AdminAuthMiddleware)
		r.Get("/products", ar.ListAllProducts)

		// Order management routes
		r.Get("/orders", ar.ListOrders)
		r.Get("/orders/{id}", ar.GetOrderDetails)

		// Protected routes behind CSRF
		r.Group(func(r chi.Router) {
			r.Use(ar.mw.CSRFMiddleware())
			r.Post("/products", ar.CreateProduct)
			r.Put("/products", ar.UpdateProducts)

			// Order update routes
			r.Post("/orders/{id}/payment-link", ar.AttachPaymentLink)
			r.Post("/orders/{id}/mark-paid", ar.MarkOrderAsPaid)
			r.Put("/orders/{id}/status", ar.UpdateOrderStatus)
			r.Delete("/orders/{id}", ar.DeleteOrder)
		})
	})
}
