package admin

import (
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

func (ar *AdminRoutesManager) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productId := chi.URLParam(r, "id")
	if productId == "" {
		gecho.BadRequest(w, gecho.WithMessage("error.products.selectProductToDelete"), gecho.Send())
		return
	}
	total, err := ar.productService.DeleteProductByID(r.Context(), productId)
	if err != nil {
		ar.logger.Error("Failed to delete product", gecho.Field("error", err), gecho.Field("product_id", productId))
		gecho.InternalServerError(w, gecho.WithMessage("error.products.unableToDelete"), gecho.Send())
		return
	}
	if total == 0 {
		gecho.NotFound(w, gecho.WithMessage("error.products.notFound"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithMessage("success.products.deleted"),
		gecho.WithData(map[string]int{"deleted_count": total}),
		gecho.Send(),
	)
}
