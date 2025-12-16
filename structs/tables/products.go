package tables

import (
	"mamabloemetjes_server/structs"
	"time"

	"github.com/google/uuid"
)

type Product struct {
	tableName   struct{}            `bun:"table:products,alias:p"`
	ID          uuid.UUID           `bun:"id,pk,type:uuid" json:"id"`
	Name        string              `bun:"name,notnull" json:"name"`
	SKU         string              `bun:"sku,notnull" json:"sku"`             // Stock Keeping Unit for better inventory tracking
	Price       uint64              `bun:"price,notnull" json:"price"`         // stored in cents
	Discount    uint64              `bun:"discount" json:"discount,omitempty"` // stored in cents
	Tax         uint64              `bun:"tax,notnull" json:"tax"`             // stored in cents
	Subtotal    uint64              `bun:"subtotal,notnull" json:"subtotal"`   // computed: Price - Discount + Tax
	Description string              `bun:"description,notnull" json:"description"`
	IsActive    bool                `bun:"is_active,notnull" json:"is_active"`
	CreatedAt   time.Time           `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time           `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	Size        structs.Size        `bun:"size" json:"size,omitempty"`                              // optional, use omitempty if nil/empty
	Colors      []structs.Color     `bun:"colors,array" json:"colors,omitempty"`                    // optional, can be empty
	ProductType structs.ProductType `bun:"product_type" json:"product_type,omitempty"`              // optional, use omitempty
	Stock       uint16              `bun:"stock" json:"stock,omitempty"`                            // changed to uint16 for higher inventory
	Images      []ProductImage      `bun:"rel:has-many,join:id=product_id" json:"images,omitempty"` // slice is nil if no images
}

// ProductImage represents an image for a product
type ProductImage struct {
	tableName struct{}  `bun:"table:product_images,alias:pi"`
	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `bun:"product_id,type:uuid,notnull" json:"product_id"`
	URL       string    `bun:"url,notnull" json:"url"`
	AltText   string    `bun:"alt_text" json:"alt_text,omitempty"` // optional, empty string if none
	IsPrimary bool      `bun:"is_primary,notnull" json:"is_primary"`
}
