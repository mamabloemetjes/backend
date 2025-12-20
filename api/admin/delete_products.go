package admin

import (
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

func (ar *AdminRoutesManager) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productId := chi.URLParam(r, "id")
	if productId == "" {
		gecho.BadRequest(w, gecho.WithMessage("Please select a product to delete"), gecho.Send())
		return
	}
	total, err := ar.productService.DeleteProductByID(r.Context(), productId)
	if err != nil {
		ar.logger.Error("Failed to delete product", gecho.Field("error", err), gecho.Field("product_id", productId))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to delete product. Please try again"), gecho.Send())
		return
	}
	if total == 0 {
		gecho.NotFound(w, gecho.WithMessage("Product not found"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithMessage("Product deleted successfully"),
		gecho.WithData(map[string]int{"deleted_count": total}),
		gecho.Send(),
	)
}
