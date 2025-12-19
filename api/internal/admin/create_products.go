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
		gecho.BadRequest(w, gecho.WithMessage("Please check the product information and try again"), gecho.Send())
		return
	}

	// Debug log to check if images are received
	ar.logger.Debug("CreateProduct request received",
		gecho.Field("product_name", body.Name),
		gecho.Field("images_count", len(body.Images)),
	)
	if len(body.Images) > 0 {
		for i, img := range body.Images {
			ar.logger.Debug("Image received",
				gecho.Field("index", i),
				gecho.Field("url", img.URL),
				gecho.Field("alt_text", img.AltText),
				gecho.Field("is_primary", img.IsPrimary),
			)
		}
	}

	// Normalize colors to lowercase
	if len(body.Colors) > 0 {
		for i, color := range body.Colors {
			body.Colors[i] = structs.Color(strings.ToLower(string(color)))
		}
	}

	newProduct, err := ar.productService.CreateProduct(r.Context(), body)
	if err != nil {
		ar.logger.Error("Failed to create product", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to create product. Please try again"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithData(newProduct),
		gecho.WithMessage("Product created successfully"),
		gecho.Send(),
	)
}
