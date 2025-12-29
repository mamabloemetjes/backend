package services

import (
	"context"
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"math/rand"
	"time"

	"github.com/MonkyMars/gecho"
)

type OrderService struct {
	logger         *gecho.Logger
	cfg            *structs.Config
	db             *database.DB
	productService *ProductService
}

func NewOrderService(logger *gecho.Logger, cfg *structs.Config, db *database.DB, productService *ProductService) *OrderService {
	return &OrderService{
		logger:         logger,
		cfg:            cfg,
		db:             db,
		productService: productService,
	}
}

// CreateOrder creates a new order with the given products.
func (os *OrderService) CreateOrder(ctx context.Context, order *tables.Order) error {
	// Start a transaction
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in CreateOrder: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert the order
	_, err = tx.NewInsert().Model(order).Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	return nil
}

// CreateOrderLines creates multiple order lines in a single transaction.
func (os *OrderService) CreateOrderLines(ctx context.Context, orderLines []*tables.OrderLine) error {
	// Start a transaction
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in CreateOrderLines: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	// Insert the order lines
	_, err = tx.NewInsert().Model(&orderLines).Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	return nil
}

// GenerateOrderNumber generates a unique order number.
func (os *OrderService) GenerateOrderNumber() string {
	// Use a local rand.Source + rand.Rand
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Last 4 digits of Unix milliseconds
	timePart := fmt.Sprintf("%04d", time.Now().UnixNano()/1e6%10000)

	// 4 random alphanumeric characters
	randomPart := make([]byte, 4)
	for i := range randomPart {
		randomPart[i] = chars[r.Intn(len(chars))]
	}

	return timePart + string(randomPart)
}
