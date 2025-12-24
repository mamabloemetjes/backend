package admin

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AdminRoutesManager) CreateProduct(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[tables.Product](r)
	if err != nil {
		ar.logger.Debug("Failed to extract and validate body", err)
		gecho.BadRequest(w, gecho.WithMessage("error.products.checkProductInformation"), gecho.Send())
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

	newProduct, err := ar.productService.CreateProduct(r.Context(), body)
	if err != nil {
		ar.logger.Error("Failed to create product", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("error.products.unableToCreate"), gecho.Send())
		return
	}

	gecho.Success(w,
		gecho.WithData(newProduct),
		gecho.WithMessage("success.products.created"),
		gecho.Send(),
	)
}
