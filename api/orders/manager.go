package orders

import (
	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"

	"github.com/mamabloemetjes/backend/api/middleware"
	"github.com/mamabloemetjes/backend/services"
)

type OrderRoutesManager struct {
	productService *services.ProductService
	orderService   *services.OrderService
	emailService   *services.EmailService
	middleware     *middleware.Middleware
	logger         *gecho.Logger
}

func NewOrderRoutesManager(productService *services.ProductService, orderService *services.OrderService, emailService *services.EmailService, middleware *middleware.Middleware, logger *gecho.Logger) *OrderRoutesManager {
	return &OrderRoutesManager{
		productService: productService,
		orderService:   orderService,
		emailService:   emailService,
		middleware:     middleware,
		logger:         logger,
	}
}

func (orm *OrderRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		r.With(orm.middleware.CSRFMiddleware()).Post("/create", orm.CreateOrder)
		r.Route("/", func(r chi.Router) {
			r.Use(orm.middleware.UserAuthMiddleware)
			r.Get("/my-orders", orm.GetMyOrders)         // Requires authentication
			r.Get("/my-orders/{id}", orm.GetMyOrderById) // Get specific order details
		})
	})
}
