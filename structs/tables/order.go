package tables

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	tableName   struct{}  `bun:"table:orders,alias:o"`
	Id          uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	OrderNumber string    `bun:"order_number,notnull,unique"`
	Name        string    `bun:"name,notnull"`
	Email       string    `bun:"email,notnull"`
	PaymentLink string    `bun:"payment_link,notnull"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}

type OrderLine struct {
	tableName struct{}  `bun:"table:order_lines,alias:ol"`
	id        uuid.UUID `bun:"id,pk,notnull"`
	OrderId   uuid.UUID `bun:"order_id,notnull,type:uuid"`
	ProductId uuid.UUID `bun:"product_id,notnull,type:uuid"`
	Quantity  int       `bun:"quantity,notnull"`
}
