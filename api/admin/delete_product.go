package admin

import (
	"mamabloemetjes_server/api/response"
	"mamabloemetjes_server/lib"
	"net/http"

	"github.com/google/uuid"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

func (ar *AdminRoutesManager) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productIDStr := chi.URLParam(r, "id")
	if productIDStr == "" {
		ar.logger.Warn("Product ID is missing in the request")
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.products.missingProductID")))
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.products.invalidProductID")))
		return
	}

	err = ar.productService.DeleteProduct(r.Context(), productID)
	if err != nil {
		ar.logger.Error("Failed to delete product", gecho.Field("error", err))
		if lib.IsNotFound(err) {
			response.Send(gecho.NotFound(w, gecho.WithMessage("error.products.notFound")))
			return
		}
		response.Send(gecho.InternalServerError(w, gecho.WithMessage("error.products.unableToDelete")))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.products.deleted"),
	))
}
