package structs

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	tableName   struct{}       `pg:"products,alias:p"`
	ID          uuid.UUID      `pg:"id,pk,type:uuid" json:"id"`
	Name        string         `pg:"name,notnull" json:"name"`
	SKU         string         `pg:"sku,notnull" json:"sku"`                      // Stock Keeping Unit for better inventory tracking
	Price       uint64         `pg:"price,notnull,use_zero" json:"price"`         // stored in cents
	Discount    uint64         `pg:"discount,use_zero" json:"discount,omitempty"` // stored in cents
	Tax         uint64         `pg:"tax,notnull,use_zero" json:"tax"`             // stored in cents
	Subtotal    uint64         `pg:"subtotal,notnull,use_zero" json:"subtotal"`   // computed: Price - Discount + Tax
	Description string         `pg:"description,notnull" json:"description"`
	IsActive    bool           `pg:"is_active,notnull,use_zero" json:"is_active"`
	CreatedAt   time.Time      `pg:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt   time.Time      `pg:"updated_at,notnull,default:now()" json:"updated_at"`
	Size        Size           `pg:"size" json:"size,omitempty"`                              // optional, use omitempty if nil/empty
	Colors      []Color        `pg:"colors,array" json:"colors,omitempty"`                    // optional, can be empty
	ProductType ProductType    `pg:"product_type" json:"product_type,omitempty"`              // optional, use omitempty
	Stock       uint16         `pg:"stock,use_zero" json:"stock,omitempty"`                   // changed to uint16 for higher inventory
	Images      []ProductImage `pg:"rel:has-many,join_fk:product_id" json:"images,omitempty"` // slice is nil if no images
}

// ProductImage represents an image for a product
type ProductImage struct {
	tableName struct{}  `pg:"product_images,alias:pi"`
	ID        uuid.UUID `pg:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `pg:"product_id,type:uuid,notnull" json:"product_id"`
	URL       string    `pg:"url,notnull" json:"url"`
	AltText   string    `pg:"alt_text" json:"alt_text,omitempty"` // optional, empty string if none
	IsPrimary bool      `pg:"is_primary,notnull,use_zero" json:"is_primary"`
}

// Size of the product
type Size string

const (
	SizeSmall  Size = "small"
	SizeMedium Size = "medium"
	SizeLarge  Size = "large"
)

// ProductType enum
type ProductType string

const (
	Flower  ProductType = "flower"
	Bouquet ProductType = "bouquet"
)

// Color enum
type Color string

const (
	ColorRed    Color = "red"
	ColorBlue   Color = "blue"
	ColorGreen  Color = "green"
	ColorYellow Color = "yellow"
	ColorBlack  Color = "black"
	ColorWhite  Color = "white"
	ColorPurple Color = "purple"
	ColorOrange Color = "orange"
	ColorPink   Color = "pink"
)
