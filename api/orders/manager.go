package orders

import (
	"mamabloemetjes_server/services"

	"github.com/go-chi/chi/v5"
)

type OrderRoutesManager struct {
	productService *services.ProductService
	orderService   *services.OrderService
}

func NewOrderRoutesManager(productService *services.ProductService, orderService *services.OrderService) *OrderRoutesManager {
	return &OrderRoutesManager{
		productService: productService,
		orderService:   orderService,
	}
}

func (orm *OrderRoutesManager) RegisterRoutes(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		r.Post("/create", orm.CreateOrder)
	})
}
