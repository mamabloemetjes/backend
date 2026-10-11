package admin

import (
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/mamabloemetjes/backend/api/response"
	"github.com/mamabloemetjes/backend/lib"
	"github.com/mamabloemetjes/backend/structs/tables"
)

type AttachPaymentLinkRequest struct {
	PaymentLink string `json:"payment_link" validate:"required,url"`
}

type UpdateOrderStatusRequest struct {
	Status tables.OrderStatus `json:"status" validate:"required,oneof=pending paid processing shipped delivered cancelled refunded"`
}

// AttachPaymentLink attaches a Tikkie payment link to an order and sends email to customer
func (ar *AdminRoutesManager) AttachPaymentLink(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL
	orderIdStr := chi.URLParam(r, "id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidOrderId"),
		))
		return
	}

	// Extract and validate request body
	body, err := lib.ExtractAndValidateBody[AttachPaymentLinkRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate payment link request",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId),
		)
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidRequestBody"),
		))
		return
	}

	// Attach payment link (service handles email sending)
	err = ar.orderService.AttachPaymentLink(r.Context(), orderId, body.PaymentLink)
	if err != nil {
		ar.logger.Error("Failed to attach payment link",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.attachingPaymentLink"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.order.paymentLinkAttached"),
	))
}

// MarkOrderAsPaid marks an order as paid
func (ar *AdminRoutesManager) MarkOrderAsPaid(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL
	orderIdStr := chi.URLParam(r, "id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidOrderId"),
		))
		return
	}

	// Mark order as paid
	err = ar.orderService.MarkOrderAsPaid(r.Context(), orderId)
	if err != nil {
		ar.logger.Error("Failed to mark order as paid",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.markingAsPaid"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.order.markedAsPaid"),
	))
}

// UpdateOrderStatus updates the status of an order
func (ar *AdminRoutesManager) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL
	orderIdStr := chi.URLParam(r, "id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidOrderId"),
		))
		return
	}

	// Parse status from form
	body, err := lib.ExtractAndValidateBody[UpdateOrderStatusRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate order status request",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId),
		)
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidRequestBody"),
		))
		return
	}

	if body.Status == "" {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("Body incomplete"),
		))
		return
	}

	// Update order status
	err = ar.orderService.UpdateOrderStatus(r.Context(), orderId, body.Status)
	if err != nil {
		ar.logger.Error("Failed to update order status",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId),
		)

		// Check if it's a validation error
		if err.Error() == "invalid status transition" {
			response.Send(gecho.BadRequest(w,
				gecho.WithMessage("error.order.invalidStatusTransition"),
			))
			return
		}
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.updatingStatus"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.order.statusUpdated"),
	))
}

// DeleteOrder soft deletes an order
func (ar *AdminRoutesManager) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	// Get order ID from URL
	orderIdStr := chi.URLParam(r, "id")
	orderId, err := uuid.Parse(orderIdStr)
	if err != nil {
		response.Send(gecho.BadRequest(w,
			gecho.WithMessage("error.order.invalidOrderId"),
		))
		return
	}

	// Soft delete order
	err = ar.orderService.SoftDeleteOrder(r.Context(), orderId)
	if err != nil {
		ar.logger.Error("Failed to delete order",
			gecho.Field("error", err),
			gecho.Field("order_id", orderId))
		response.Send(gecho.InternalServerError(w,
			gecho.WithMessage("error.order.deletingOrder"),
		))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.order.deleted"),
	))
}
