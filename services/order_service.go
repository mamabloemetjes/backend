package services

import (
	"context"
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"runtime/debug"
	"slices"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
)

type OrderService struct {
	logger         *gecho.Logger
	cfg            *structs.Config
	db             *database.DB
	productService *ProductService
	emailService   *EmailService
}

func NewOrderService(
	logger *gecho.Logger,
	cfg *structs.Config,
	db *database.DB,
	productService *ProductService,
	emailService *EmailService,
) *OrderService {
	return &OrderService{
		logger:         logger,
		cfg:            cfg,
		db:             db,
		productService: productService,
		emailService:   emailService,
	}
}

// CreateOrderFromRequest creates a complete order with address, order lines, and sends confirmation email
func (os *OrderService) CreateOrderFromRequest(ctx context.Context, req *structs.OrderRequest, userId *uuid.UUID) (order *tables.Order, err error) {
	os.logger.Info("CreateOrderFromRequest started", gecho.Field("products_count", len(req.Products)))

	// Start a Bun transaction
	os.logger.Info("Starting transaction")
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		os.logger.Error("Failed to begin transaction", gecho.Field("error", err))
		return nil, lib.MapPgError(err)
	}
	os.logger.Info("Transaction started successfully", gecho.Field("tx_type", fmt.Sprintf("%T", tx)))

	defer func() {
		if p := recover(); p != nil {
			stackTrace := string(debug.Stack())
			os.logger.Error(fmt.Sprintf("PANIC RECOVERED: %v", p),
				gecho.Field("panic_value", p),
				gecho.Field("stack_trace", stackTrace))
			tx.Rollback()
			err = fmt.Errorf("panic recovered: %v", p)
		} else if err != nil {
			os.logger.Info("Rolling back transaction due to error", gecho.Field("error", err))
			tx.Rollback()
		} else {
			os.logger.Info("Committing transaction")
			err = tx.Commit()
		}
	}()

	// Validate all products exist and are active
	os.logger.Info("Validating product IDs")
	productIds := make([]uuid.UUID, 0, len(req.Products))
	for idStr := range req.Products {
		id, parseErr := uuid.Parse(idStr)
		if parseErr != nil {
			err = fmt.Errorf("invalid product ID: %s", idStr)
			return nil, err
		}
		productIds = append(productIds, id)
	}
	os.logger.Info("Product IDs validated", gecho.Field("count", len(productIds)))

	// Fetch all products
	os.logger.Info("Fetching products by IDs", gecho.Field("ids", productIds))
	products, fetchErr := os.productService.GetProductsByIds(ctx, productIds)
	if fetchErr != nil {
		os.logger.Error("Failed to fetch products", gecho.Field("error", fetchErr))
		err = fetchErr
		return nil, err
	}
	os.logger.Info("Products fetched successfully", gecho.Field("count", len(products)))

	// Validate all products are active
	os.logger.Info("Building product map and validating active status")
	productMap := make(map[string]*tables.Product)
	for _, product := range products {
		os.logger.Info("Processing product",
			gecho.Field("id", product.ID),
			gecho.Field("name", product.Name),
			gecho.Field("active", product.IsActive))

		if !product.IsActive {
			err = fmt.Errorf("product %s (%s) is no longer available", product.Name, product.SKU)
			return nil, err
		}
		productMap[product.ID.String()] = product
	}
	os.logger.Info("Product map built", gecho.Field("map_size", len(productMap)))

	// Check if all requested products exist
	unavailableProducts := []string{}
	for idStr := range req.Products {
		if _, exists := productMap[idStr]; !exists {
			unavailableProducts = append(unavailableProducts, idStr)
		}
	}
	if len(unavailableProducts) > 0 {
		err = fmt.Errorf("products not found: %v", unavailableProducts)
		return nil, err
	}

	// Encrypt sensitive data BEFORE creating address/order
	os.logger.Info("Starting encryption of sensitive data")

	// Encrypt customer contact info
	encryptedEmail, err := lib.Encrypt(req.Email, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt email", gecho.Field("error", err))
		return nil, err
	}
	os.logger.Info("Email encrypted successfully")

	encryptedPhone, err := lib.Encrypt(req.Phone, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt phone", gecho.Field("error", err))
		return nil, err
	}
	os.logger.Info("Phone encrypted successfully")

	encryptedName, err := lib.Encrypt(req.Name, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt name", gecho.Field("error", err))
		return nil, err
	}
	os.logger.Info("Name encrypted successfully")

	// Encrypt customer note if provided
	var encryptedNote string
	if req.CustomerNote != "" {
		encryptedNote, err = lib.Encrypt(req.CustomerNote, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Error("Failed to encrypt note", gecho.Field("error", err))
			return nil, err
		}
		os.logger.Info("Note encrypted successfully")
	}

	// Encrypt address fields
	encryptedStreet, err := lib.Encrypt(req.Street, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt street", gecho.Field("error", err))
		return nil, err
	}

	encryptedHouseNo, err := lib.Encrypt(req.HouseNo, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt house number", gecho.Field("error", err))
		return nil, err
	}

	encryptedPostalCode, err := lib.Encrypt(req.PostalCode, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt postal code", gecho.Field("error", err))
		return nil, err
	}

	encryptedCity, err := lib.Encrypt(req.City, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to encrypt city", gecho.Field("error", err))
		return nil, err
	}
	os.logger.Info("All address fields encrypted successfully")

	// Create address with encrypted fields
	addressId := uuid.New()
	address := &tables.Address{
		Id:         addressId,
		UserId:     userId, // Pointer, nullable for guest orders
		Street:     encryptedStreet,
		HouseNo:    encryptedHouseNo,
		PostalCode: encryptedPostalCode,
		City:       encryptedCity,
		Country:    req.Country, // Country code can remain unencrypted for regional statistics
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	os.logger.Info("Inserting address", gecho.Field("address_id", addressId))
	_, err = tx.NewInsert().Model(address).Exec(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}
	os.logger.Info("Address inserted successfully")

	// Create order
	orderId := uuid.New()
	orderNumber := lib.GenerateOrderNumber()

	order = &tables.Order{
		Id:            orderId,
		OrderNumber:   orderNumber,
		Name:          encryptedName,
		Email:         encryptedEmail,
		Phone:         encryptedPhone,
		Note:          encryptedNote,
		AddressId:     addressId,
		PaymentLink:   "",
		PaymentStatus: tables.PaymentStatusUnpaid,
		Status:        tables.OrderStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	os.logger.Info("Inserting order",
		gecho.Field("order_id", orderId),
		gecho.Field("order_number", orderNumber))
	_, err = tx.NewInsert().Model(order).Exec(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Create order lines with pricing snapshots
	orderLines := make([]*tables.OrderLine, 0, len(req.Products))
	for idStr, quantity := range req.Products {
		product := productMap[idStr]

		lineTotal := product.Subtotal * uint64(quantity)

		orderLine := &tables.OrderLine{
			Id:           uuid.New(),
			OrderId:      orderId,
			ProductId:    product.ID,
			Quantity:     quantity,
			UnitPrice:    product.Price,
			UnitDiscount: product.Discount,
			UnitTax:      product.Tax,
			UnitSubtotal: product.Subtotal,
			LineTotal:    lineTotal,
			ProductName:  product.Name,
			ProductSKU:   product.SKU,
		}

		orderLines = append(orderLines, orderLine)
	}

	os.logger.Info("Inserting order lines", gecho.Field("count", len(orderLines)))
	_, err = tx.NewInsert().Model(&orderLines).Exec(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Deactivate products that were purchased
	os.logger.Info("Deactivating purchased products")
	for idStr := range req.Products {
		product := productMap[idStr]

		os.logger.Info("Deactivating product",
			gecho.Field("product_id", product.ID),
			gecho.Field("product_name", product.Name),
			gecho.Field("product_sku", product.SKU))

		// Set product as inactive
		_, err = tx.NewUpdate().
			Model((*tables.Product)(nil)).
			Set("is_active = ?", false).
			Set("updated_at = ?", time.Now()).
			Where("id = ?", product.ID).
			Exec(ctx)

		if err != nil {
			os.logger.Error("Failed to deactivate product",
				gecho.Field("error", err),
				gecho.Field("product_id", product.ID))
			return nil, lib.MapPgError(err)
		}
	}
	os.logger.Info("All purchased products deactivated successfully")

	// Send order confirmation email asynchronously (using original unencrypted data)
	go func() {
		emailErr := os.emailService.SendOrderConfirmationEmail(req.Email, req.Name, orderNumber, orderLines, &tables.Address{
			Street:     req.Street,
			HouseNo:    req.HouseNo,
			PostalCode: req.PostalCode,
			City:       req.City,
			Country:    req.Country,
		})
		if emailErr != nil {
			os.logger.Error("Failed to send order confirmation email",
				gecho.Field("error", emailErr),
				gecho.Field("order_id", orderId),
				gecho.Field("email", req.Email))
		} else {
			os.logger.Info("Order confirmation email sent",
				gecho.Field("order_id", orderId),
				gecho.Field("order_number", orderNumber))
		}
	}()

	os.logger.Info("Order created successfully",
		gecho.Field("order_id", orderId),
		gecho.Field("order_number", orderNumber))
	return order, nil
}

// GetOrderById retrieves an order by ID with decrypted PII
func (os *OrderService) GetOrderById(ctx context.Context, orderId uuid.UUID) (*tables.Order, error) {
	order, err := database.Query[tables.Order](os.db).
		Where("id", orderId).
		WhereRaw("deleted_at IS NULL").
		First(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Decrypt sensitive fields
	order.Name, err = lib.Decrypt(order.Name, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt order name", gecho.Field("error", err))
		// Continue with encrypted value rather than failing
	}

	order.Email, err = lib.Decrypt(order.Email, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to decrypt email", gecho.Field("error", err))
		return nil, err
	}

	order.Phone, err = lib.Decrypt(order.Phone, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt order phone", gecho.Field("error", err))
		// Continue with encrypted value rather than failing
	}

	// Decrypt customer note if present
	if order.Note != "" {
		order.Note, err = lib.Decrypt(order.Note, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt order note", gecho.Field("error", err))
			// Continue with encrypted value rather than failing
		}
	}

	return order, nil
}

// GetAddressById retrieves an address by ID with decrypted fields
func (os *OrderService) GetAddressById(ctx context.Context, addressId uuid.UUID) (*tables.Address, error) {
	address, err := database.Query[tables.Address](os.db).
		Where("id", addressId).
		First(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Decrypt address fields
	address.Street, err = lib.Decrypt(address.Street, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt street", gecho.Field("error", err))
	}

	address.HouseNo, err = lib.Decrypt(address.HouseNo, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt house number", gecho.Field("error", err))
	}

	address.PostalCode, err = lib.Decrypt(address.PostalCode, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt postal code", gecho.Field("error", err))
	}

	address.City, err = lib.Decrypt(address.City, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt city", gecho.Field("error", err))
	}

	return address, nil
}

// GetOrderByOrderNumber retrieves an order by order number with decrypted PII
func (os *OrderService) GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*tables.Order, error) {
	order, err := database.Query[tables.Order](os.db).
		Where("order_number", orderNumber).
		WhereRaw("deleted_at IS NULL").
		First(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Decrypt sensitive fields
	order.Name, err = lib.Decrypt(order.Name, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Warn("Failed to decrypt order name", gecho.Field("error", err))
		// Continue with encrypted value rather than failing
	}

	order.Email, err = lib.Decrypt(order.Email, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to decrypt email", gecho.Field("error", err))
		return nil, err
	}

	order.Phone, err = lib.Decrypt(order.Phone, os.cfg.Encryption.Key)
	if err != nil {
		os.logger.Error("Failed to decrypt phone", gecho.Field("error", err))
		return nil, err
	}

	// Decrypt customer note if present
	if order.Note != "" {
		order.Note, err = lib.Decrypt(order.Note, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt order note", gecho.Field("error", err))
			// Continue with encrypted value rather than failing
		}
	}

	return order, nil
}

// GetAllOrders retrieves all orders with optional filtering
func (os *OrderService) GetAllOrders(ctx context.Context, status *tables.OrderStatus, paymentStatus *tables.PaymentStatus, limit, offset int) ([]*tables.Order, int, error) {
	query := database.Query[tables.Order](os.db).
		WhereRaw("deleted_at IS NULL")

	if status != nil {
		query = query.Where("status", *status)
	}

	if paymentStatus != nil {
		query = query.Where("payment_status", *paymentStatus)
	}

	// Get total count
	count, err := query.Count(ctx)
	if err != nil {
		return nil, 0, lib.MapPgError(err)
	}

	// Get paginated results
	orders, err := query.
		OrderBy("created_at", database.DESC).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, lib.MapPgError(err)
	}

	// Convert to pointer slice and decrypt sensitive fields
	result := make([]*tables.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]

		result[i].Name, err = lib.Decrypt(result[i].Name, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt name", gecho.Field("error", err))
		}

		result[i].Email, err = lib.Decrypt(result[i].Email, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Error("Failed to decrypt email", gecho.Field("error", err))
			continue
		}

		result[i].Phone, err = lib.Decrypt(result[i].Phone, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Error("Failed to decrypt phone", gecho.Field("error", err))
			continue
		}

		// Decrypt note if present
		if result[i].Note != "" {
			result[i].Note, err = lib.Decrypt(result[i].Note, os.cfg.Encryption.Key)
			if err != nil {
				os.logger.Warn("Failed to decrypt note", gecho.Field("error", err))
			}
		}
	}

	return result, count, nil
}

// GetOrdersByUserId retrieves all orders for a specific user
func (os *OrderService) GetOrdersByUserId(ctx context.Context, userId uuid.UUID) ([]*tables.Order, error) {
	// First get all addresses for the user
	addresses, err := database.Query[tables.Address](os.db).
		WhereRaw("user_id = ?", userId).
		All(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	addressIds := make([]uuid.UUID, len(addresses))
	for i, addr := range addresses {
		addressIds[i] = addr.Id
	}

	os.logger.Info("Found addresses for user", gecho.Field("user_id", userId), gecho.Field("address_count", len(addressIds)))

	if len(addressIds) == 0 {
		return []*tables.Order{}, nil
	}

	// Convert to interface slice for WhereIn
	addressIdsIface := make([]any, len(addressIds))
	for i, id := range addressIds {
		addressIdsIface[i] = id
	}

	// Get orders with those addresses
	orders, err := database.Query[tables.Order](os.db).
		WhereIn("address_id", addressIdsIface).
		WhereRaw("deleted_at IS NULL").
		OrderBy("created_at", database.DESC).
		All(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Convert to pointer slice and decrypt sensitive fields
	result := make([]*tables.Order, len(orders))
	for i := range orders {
		result[i] = &orders[i]

		result[i].Name, err = lib.Decrypt(result[i].Name, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt name", gecho.Field("error", err))
		}

		result[i].Email, err = lib.Decrypt(result[i].Email, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Error("Failed to decrypt email", gecho.Field("error", err))
			continue
		}

		result[i].Phone, err = lib.Decrypt(result[i].Phone, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Error("Failed to decrypt phone", gecho.Field("error", err))
			continue
		}

		// Decrypt note if present
		if result[i].Note != "" {
			result[i].Note, err = lib.Decrypt(result[i].Note, os.cfg.Encryption.Key)
			if err != nil {
				os.logger.Warn("Failed to decrypt note", gecho.Field("error", err))
			}
		}
	}

	return result, nil
}

// GetOrderLinesByOrderId retrieves all order lines for an order
func (os *OrderService) GetOrderLinesByOrderId(ctx context.Context, orderId uuid.UUID) ([]*tables.OrderLine, error) {
	orderLines, err := database.Query[tables.OrderLine](os.db).
		Where("order_id", orderId).
		All(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Convert slice to pointer slice
	result := make([]*tables.OrderLine, len(orderLines))
	for i := range orderLines {
		result[i] = &orderLines[i]
	}

	return result, nil
}

// UpdateOrderStatus updates the order status with validation
func (os *OrderService) UpdateOrderStatus(ctx context.Context, orderId uuid.UUID, newStatus tables.OrderStatus) error {
	// Get current order
	order, err := os.GetOrderById(ctx, orderId)
	if err != nil {
		return err
	}

	// Validate status transitions
	if !os.isValidStatusTransition(order.Status, newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", order.Status, newStatus)
	}

	// Update status
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in UpdateOrderStatus: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.NewUpdate().
		Model(&tables.Order{}).
		Set("status = ?", newStatus).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", orderId).
		Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	os.logger.Info("Order status updated",
		gecho.Field("order_id", orderId),
		gecho.Field("old_status", order.Status),
		gecho.Field("new_status", newStatus))

	return nil
}

// AttachPaymentLink adds a payment link to an order and sends email to customer
func (os *OrderService) AttachPaymentLink(ctx context.Context, orderId uuid.UUID, paymentLink string) error {
	// Get order
	order, err := os.GetOrderById(ctx, orderId)
	if err != nil {
		return err
	}

	// Update payment link
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in AttachPaymentLink: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.NewUpdate().
		Model(&tables.Order{}).
		Set("payment_link = ?", paymentLink).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", orderId).
		Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	// Send payment link email asynchronously
	go func() {
		emailErr := os.emailService.SendPaymentLinkEmail(order.Email, order.Name, order.OrderNumber, paymentLink)
		if emailErr != nil {
			os.logger.Error("Failed to send payment link email",
				gecho.Field("error", emailErr),
				gecho.Field("order_id", orderId),
				gecho.Field("email", order.Email))
		} else {
			os.logger.Info("Payment link email sent",
				gecho.Field("order_id", orderId),
				gecho.Field("order_number", order.OrderNumber))
		}
	}()

	return nil
}

// MarkOrderAsPaid marks an order as paid
func (os *OrderService) MarkOrderAsPaid(ctx context.Context, orderId uuid.UUID) error {
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in MarkOrderAsPaid: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.NewUpdate().
		Model(&tables.Order{}).
		Set("payment_status = ?", tables.PaymentStatusPaid).
		Set("status = ?", tables.OrderStatusPaid).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", orderId).
		Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	os.logger.Info("Order marked as paid", gecho.Field("order_id", orderId))

	return nil
}

// SoftDeleteOrder soft deletes an order
func (os *OrderService) SoftDeleteOrder(ctx context.Context, orderId uuid.UUID) error {
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in SoftDeleteOrder: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.NewUpdate().
		Model(&tables.Order{}).
		Set("deleted_at = ?", time.Now()).
		Where("id = ?", orderId).
		Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	os.logger.Info("Order soft deleted", gecho.Field("order_id", orderId))

	return nil
}

// isValidStatusTransition validates if a status transition is allowed
func (os *OrderService) isValidStatusTransition(current, next tables.OrderStatus) bool {
	// Define allowed transitions
	transitions := map[tables.OrderStatus][]tables.OrderStatus{
		tables.OrderStatusPending: {
			tables.OrderStatusPaid,
			tables.OrderStatusCancelled,
		},
		tables.OrderStatusPaid: {
			tables.OrderStatusProcessing,
			tables.OrderStatusCancelled,
			tables.OrderStatusRefunded,
		},
		tables.OrderStatusProcessing: {
			tables.OrderStatusShipped,
			tables.OrderStatusCancelled,
			tables.OrderStatusRefunded,
		},
		tables.OrderStatusShipped: {
			tables.OrderStatusDelivered,
		},
		tables.OrderStatusDelivered: {
			tables.OrderStatusRefunded,
		},
		tables.OrderStatusCancelled: {},
		tables.OrderStatusRefunded:  {},
	}

	allowedNextStates, exists := transitions[current]
	if !exists {
		return false
	}

	if slices.Contains(allowedNextStates, next) {
		return true
	}

	return false
}

// CreateAddress creates a new address
func (os *OrderService) CreateAddress(ctx context.Context, address *tables.Address) error {
	tx, err := os.db.BeginTx(ctx, nil)
	if err != nil {
		return lib.MapPgError(err)
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			os.logger.Error(fmt.Sprintf("panic in CreateAddress: %v", p))
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.NewInsert().Model(address).Exec(ctx)
	if err != nil {
		return lib.MapPgError(err)
	}

	return nil
}

// GetUserAddresses retrieves all addresses for a user with decrypted fields
func (os *OrderService) GetUserAddresses(ctx context.Context, userId uuid.UUID) ([]*tables.Address, error) {
	addresses, err := database.Query[tables.Address](os.db).
		Where("user_id", userId).
		OrderBy("created_at", database.DESC).
		All(ctx)
	if err != nil {
		return nil, lib.MapPgError(err)
	}

	// Convert to pointer slice and decrypt fields
	result := make([]*tables.Address, len(addresses))
	for i := range addresses {
		addr := &addresses[i]

		// Decrypt address fields
		addr.Street, err = lib.Decrypt(addr.Street, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt street", gecho.Field("error", err))
		}

		addr.HouseNo, err = lib.Decrypt(addr.HouseNo, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt house number", gecho.Field("error", err))
		}

		addr.PostalCode, err = lib.Decrypt(addr.PostalCode, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt postal code", gecho.Field("error", err))
		}

		addr.City, err = lib.Decrypt(addr.City, os.cfg.Encryption.Key)
		if err != nil {
			os.logger.Warn("Failed to decrypt city", gecho.Field("error", err))
		}

		result[i] = addr
	}

	return result, nil
}
