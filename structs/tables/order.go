package tables

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	// Table Name and identifiers
	tableName   struct{}  `bun:"table:orders,alias:o"`
	Id          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	OrderNumber string    `bun:"order_number,notnull,unique" json:"order_number"`

	// Customer Data
	Name  string `bun:"name,notnull" json:"name"`
	Email string `bun:"email,notnull" json:"email"`
	Phone string `bun:"phone,notnull" json:"phone"`
	Note  string `bun:"note" json:"note,omitempty"` // Customer Note

	// Address Data (Reference to Address table)
	AddressId uuid.UUID `bun:"address_id,notnull,type:uuid" json:"address_id"`

	// Payment Data
	PaymentLink   string        `bun:"payment_link" json:"payment_link,omitempty"` // Can be null initially because payment link will be attached later
	PaymentStatus PaymentStatus `bun:"payment_status,notnull,default:'unpaid'" json:"payment_status"`

	// Order Data
	Status    OrderStatus `bun:"status,notnull,default:'pending'" json:"status"`
	CreatedAt time.Time   `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time   `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt *time.Time  `bun:"deleted_at,nullzero" json:"deleted_at,omitempty"`
}

type OrderLine struct {
	tableName struct{}  `bun:"table:order_lines,alias:ol"`
	Id        uuid.UUID `bun:"id,pk,notnull" json:"id"`
	OrderId   uuid.UUID `bun:"order_id,notnull,type:uuid" json:"order_id"`
	ProductId uuid.UUID `bun:"product_id,notnull,type:uuid" json:"product_id"`
	Quantity  int       `bun:"quantity,notnull" json:"quantity"`

	// Snapshot of pricing at time of order
	UnitPrice    uint64 `bun:"unit_price,notnull" json:"unit_price"`       // Price when ordered
	UnitDiscount uint64 `bun:"unit_discount,notnull" json:"unit_discount"` // Discount when ordered
	UnitTax      uint64 `bun:"unit_tax,notnull" json:"unit_tax"`           // Tax when ordered
	UnitSubtotal uint64 `bun:"unit_subtotal,notnull" json:"unit_subtotal"` // Subtotal when ordered
	LineTotal    uint64 `bun:"line_total,notnull" json:"line_total"`       // quantity * unit_subtotal

	// Keep reference to product for name/SKU changes
	ProductName string `bun:"product_name,notnull" json:"product_name"` // Name when ordered
	ProductSKU  string `bun:"product_sku,notnull" json:"product_sku"`   // SKU when ordered
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
