package orders

import (
	"mamabloemetjes_server/api/response"
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetMyOrders returns all orders for the authenticated user
func (orm *OrderRoutesManager) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	claims, err := lib.ExtractClaims(r)
	if err != nil {
		orm.logger.Warn("Failed to extract claims in GetMyOrders", gecho.Field("error", err))
		response.Send(gecho.Unauthorized(w,
			gecho.WithMessage("error.auth.invalidOrMissingAccessToken"),
		))
		return
	}

	orm.logger.Info("Fetching orders for user", gecho.Field("user_id", claims.Sub))

	// Get orders for user
	orders, err := orm.orderService.GetOrdersByUserId(r.Context(), claims.Sub)
	if err != nil {
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingOrders"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.order.ordersFetched"),
		gecho.WithData(map[string]any{
			"orders": orders,
			"count":  len(orders),
		}),
	))
}

// GetMyOrderById returns detailed information about a specific order for the authenticated user
func (orm *OrderRoutesManager) GetMyOrderById(w http.ResponseWriter, r *http.Request) {
	claims, err := lib.ExtractClaims(r)
	if err != nil {
		orm.logger.Warn("Failed to extract claims in GetMyOrderById", gecho.Field("error", err))
		response.Send(gecho.Unauthorized(w,
			gecho.WithMessage("error.auth.invalidOrMissingAccessToken"),
		))
		return
	}

	// Get order ID from URL
	orderIdStr := chi.URLParam(r, "id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		orm.logger.Warn("Invalid order ID format", gecho.Field("order_id", orderIdStr))
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidOrderId"),
		))
		return
	}

	orm.logger.Info("Fetching order details for user", gecho.Field("user_id", claims.Sub), gecho.Field("order_id", orderId))

	// Get order
	order, err := orm.orderService.GetOrderById(r.Context(), orderId)
	if err != nil {
		orm.logger.Error("Failed to get order", gecho.Field("error", err), gecho.Field("order_id", orderId))
		response.Send(gecho.NotFound(w,
			gecho.WithMessage("error.order.notFound"),
		))
		return
	}

	// Get address first to verify ownership
	address, err := orm.orderService.GetAddressById(r.Context(), order.AddressId)
	if err != nil {
		orm.logger.Error("Failed to get address",
			gecho.Field("error", err),
			gecho.Field("address_id", order.AddressId))
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingAddress"),
		))
		return
	}

	// Verify that the order belongs to this user (via address)
	if address.UserId == nil || *address.UserId != claims.Sub {
		orm.logger.Warn("User attempted to access order they don't own",
			gecho.Field("user_id", claims.Sub),
			gecho.Field("order_id", orderId),
			gecho.Field("address_user_id", address.UserId),
		)
		response.Send(gecho.Forbidden(w,
			gecho.WithMessage("error.auth.accessDenied"),
		))
		return
	}

	// Get order lines
	orderLines, err := orm.orderService.GetOrderLinesByOrderId(r.Context(), orderId)
	if err != nil {
		orm.logger.Error("Failed to get order lines",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingOrderLines"),
		))
		return
	}

	// Calculate total
	var total uint64
	for _, line := range orderLines {
		total += line.LineTotal
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.order.orderDetailsFetched"),
		gecho.WithData(map[string]interface{}{
			"order":       order,
			"order_lines": orderLines,
			"address":     address,
			"total":       total,
		}),
	))
}
