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
	Name        *string               `json:"name,omitempty" validate:"omitempty,min=2,max=200"`
	SKU         *string               `json:"sku,omitempty" validate:"omitempty,min=3,max=50"`
	Price       *uint64               `json:"price,omitempty" validate:"omitempty,gte=0"`
	Discount    *uint64               `json:"discount,omitempty" validate:"omitempty,gte=0"`
	Tax         *uint64               `json:"tax,omitempty" validate:"omitempty,gte=0"`
	Subtotal    *uint64               `json:"subtotal,omitempty" validate:"omitempty,gte=0"`
	Description *string               `json:"description,omitempty" validate:"omitempty,min=10,max=2000"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	Size        *string               `json:"size,omitempty" validate:"omitempty,min=1,max=50"`
	Colors      []string              `json:"colors,omitempty" validate:"omitempty,dive,min=2,max=50"`
	ProductType *string               `json:"product_type,omitempty" validate:"omitempty,min=2,max=100"`
	Stock       *uint16               `json:"stock,omitempty" validate:"omitempty,gte=0"`
	Images      []tables.ProductImage `json:"images,omitempty" validate:"omitempty,dive"`
}

type UpdateProductsRequest struct {
	Products map[string]UpdateProductRequest `json:"products" validate:"required,min=1,dive,keys,uuid4,endkeys"`
}

type UpdateProductsStockRequest struct {
	Stocks map[string]int `json:"stocks" validate:"required,min=1,dive,keys,uuid4,endkeys,required,gte=0"`
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
