package admin

import (
	"mamabloemetjes_server/structs/tables"
	"net/http"
	"strconv"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ListOrders returns a paginated list of orders with optional filtering
func (ar *AdminRoutesManager) ListOrders(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	// Pagination
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	// Filters
	var status *tables.OrderStatus
	if statusStr := query.Get("status"); statusStr != "" {
		s := tables.OrderStatus(statusStr)
		status = &s
	}

	var paymentStatus *tables.PaymentStatus
	if paymentStatusStr := query.Get("payment_status"); paymentStatusStr != "" {
		ps := tables.PaymentStatus(paymentStatusStr)
		paymentStatus = &ps
	}

	// Get orders from service
	orders, total, err := ar.orderService.GetAllOrders(r.Context(), status, paymentStatus, pageSize, offset)
	if err != nil {
		ar.logger.Error("Failed to get orders",
			gecho.Field("error", err),
			gecho.Field("page", page),
			gecho.Field("page_size", pageSize))
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingOrders"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	// Calculate pagination metadata
	totalPages := (total + pageSize - 1) / pageSize

	gecho.Success(w,
		gecho.WithMessage("success.order.ordersFetched"),
		gecho.WithData(map[string]interface{}{
			"orders": orders,
			"pagination": map[string]interface{}{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": totalPages,
			},
		}),
		gecho.Send(),
	)
}

// GetOrderDetails returns detailed information about a specific order
func (ar *AdminRoutesManager) GetOrderDetails(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL
	orderIdStr := chi.URLParam(r, "id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidOrderId"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	// Get order
	order, err := ar.orderService.GetOrderById(r.Context(), orderId)
	if err != nil {
		ar.logger.Error("Failed to get order",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		gecho.NotFound(w,
			gecho.WithMessage("error.order.notFound"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	// Get order lines
	orderLines, err := ar.orderService.GetOrderLinesByOrderId(r.Context(), orderId)
	if err != nil {
		ar.logger.Error("Failed to get order lines",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingOrderLines"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	// Get address
	address, err := ar.orderService.GetAddressById(r.Context(), order.AddressId)
	if err != nil {
		ar.logger.Error("Failed to get address",
			gecho.Field("error", err),
			gecho.Field("address_id", order.AddressId))
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingAddress"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	// Calculate total
	var total uint64
	for _, line := range orderLines {
		total += line.LineTotal
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.orderDetailsFetched"),
		gecho.WithData(map[string]interface{}{
			"order":       order,
			"order_lines": orderLines,
			"address":     address,
			"total":       total,
		}),
		gecho.Send(),
	)
}
