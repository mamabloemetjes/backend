package tables

import (
	"github.com/google/uuid"
)

type Inventory struct {
	tableName   struct{}       `bun:"table:inventory,alias:inv"`
	ID          uuid.UUID      `bun:"id,pk,type:uuid" json:"id" validate:"omitempty,uuid4"`
	ProductID   uuid.UUID     `bun:"product_id,type:uuid" json:"product_id" validate:"omitempty,uuid4"`
	StockCount  int 		   `bun:"stock_count,type:int" json:"stock_count" validate:"omitempty,int"`
}
