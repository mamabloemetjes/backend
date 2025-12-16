package admin

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"net/http"
	"strings"

	"github.com/MonkyMars/gecho"
)

func (ar *AdminRoutesManager) CreateProduct(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[tables.Product](r)
	if err != nil {
		ar.logger.Debug("Failed to extract and validate body", err)
		gecho.BadRequest(w, gecho.WithMessage("Invalid request body"), gecho.Send())
		return
	}

	// Normalize colors to lowercase
	if body.Colors != nil && len(body.Colors) > 0 {
		for i, color := range body.Colors {
			body.Colors[i] = structs.Color(strings.ToLower(string(color)))
		}
	}

	newProduct, err := ar.productService.CreateProduct(r.Context(), body)
	if err != nil {
		ar.logger.Error("Failed to create product", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Failed to create product"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithData(newProduct),
		gecho.WithMessage("Product created successfully"),
		gecho.Send(),
	)
}
