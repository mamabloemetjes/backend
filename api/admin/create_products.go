package admin

import (
	"net/http"

	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/api/response"
	"github.com/mamabloemetjes/backend/lib"
	"github.com/mamabloemetjes/backend/structs/tables"
)

func (ar *AdminRoutesManager) CreateProduct(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[tables.Product](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate body", err)
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.products.checkProductInformation")))
		return
	}

	body.SKU, err = lib.GenerateSKU(body.Name, 5)
	ar.logger.Debug("Generated SKU", gecho.Field("sku", body.SKU))
	if err != nil {
		ar.logger.Error("Failed to generate SKU", gecho.Field("error", err))
		response.Send(gecho.InternalServerError(w, gecho.WithMessage("error.products.unableToCreate")))
		return
	}

	ar.logger.Debug("CreateProduct request received",
		gecho.Field("product_name", body.Name),
		gecho.Field("images_count", len(body.Images)),
	)
	if len(body.Images) > 0 {
		for i, img := range body.Images {
			ar.logger.Debug("Image received",
				gecho.Field("index", i),
				gecho.Field("name", img.Name),
				gecho.Field("alt_text", img.AltText),
				gecho.Field("is_primary", img.IsPrimary),
			)
		}
	}

	newProduct, err := ar.productService.CreateProduct(r.Context(), body)
	if err != nil {
		ar.logger.Error("Failed to create product", gecho.Field("error", err))
		response.Send(gecho.InternalServerError(w, gecho.WithMessage("error.products.unableToCreate")))
		return
	}

	response.Send(gecho.Success(w,
		gecho.WithData(newProduct),
		gecho.WithMessage("success.products.created"),
	))
}
