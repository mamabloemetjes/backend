package admin

import (
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (ar *AdminRoutesManager) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	productId := chi.URLParam(r, "id")
	if productId == "" {
		gecho.BadRequest(w, gecho.WithMessage("Please select a product to delete"), gecho.Send())
		return
	}
	total, err := database.Query[tables.Product](ar.db).Where("id", productId).Delete(r.Context())
	if err != nil {
		ar.logger.Error("Failed to delete product", gecho.Field("error", err), gecho.Field("product_id", productId))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to delete product. Please try again"), gecho.Send())
		return
	}
	if total == 0 {
		gecho.NotFound(w, gecho.WithMessage("Product not found"), gecho.Send())
		return
	}

	// Invalidate product caches asynchronously
	productUUID, parseErr := uuid.Parse(productId)
	if parseErr == nil {
		go func(id uuid.UUID) {
			if cacheErr := ar.cacheService.InvalidateProductCaches(id); cacheErr != nil {
				ar.logger.Warn("Failed to invalidate product caches after deletion",
					gecho.Field("error", cacheErr),
					gecho.Field("product_id", id),
				)
			}
		}(productUUID)
	}

	gecho.Success(w,
		gecho.WithMessage("Product deleted successfully"),
		gecho.WithData(map[string]int{"deleted_count": total}),
		gecho.Send(),
	)
}
