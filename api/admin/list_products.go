package admin

import (
	"mamabloemetjes_server/handling"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AdminRoutesManager) ListAllProducts(w http.ResponseWriter, r *http.Request) {
	opts, err := handling.ParseProductListOptions(r)
	if err != nil {
		ar.logger.Warn("Failed to parse product list options", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("error.invalidQueryParameters"), gecho.Send())
		return
	}
	products, err := ar.productService.GetAllProducts(r.Context(), opts)
	if err != nil {
		ar.logger.Error("Failed to list products", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("error.products.failedToList"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithData(products),
		gecho.WithMessage("success.products.retrieved"),
		gecho.Send(),
	)
}
