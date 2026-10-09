package orders

import (
	"errors"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"
	"strings"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
)

func (orm *OrderRoutesManager) CreateOrder(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.OrderRequest](r)
	if err != nil {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidRequestBody"),
			gecho.Send(),
		)
		return
	}

	// Default country to NL if not provided
	if body.Country == "" {
		body.Country = "NL"
	}

	// Check if user is authenticated (optional - for linking orders to user accounts)
	var userId *uuid.UUID
	if claims, err := lib.ExtractClaims(r); err == nil {
		userId = &claims.Sub
	}

	// Create order using service (handles validation, pricing snapshots, email sending)
	idempotencyKey := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if len(idempotencyKey) > 128 {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidIdempotencyKey"),
			gecho.Send(),
		)
		return
	}

	order, err := orm.orderService.CreateOrderFromRequest(r.Context(), body, userId, idempotencyKey)
	if err != nil {
		orm.logger.Error("Failed to create order", gecho.Field("error", err))
		// Check for specific business logic errors
		if errors.Is(err, lib.ErrProductUnavailable) {
			gecho.BadRequest(w,
				gecho.WithMessage("error.order.productUnavailable"),
				gecho.Send(),
			)
			return
		}

		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.creationFailed"),
			gecho.Send(),
		)
		return
	}

	// Send confirmation email to customer and shop owner
	if !order.Existing {
		go func() {
			err := orm.emailService.SendOrderConfirmationEmail(order.Order.Email, order.Order.Name, order.Order.OrderNumber, order.OrderLines, order.Address)
			if err != nil {
				orm.logger.Error("Failed to send order confirmation email",
					gecho.Field("error", err),
					gecho.Field("email", order.Order.Email),
					gecho.Field("order_number", order.Order.OrderNumber),
				)
			} else {
				orm.logger.Info("Order confirmation email sent",
					gecho.Field("order_number", order.Order.OrderNumber))
			}
		}()
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.created"),
		gecho.WithData(map[string]any{
			"order_number": order.Order.OrderNumber,
			"order_id":     order.Order.Id,
			"status":       order.Order.Status,
		}),
		gecho.Send(),
	)
}
