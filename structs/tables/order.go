package tables

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	// Table Name and identifiers
	tableName   struct{}  `bun:"table:orders,alias:o"`
	Id          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id" validate:"omitempty,uuid4"`
	OrderNumber string    `bun:"order_number,notnull,unique" json:"order_number" validate:"omitempty,min=8,max=50"`

	// Customer Data
	Name  string `bun:"name,notnull" json:"name" validate:"required,min=2,max=100"`
	Email string `bun:"email,notnull" json:"email" validate:"required,email"`
	Phone string `bun:"phone,notnull" json:"phone" validate:"required,min=10,max=20"`
	Note  string `bun:"note" json:"note,omitempty" validate:"omitempty,max=500"` // Customer Note

	// Address Data (Reference to Address table)
	AddressId uuid.UUID `bun:"address_id,notnull,type:uuid" json:"address_id" validate:"required,uuid4"`

	// Payment Data
	PaymentLink   string        `bun:"payment_link" json:"payment_link,omitempty" validate:"omitempty,url"` // Can be null initially because payment link will be attached later
	PaymentStatus PaymentStatus `bun:"payment_status,notnull,default:'unpaid'" json:"payment_status" validate:"required,oneof=unpaid paid"`

	// Shipping
	ShippingCents uint64 `bun:"shipping_cents" json:"shipping_cents"`

	// Order Data
	Status    OrderStatus `bun:"status,notnull,default:'pending'" json:"status" validate:"required,oneof=pending paid processing shipped delivered cancelled refunded"`
	CreatedAt time.Time   `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time   `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt *time.Time  `bun:"deleted_at,nullzero" json:"deleted_at,omitempty"`
}

type OrderLine struct {
	tableName struct{}  `bun:"table:order_lines,alias:ol"`
	Id        uuid.UUID `bun:"id,pk,notnull" json:"id" validate:"omitempty,uuid4"`
	OrderId   uuid.UUID `bun:"order_id,notnull,type:uuid" json:"order_id" validate:"required,uuid4"`
	ProductId uuid.UUID `bun:"product_id,notnull,type:uuid" json:"product_id" validate:"required,uuid4"`
	Quantity  int       `bun:"quantity,notnull" json:"quantity" validate:"required,min=1"`

	// Snapshot of pricing at time of order
	UnitPrice    uint64 `bun:"unit_price,notnull" json:"unit_price" validate:"required,gte=0"`        // Price when ordered
	UnitDiscount uint64 `bun:"unit_discount,notnull" json:"unit_discount" validate:"omitempty,gte=0"` // Discount when ordered
	UnitTax      uint64 `bun:"unit_tax,notnull" json:"unit_tax" validate:"required,gte=0"`            // Tax when ordered
	UnitSubtotal uint64 `bun:"unit_subtotal,notnull" json:"unit_subtotal" validate:"required,gte=0"`  // Subtotal when ordered
	LineTotal    uint64 `bun:"line_total,notnull" json:"line_total" validate:"required,gte=0"`        // quantity * unit_subtotal

	// Keep reference to product for name/SKU changes
	ProductName string `bun:"product_name,notnull" json:"product_name" validate:"required,min=2,max=200"` // Name when ordered
	ProductSKU  string `bun:"product_sku,notnull" json:"product_sku" validate:"required,min=3,max=50"`    // SKU when ordered
}

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusDelivered  OrderStatus = "delivered"
	OrderStatusCancelled  OrderStatus = "cancelled"
	OrderStatusRefunded   OrderStatus = "refunded"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid PaymentStatus = "unpaid"
	PaymentStatusPaid   PaymentStatus = "paid"
)
