package admin

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

func (ar *AdminRoutesManager) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productId := chi.URLParam(r, "id")
	if productId == "" {
		gecho.BadRequest(w, gecho.WithMessage("Product ID is required"), gecho.Send())
		return
	}
	total, err := database.Query[tables.Product](ar.db).Where("id", productId).Delete(r.Context())
	if err != nil {
		ar.logger.Error("Failed to delete product", gecho.Field("error", err), gecho.Field("product_id", productId))
		gecho.InternalServerError(w, gecho.WithMessage("Failed to delete product"), gecho.Send())
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
