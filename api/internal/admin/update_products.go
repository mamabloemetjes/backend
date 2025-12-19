package admin

import (
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs/tables"
	"net/http"
	"strings"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
	"github.com/uptrace/bun/dialect/pgdialect"
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
	// product ID - update data
	Products map[string]UpdateProductRequest `json:"products"`
}

/*
 * {
   "stocks": {
     id: int,
     id: int
   }
 }
*/

type UpdateProductsStockRequest struct {
	Stocks map[string]int `json:"stocks"`
	// product ID - new stock
}

func (ar *AdminRoutesManager) UpdateProducts(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[UpdateProductsRequest](r)
	if err != nil || len(body.Products) == 0 {
		ar.logger.Debug("Failed to extract and validate body", err)
		gecho.BadRequest(w, gecho.WithMessage("Please check the product information and try again"), gecho.Send())
		return
	}

	totalErrors := make(map[string]error)
	for productId, updateReq := range body.Products {
		// Validate product ID
		productUUID, parseErr := uuid.Parse(productId)
		if parseErr != nil || productUUID == uuid.Nil {
			totalErrors[productId] = fmt.Errorf("invalid product ID: %s", productId)
			ar.logger.Error("Invalid product ID", gecho.Field("error", parseErr), gecho.Field("product_id", productId))
			continue
		}

		// Build update map with only provided fields
		updateData := make(map[string]any)

		if updateReq.Name != nil {
			updateData["name"] = *updateReq.Name
		}
		if updateReq.SKU != nil {
			updateData["sku"] = *updateReq.SKU
		}
		if updateReq.Price != nil {
			updateData["price"] = *updateReq.Price
		}
		if updateReq.Discount != nil {
			updateData["discount"] = *updateReq.Discount
		}
		if updateReq.Tax != nil {
			updateData["tax"] = *updateReq.Tax
		}
		if updateReq.Description != nil {
			updateData["description"] = *updateReq.Description
		}
		if updateReq.IsActive != nil {
			updateData["is_active"] = *updateReq.IsActive
		}
		if updateReq.Size != nil {
			updateData["size"] = *updateReq.Size
		}
		if updateReq.ProductType != nil {
			updateData["product_type"] = *updateReq.ProductType
		}
		if updateReq.Stock != nil {
			updateData["stock"] = *updateReq.Stock
		}
		if updateReq.Colors != nil {
			// Normalize colors to lowercase
			normalizedColors := make([]string, len(updateReq.Colors))
			for i, color := range updateReq.Colors {
				normalizedColors[i] = strings.ToLower(color)
			}
			updateData["colors"] = pgdialect.Array(normalizedColors)
		}

		// Handle images update if provided
		if updateReq.Images != nil {
			ar.logger.Debug("Updating product images",
				gecho.Field("product_id", productId),
				gecho.Field("new_image_count", len(updateReq.Images)),
			)

			// Delete existing images
			_, deleteErr := database.Query[tables.ProductImage](ar.db).
				Where("product_id", productUUID).
				Delete(r.Context())
			if deleteErr != nil {
				totalErrors[productId] = deleteErr
				ar.logger.Error("Failed to delete existing images", gecho.Field("error", deleteErr), gecho.Field("product_id", productId))
				continue
			}

			// Insert new images if any provided
			if len(updateReq.Images) > 0 {
				// Set product_id and ID for all images and ensure only one primary image
				hasPrimary := false
				for i := range updateReq.Images {
					// Generate UUID if not set
					if updateReq.Images[i].ID == uuid.Nil {
						updateReq.Images[i].ID = uuid.New()
					}
					updateReq.Images[i].ProductID = productUUID

					ar.logger.Debug("Processing image for update",
						gecho.Field("index", i),
						gecho.Field("image_id", updateReq.Images[i].ID),
						gecho.Field("url", updateReq.Images[i].URL),
						gecho.Field("is_primary", updateReq.Images[i].IsPrimary),
					)

					// Ensure only the first marked primary image is actually primary
					if updateReq.Images[i].IsPrimary {
						if hasPrimary {
							updateReq.Images[i].IsPrimary = false
						} else {
							hasPrimary = true
						}
					}
				}

				// If no primary image was set, make the first one primary
				if !hasPrimary {
					updateReq.Images[0].IsPrimary = true
					ar.logger.Debug("Set first image as primary", gecho.Field("url", updateReq.Images[0].URL))
				}

				// Insert new images
				result, insertErr := ar.db.NewInsert().Model(&updateReq.Images).Exec(r.Context())
				if insertErr != nil {
					totalErrors[productId] = insertErr
					ar.logger.Error("Failed to insert new images", gecho.Field("error", insertErr), gecho.Field("product_id", productId))
					continue
				}

				rowsAffected, _ := result.RowsAffected()
				ar.logger.Debug("Updated product images",
					gecho.Field("product_id", productId),
					gecho.Field("image_count", len(updateReq.Images)),
					gecho.Field("rows_affected", rowsAffected),
				)
			}
		}

		// Calculate subtotal if price, discount, or tax changed
		if updateReq.Price != nil || updateReq.Discount != nil || updateReq.Tax != nil {
			// Fetch current product to get missing values
			currentProduct, err := database.Query[tables.Product](ar.db).Where("id", productId).First(r.Context())
			if err != nil {
				totalErrors[productId] = err
				ar.logger.Error("Failed to fetch current product", gecho.Field("error", err), gecho.Field("product_id", productId))
				continue
			}

			price := currentProduct.Price
			discount := currentProduct.Discount
			tax := currentProduct.Tax

			if updateReq.Price != nil {
				price = *updateReq.Price
			}
			if updateReq.Discount != nil {
				discount = *updateReq.Discount
			}
			if updateReq.Tax != nil {
				tax = *updateReq.Tax
			}

			updateData["subtotal"] = price - discount + tax
		}

		_, err := database.Query[tables.Product](ar.db).Where("id", productId).Update(r.Context(), updateData)
		if err != nil {
			totalErrors[productId] = err
			ar.logger.Error("Failed to update product", gecho.Field("error", err), gecho.Field("product_id", productId))
		} else {
			// Invalidate product caches asynchronously
			productUUID, parseErr := uuid.Parse(productId)
			if parseErr == nil {
				go func(id uuid.UUID) {
					if cacheErr := ar.cacheService.InvalidateProductCaches(id); cacheErr != nil {
						ar.logger.Warn("Failed to invalidate product caches after update",
							gecho.Field("error", cacheErr),
							gecho.Field("product_id", id),
						)
					}
				}(productUUID)
			}
		}
	}

	if len(totalErrors) > 0 {
		gecho.InternalServerError(w,
			gecho.WithMessage("Some products failed to update"),
			gecho.WithData(map[string]any{
				"erros": totalErrors,
			}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("Products updated successfully"),
		gecho.Send(),
	)
}

func (ar *AdminRoutesManager) UpdateProductsStock(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[UpdateProductsStockRequest](r)
	if err != nil || len(body.Stocks) == 0 {
		ar.logger.Debug("Failed to extract and validate body", err)
		gecho.BadRequest(w, gecho.WithMessage("Please check the stock information and try again"), gecho.Send())
		return
	}

	totalErrors := make(map[string]error)
	for productId, newStock := range body.Stocks {
		rowsAffected, err := database.Query[tables.Product](ar.db).Where("id", productId).Update(r.Context(), map[string]any{
			"stock": newStock,
		})
		if err != nil || rowsAffected == 0 {
			totalErrors[productId] = err
			ar.logger.Error("Failed to update product stock", gecho.Field("error", err), gecho.Field("product_id", productId))
		} else {
			// Invalidate product caches asynchronously
			productUUID, parseErr := uuid.Parse(productId)
			if parseErr == nil {
				go func(id uuid.UUID) {
					if cacheErr := ar.cacheService.InvalidateProductCaches(id); cacheErr != nil {
						ar.logger.Warn("Failed to invalidate product caches after stock update",
							gecho.Field("error", cacheErr),
							gecho.Field("product_id", id),
						)
					}
				}(productUUID)
			}
		}
	}

	if len(totalErrors) > 0 {
		gecho.InternalServerError(w,
			gecho.WithMessage("Some product stocks failed to update"),
			gecho.WithData(map[string]any{
				"erros": totalErrors,
			}),
			gecho.Send(),
		)
		return
	}

	gecho.Success(w,
		gecho.WithMessage("Product stocks updated successfully"),
		gecho.Send(),
	)
}
