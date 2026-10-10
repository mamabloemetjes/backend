package admin

import (
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
		gecho.BadRequest(w, gecho.WithMessage("error.products.missingProductID"), gecho.Send())
		return
	}

	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		gecho.BadRequest(w, gecho.WithMessage("error.products.invalidProductID"), gecho.Send())
		return
	}

	err = ar.productService.DeleteProduct(r.Context(), productID)
	if err != nil {
		ar.logger.Error("Failed to delete product", gecho.Field("error", err))
		if lib.IsNotFound(err) {
			gecho.NotFound(w, gecho.WithMessage("error.products.notFound"), gecho.Send())
			return
		}
		gecho.InternalServerError(w, gecho.WithMessage("error.products.unableToDelete"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.products.deleted"),
		gecho.Send(),
	)
}
