package orders

import (
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/MonkyMars/gecho"
)

// GetMyOrders returns all orders for the authenticated user
func (orm *OrderRoutesManager) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	claims, err := lib.ExtractClaims(r)
	if err != nil {
		orm.logger.Warn("Failed to extract claims in GetMyOrders", gecho.Field("error", err))
		gecho.Unauthorized(w,
			gecho.WithMessage("error.auth.invalidOrMissingAccessToken"),
			gecho.Send(),
		)
		return
	}

	orm.logger.Info("Fetching orders for user", gecho.Field("user_id", claims.Sub))

	// Get orders for user
	orders, err := orm.orderService.GetOrdersByUserId(r.Context(), claims.Sub)
	if err != nil {
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.fetchingOrders"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.ordersFetched"),
		gecho.WithData(map[string]any{
			"orders": orders,
			"count":  len(orders),
		}),
		gecho.Send(),
	)
}
