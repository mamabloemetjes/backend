package admin

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/services"
	"mamabloemetjes_server/structs/tables"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
)

type UpdateProductRequest struct {
	Name        *string               `json:"name,omitempty"`
	SKU         *string               `json:"sku,omitempty"`
	Price       *uint64               `json:"price,omitempty"`
	Discount    *uint64               `json:"discount,omitempty"`
	Tax         *uint64               `json:"tax,omitempty"`
	Description *string               `json:"description,omitempty"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	Size        *string               `json:"size,omitempty"`
	Colors      []string              `json:"colors,omitempty"`
	ProductType *string               `json:"product_type,omitempty"`
	Stock       *uint16               `json:"stock,omitempty"`
	Images      []tables.ProductImage `json:"images,omitempty"`
}

type UpdateProductsRequest struct {
	Products map[string]UpdateProductRequest `json:"products"`
}

type UpdateProductsStockRequest struct {
	Stocks map[string]int `json:"stocks"`
}

func (ar *AdminRoutesManager) UpdateProducts(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[UpdateProductsRequest](r)
	if err != nil || len(body.Products) == 0 {
		ar.logger.Debug("Failed to extract and validate body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("error.products.checkProductInformation"), gecho.Send())
		return
	}

	totalErrors := make(map[string]string)
	for productID, updateReq := range body.Products {
		productUUID, parseErr := uuid.Parse(productID)
		if parseErr != nil {
			ar.logger.Error("Invalid product ID format", gecho.Field("error", parseErr), gecho.Field("product_id", productID))
			totalErrors[productID] = "error.products.invalidIdFormat"
			continue
		}

		// Create the service-level request from the API-level request
		serviceReq := &services.UpdateProductRequest{
			Name:        updateReq.Name,
			SKU:         updateReq.SKU,
			Price:       updateReq.Price,
			Discount:    updateReq.Discount,
			Tax:         updateReq.Tax,
			Description: updateReq.Description,
			IsActive:    updateReq.IsActive,
			Images:      updateReq.Images,
		}

		if err := ar.productService.UpdateProduct(r.Context(), productUUID, serviceReq); err != nil {
			ar.logger.Error("Failed to update product", gecho.Field("error", err), gecho.Field("product_id", productID))
			totalErrors[productID] = err.Error()
		}
	}

	if len(totalErrors) > 0 {
		gecho.InternalServerError(w,
			gecho.WithMessage("error.products.someFailedToUpdate"),
			gecho.WithData(map[string]any{"errors": totalErrors}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w, gecho.WithMessage("success.products.updated"), gecho.Send())
}
