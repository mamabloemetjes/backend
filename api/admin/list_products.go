package admin

import (
	"mamabloemetjes_server/api/response"
	"mamabloemetjes_server/handling"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AdminRoutesManager) ListAllProducts(w http.ResponseWriter, r *http.Request) {
	opts, err := handling.ParseProductListOptions(r)
	if err != nil {
		ar.logger.Warn("Failed to parse product list options", gecho.Field("error", err))
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.invalidQueryParameters")))
		return
	}
	products, err := ar.productService.GetAllProducts(r.Context(), opts)
	if err != nil {
		ar.logger.Error("Failed to list products", gecho.Field("error", err))
		response.Send(gecho.InternalServerError(w, gecho.WithMessage("error.products.failedToList")))
		return
	}
	response.Send(gecho.Success(w,
		gecho.WithData(products),
		gecho.WithMessage("success.products.retrieved"),
	))
}
