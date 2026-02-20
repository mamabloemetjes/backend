package tables

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	tableName   struct{}       `bun:"table:products,alias:p"`
	ID          uuid.UUID      `bun:"id,pk,type:uuid" json:"id" validate:"omitempty,uuid4"`
	Name        string         `bun:"name,notnull" json:"name" validate:"required,min=2,max=200"`
	SKU         string         `bun:"sku,notnull" json:"sku" validate:"omitempty,min=3,max=50"`
	Price       uint64         `bun:"price,notnull" json:"price" validate:"required,gte=0"`          // stored in cents
	Discount    uint64         `bun:"discount" json:"discount,omitempty" validate:"omitempty,gte=0"` // stored in cents
	Tax         uint64         `bun:"tax,notnull" json:"tax" validate:"gte=0"`                       // stored in cents
	Subtotal    uint64         `bun:"subtotal,notnull" json:"subtotal" validate:"omitempty,gte=0"`   // computed: Price - Discount + Tax
	Description string         `bun:"description,notnull" json:"description" validate:"required,min=10,max=2000"`
	ProductType string         `bun:"product_type" json:"product_type" validate:"omitempty,oneof='wedding' 'funeral' 'birth'"`
	IsActive    bool           `bun:"is_active,notnull" json:"is_active"`
	CreatedAt   time.Time      `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time      `bun:"updated_at,notnull,default:now()" json:"updated_at"`
	Images      []ProductImage `bun:"rel:has-many,join:id=product_id" json:"images,omitempty" validate:"omitempty,dive"` // slice is nil if no images
}

// ProductImage represents an image for a product
type ProductImage struct {
	tableName struct{}  `bun:"table:product_images,alias:pi"`
	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id" validate:"omitempty,uuid4"`
	ProductID uuid.UUID `bun:"product_id,type:uuid,notnull" json:"product_id" validate:"omitempty,uuid4"`
	URL       string    `bun:"url,notnull" json:"url" validate:"required,url"`
	AltText   string    `bun:"alt_text" json:"alt_text,omitempty" validate:"omitempty,max=200"` // optional, empty string if none
	IsPrimary bool      `bun:"is_primary,notnull" json:"is_primary"`
}
