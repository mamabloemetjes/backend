package admin

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AttachPaymentLinkRequest struct {
	PaymentLink string `json:"payment_link" validate:"required,url"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid processing shipped delivered cancelled refunded"`
}

// AttachPaymentLink attaches a Tikkie payment link to an order and sends email to customer
func (ar *AdminRoutesManager) AttachPaymentLink(w http.ResponseWriter, r *http.Request) {
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

	// Extract and validate request body
	body, err := lib.ExtractAndValidateBody[AttachPaymentLinkRequest](r)
	if err != nil {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidRequestBody"),
			gecho.WithData(err),
			gecho.Send(),
		)
		return
	}

	if body.PaymentLink == "" {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.paymentLinkRequired"),
			gecho.Send(),
		)
		return
	}

	// Attach payment link (service handles email sending)
	err = ar.orderService.AttachPaymentLink(r.Context(), orderId, body.PaymentLink)
	if err != nil {
		ar.logger.Error("Failed to attach payment link",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.attachingPaymentLink"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.paymentLinkAttached"),
		gecho.Send(),
	)
}

// MarkOrderAsPaid marks an order as paid
func (ar *AdminRoutesManager) MarkOrderAsPaid(w http.ResponseWriter, r *http.Request) {
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

	// Mark order as paid
	err = ar.orderService.MarkOrderAsPaid(r.Context(), orderId)
	if err != nil {
		ar.logger.Error("Failed to mark order as paid",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.markingAsPaid"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.markedAsPaid"),
		gecho.Send(),
	)
}

// UpdateOrderStatus updates the status of an order
func (ar *AdminRoutesManager) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
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

	// Parse status from form
	if err := r.ParseForm(); err != nil {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidRequestBody"),
			gecho.Send(),
		)
		return
	}

	statusStr := r.FormValue("status")
	if statusStr == "" {
		gecho.BadRequest(w,
			gecho.WithMessage("error.order.statusRequired"),
			gecho.Send(),
		)
		return
	}

	// Update order status
	err = ar.orderService.UpdateOrderStatus(r.Context(), orderId, tables.OrderStatus(statusStr))
	if err != nil {
		ar.logger.Error("Failed to update order status",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId),
			gecho.Field("status", statusStr))

		// Check if it's a validation error
		if err.Error() == "invalid status transition" {
			gecho.BadRequest(w,
				gecho.WithMessage("error.order.invalidStatusTransition"),
				gecho.WithData(map[string]string{"error": err.Error()}),
				gecho.Send(),
			)
			return
		}

		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.updatingStatus"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.statusUpdated"),
		gecho.Send(),
	)
}

// DeleteOrder soft deletes an order
func (ar *AdminRoutesManager) DeleteOrder(w http.ResponseWriter, r *http.Request) {
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

	// Soft delete order
	err = ar.orderService.SoftDeleteOrder(r.Context(), orderId)
	if err != nil {
		ar.logger.Error("Failed to delete order",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		gecho.InternalServerError(w,
			gecho.WithMessage("error.order.deletingOrder"),
			gecho.WithData(map[string]string{"error": err.Error()}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.order.deleted"),
		gecho.Send(),
	)
}
