package services

import (
	"context"
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/structs/tables"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ProductService struct {
	logger       *gecho.Logger
	db           *database.DB
	cacheService *CacheService
}

func NewProductService(logger *gecho.Logger, db *database.DB, cacheService *CacheService) *ProductService {
	return &ProductService{
		logger:       logger,
		db:           db,
		cacheService: cacheService,
	}
}

// ProductListOptions contains filtering and pagination options for product queries
type ProductListOptions struct {
	// Pagination
	Page     int `json:"page"`
	PageSize int `json:"page_size"`

	// Filters
	IsActive      *bool      `json:"is_active,omitempty"`      // Filter by active status
	MinPrice      *uint64    `json:"min_price,omitempty"`      // Minimum price in cents
	MaxPrice      *uint64    `json:"max_price,omitempty"`      // Maximum price in cents
	SearchTerm    string     `json:"search_term,omitempty"`    // Search in name, description, SKU
	SKUs          []string   `json:"skus,omitempty"`           // Filter by specific SKUs
	ExcludeSKUs   []string   `json:"exclude_skus,omitempty"`   // Exclude specific SKUs
	CreatedAfter  *time.Time `json:"created_after,omitempty"`  // Products created after this date
	CreatedBefore *time.Time `json:"created_before,omitempty"` // Products created before this date

	// Sorting
	SortBy        string `json:"sort_by"`        // Field to sort by (created_at, price, name)
	SortDirection string `json:"sort_direction"` // ASC or DESC

	// Relations
	IncludeImages bool `json:"include_images"` // Preload product images

	// Performance
	Timeout time.Duration `json:"-"` // Query timeout (not exposed in JSON)
}

// ProductListResult wraps the product list response with metadata
type ProductListResult struct {
	Products   []tables.Product    `json:"products"`
	Pagination database.Pagination `json:"pagination"`
	Filters    ProductListOptions  `json:"filters"`
	QueryTime  time.Duration       `json:"query_time"`
}

// GetAllProducts retrieves products with comprehensive filtering, pagination, and error handling
// This is the main production-ready method for listing products
func (ps *ProductService) GetAllProducts(ctx context.Context, opts *ProductListOptions) (*ProductListResult, error) {
	startTime := time.Now()

	// Validate and apply defaults to options
	if opts == nil {
		opts = &ProductListOptions{}
	}
	ps.applyDefaultOptions(opts)

	// Validate options
	if err := ps.validateOptions(opts); err != nil {
		ps.logger.Error("Invalid product list options", gecho.Field("error", err), gecho.Field("options", opts))
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Add query timeout if not set
	queryCtx := ctx
	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		queryCtx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	// Build the query
	query := database.Query[tables.Product](ps.db)

	// Apply filters
	query = ps.applyFilters(query, opts)

	// Apply sorting
	query = ps.applySorting(query, opts)

	// Preload images if requested
	if opts.IncludeImages {
		query = query.Relation("Images")
	}

	// Execute paginated query
	result, err := database.Paginate(query, queryCtx, opts.Page, opts.PageSize)
	if err != nil {
		ps.logger.Error("Failed to fetch products",
			gecho.Field("error", err),
			gecho.Field("page", opts.Page),
			gecho.Field("pageSize", opts.PageSize),
			gecho.Field("duration", time.Since(startTime)))
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Log successful query
	ps.logger.Debug("Products fetched successfully",
		gecho.Field("count", len(result.Data)),
		gecho.Field("total", result.Pagination.Total),
		gecho.Field("page", result.Pagination.Page),
		gecho.Field("pageSize", result.Pagination.PageSize),
		gecho.Field("duration", time.Since(startTime)),
	)

	// Build response
	return &ProductListResult{
		Products:   result.Data,
		Pagination: result.Pagination,
		Filters:    *opts,
		QueryTime:  time.Since(startTime),
	}, nil
}

// GetProductByID retrieves a single product by ID with optional image preloading
func (ps *ProductService) GetProductByID(ctx context.Context, id string, includeImages bool) (*tables.Product, error) {
	startTime := time.Now()

	// Try to get from cache first
	cachedProduct, err := ps.cacheService.GetProductByID(id, includeImages)
	if err != nil {
		ps.logger.Warn("Failed to get product from cache", gecho.Field("error", err), gecho.Field("id", id))
	} else if cachedProduct != nil {
		ps.logger.Debug("Product retrieved from cache", gecho.Field("id", id), gecho.Field("duration", time.Since(startTime)))
		return cachedProduct, nil
	}

	// Cache miss - fetch from database
	query := database.Query[tables.Product](ps.db).
		Where("id", id).
		Timeout(5 * time.Second)

	if includeImages {
		query = query.Relation("Images")
	}

	product, err := query.First(ctx)
	if err != nil {
		ps.logger.Error("Failed to fetch product by ID",
			gecho.Field("id", id),
			gecho.Field("error", err),
			gecho.Field("duration", time.Since(startTime)),
		)
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	if product == nil {
		ps.logger.Warn("Product not found", gecho.Field("id", id))
		return nil, fmt.Errorf("product not found")
	}

	// Cache the product asynchronously
	go func() {
		if err := ps.cacheService.SetProductByID(product, includeImages); err != nil {
			ps.logger.Warn("Failed to cache product", gecho.Field("error", err), gecho.Field("id", id))
		}
	}()

	ps.logger.Debug("Product fetched by ID",
		gecho.Field("id", id),
		gecho.Field("duration", time.Since(startTime)),
	)
	return product, nil
}

// GetActiveProducts is a convenience method to get only active products with caching
func (ps *ProductService) GetActiveProducts(ctx context.Context, page, pageSize int, includeImages bool) (*ProductListResult, error) {
	startTime := time.Now()

	// Try to get from cache first
	cachedProducts, err := ps.cacheService.GetActiveProductsList(page, pageSize, includeImages)
	if err != nil {
		ps.logger.Warn("Failed to get active products from cache", gecho.Field("error", err))
	} else if cachedProducts != nil {
		ps.logger.Debug("Active products retrieved from cache",
			gecho.Field("count", len(cachedProducts)),
			gecho.Field("page", page),
			gecho.Field("duration", time.Since(startTime)),
		)

		// Build result from cache (pagination info needs to be fetched or cached separately)
		// For now, return a simple result - you may want to cache pagination metadata too
		return &ProductListResult{
			Products: cachedProducts,
			Pagination: database.Pagination{
				Page:     page,
				PageSize: pageSize,
				Total:    len(cachedProducts), // This is approximate - real total would need separate query/cache
			},
			Filters: ProductListOptions{
				Page:          page,
				PageSize:      pageSize,
				IncludeImages: includeImages,
			},
			QueryTime: time.Since(startTime),
		}, nil
	}

	// Cache miss - fetch from database
	isActive := true
	opts := &ProductListOptions{
		Page:          page,
		PageSize:      pageSize,
		IsActive:      &isActive,
		IncludeImages: includeImages,
		SortBy:        "created_at",
		SortDirection: "DESC",
	}

	result, err := ps.GetAllProducts(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Cache the products asynchronously
	go func() {
		if err := ps.cacheService.SetActiveProductsList(page, pageSize, includeImages, result.Products); err != nil {
			ps.logger.Warn("Failed to cache active products", gecho.Field("error", err))
		}
	}()

	return result, nil
}

// GetProductsBySKUs retrieves multiple products by their SKUs
func (ps *ProductService) GetProductsBySKUs(ctx context.Context, skus []string, includeImages bool) ([]tables.Product, error) {
	startTime := time.Now()

	if len(skus) == 0 {
		return []tables.Product{}, nil
	}

	// Convert SKUs to interface slice
	skuInterfaces := make([]any, len(skus))
	for i, sku := range skus {
		skuInterfaces[i] = sku
	}

	query := database.Query[tables.Product](ps.db).
		WhereIn("sku", skuInterfaces).
		Timeout(10 * time.Second)

	if includeImages {
		query = query.Relation("Images")
	}

	products, err := query.All(ctx)
	if err != nil {
		ps.logger.Error("Failed to fetch products by SKUs",
			gecho.Field("skus", skus),
			gecho.Field("error", err),
			gecho.Field("duration", time.Since(startTime)),
		)
		return nil, fmt.Errorf("failed to fetch products by SKUs: %w", err)
	}

	return products, nil
}

// GetProductCount returns the total count of products matching the filters
func (ps *ProductService) GetProductCount(ctx context.Context, opts *ProductListOptions) (int, error) {
	if opts == nil {
		opts = &ProductListOptions{}
	}

	query := database.Query[tables.Product](ps.db)
	query = ps.applyFilters(query, opts)

	count, err := query.Count(ctx)
	if err != nil {
		ps.logger.Error("Failed to count products", gecho.Field("error", err), gecho.Field("options", opts))
		return 0, fmt.Errorf("failed to count products: %w", err)
	}

	return count, nil
}

// applyDefaultOptions sets default values for unspecified options
func (ps *ProductService) applyDefaultOptions(opts *ProductListOptions) {
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 20
	}
	if opts.PageSize > 100 {
		opts.PageSize = 100 // Max page size for performance
	}
	if opts.SortBy == "" {
		opts.SortBy = "created_at"
	}
	if opts.SortDirection == "" {
		opts.SortDirection = "DESC"
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second // Default 30s timeout
	}
}

// validateOptions validates the provided options
func (ps *ProductService) validateOptions(opts *ProductListOptions) error {
	// Validate sort field
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"price":      true,
		"name":       true,
		"sku":        true,
	}
	if !validSortFields[opts.SortBy] {
		return fmt.Errorf("invalid sort field: %s", opts.SortBy)
	}

	// Validate sort direction
	if opts.SortDirection != "ASC" && opts.SortDirection != "DESC" {
		return fmt.Errorf("invalid sort direction: %s (must be ASC or DESC)", opts.SortDirection)
	}

	// Validate price range
	if opts.MinPrice != nil && opts.MaxPrice != nil && *opts.MinPrice > *opts.MaxPrice {
		return fmt.Errorf("min_price cannot be greater than max_price")
	}

	return nil
}

// applyFilters applies all filter conditions to the query
func (ps *ProductService) applyFilters(query *database.QueryBuilder[tables.Product], opts *ProductListOptions) *database.QueryBuilder[tables.Product] {
	// Filter by active status (default to active only if not specified)
	if opts.IsActive != nil {
		query = query.Where("is_active", *opts.IsActive)
	}

	// Filter by price range
	if opts.MinPrice != nil {
		query = query.WhereOp("price", ">=", *opts.MinPrice)
	}
	if opts.MaxPrice != nil {
		query = query.WhereOp("price", "<=", *opts.MaxPrice)
	}

	// Search in name, description, or SKU
	if opts.SearchTerm != "" {
		searchPattern := "%" + opts.SearchTerm + "%"
		query = query.WhereRaw(
			"(name ILIKE ? OR description ILIKE ? OR sku ILIKE ?)",
			searchPattern, searchPattern, searchPattern,
		)
	}

	// Filter by specific SKUs
	if len(opts.SKUs) > 0 {
		skuInterfaces := make([]any, len(opts.SKUs))
		for i, sku := range opts.SKUs {
			skuInterfaces[i] = sku
		}
		query = query.WhereIn("sku", skuInterfaces)
	}

	// Exclude specific SKUs
	if len(opts.ExcludeSKUs) > 0 {
		skuInterfaces := make([]any, len(opts.ExcludeSKUs))
		for i, sku := range opts.ExcludeSKUs {
			skuInterfaces[i] = sku
		}
		query = query.WhereNotIn("sku", skuInterfaces)
	}

	// Filter by creation date range
	if opts.CreatedAfter != nil {
		query = query.WhereOp("created_at", ">=", *opts.CreatedAfter)
	}
	if opts.CreatedBefore != nil {
		query = query.WhereOp("created_at", "<=", *opts.CreatedBefore)
	}

	return query
}

// applySorting applies sorting to the query
func (ps *ProductService) applySorting(query *database.QueryBuilder[tables.Product], opts *ProductListOptions) *database.QueryBuilder[tables.Product] {
	var direction database.OrderDirection
	if opts.SortDirection == "ASC" {
		direction = database.ASC
	} else {
		direction = database.DESC
	}

	query = query.OrderBy(opts.SortBy, direction)

	// Add secondary sort by ID for consistent ordering
	query = query.OrderBy("id", database.ASC)

	return query
}

// Create new product

func (ps *ProductService) CreateProduct(ctx context.Context, product *tables.Product) (*tables.Product, error) {
	startTime := time.Now()

	// Generate UUID for product if not set (needed for image references)
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	// Calculate subtotal
	product.Subtotal = product.Price - product.Discount + product.Tax

	// Store images separately to insert them after product creation
	images := product.Images
	product.Images = nil // Remove images from product to avoid relation insert issues

	// Insert product into database
	product, err := database.Query[tables.Product](ps.db).Insert(ctx, product)
	if err != nil {
		ps.logger.Error("Failed to create product",
			gecho.Field("error", err),
			gecho.Field("product_name", product.Name),
			gecho.Field("duration", time.Since(startTime)),
		)
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	// Insert images if any
	if len(images) > 0 {
		for i := range images {
			// Generate UUID for image if not set
			if images[i].ID == uuid.Nil {
				images[i].ID = uuid.New()
			}
			images[i].ProductID = product.ID
		}

		_, imgErr := database.Query[tables.ProductImage](ps.db).InsertMany(ctx, images)
		if imgErr != nil {
			ps.logger.Error("Failed to insert product images",
				gecho.Field("error", imgErr),
				gecho.Field("product_id", product.ID),
			)
			return nil, fmt.Errorf("failed to insert product images: %w", imgErr)
		}
	}

	// Restore images to the product object for the response
	product.Images = images

	// Invalidate product caches asynchronously
	go func() {
		if err := ps.cacheService.InvalidateProductCaches(product.ID); err != nil {
			ps.logger.Warn("Failed to invalidate product caches after creation",
				gecho.Field("error", err),
				gecho.Field("product_id", product.ID),
			)
		}
	}()

	ps.logger.Info("Product created successfully",
		gecho.Field("id", product.ID),
		gecho.Field("image_count", len(images)),
		gecho.Field("duration", time.Since(startTime)),
	)
	return product, nil
}

type UpdateProductRequest struct {
	Name        *string               `json:"name,omitempty"`
	SKU         *string               `json:"sku,omitempty"`
	Price       *uint64               `json:"price,omitempty"`
	Discount    *uint64               `json:"discount,omitempty"`
	Tax         *uint64               `json:"tax,omitempty"`
	Description *string               `json:"description,omitempty"`
	IsActive    *bool                 `json:"is_active,omitempty"`
	Images      []tables.ProductImage `json:"images,omitempty"`
}

func (ps *ProductService) UpdateProduct(ctx context.Context, productID uuid.UUID, req *UpdateProductRequest) error {
	return database.Transaction(ps.db, ctx, func(tx bun.Tx) error {
		// Build update map with only provided fields
		updateData := make(map[string]any)

		if req.Name != nil {
			updateData["name"] = *req.Name
		}
		if req.SKU != nil {
			updateData["sku"] = *req.SKU
		}
		if req.Price != nil {
			updateData["price"] = *req.Price
		}
		if req.Discount != nil {
			updateData["discount"] = *req.Discount
		}
		if req.Tax != nil {
			updateData["tax"] = *req.Tax
		}
		if req.Description != nil {
			updateData["description"] = *req.Description
		}
		if req.IsActive != nil {
			updateData["is_active"] = *req.IsActive
		}

		// Handle images update if provided
		if req.Images != nil {
			// Delete existing images
			if _, err := database.Query[tables.ProductImage](ps.db).Where("product_id", productID.String()).Delete(ctx); err != nil {
				return fmt.Errorf("failed to delete existing images: %w", err)
			}

			// Insert new images if any provided
			if len(req.Images) > 0 {
				hasPrimary := false
				for i := range req.Images {
					if req.Images[i].ID == uuid.Nil {
						req.Images[i].ID = uuid.New()
					}
					req.Images[i].ProductID = productID
					if req.Images[i].IsPrimary {
						if hasPrimary {
							req.Images[i].IsPrimary = false
						} else {
							hasPrimary = true
						}
					}
				}

				if !hasPrimary && len(req.Images) > 0 {
					req.Images[0].IsPrimary = true
				}

				if _, err := ps.db.NewInsert().Model(&req.Images).Exec(ctx); err != nil {
					return fmt.Errorf("failed to insert new images: %w", err)
				}
			}
		}

		// Calculate subtotal if price, discount, or tax changed
		if req.Price != nil || req.Discount != nil || req.Tax != nil {
			currentProduct, err := database.Query[tables.Product](ps.db).Where("id", productID).First(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch current product for subtotal calculation: %w", err)
			}

			price := currentProduct.Price
			discount := currentProduct.Discount
			tax := currentProduct.Tax

			if req.Price != nil {
				price = *req.Price
			}
			if req.Discount != nil {
				discount = *req.Discount
			}
			if req.Tax != nil {
				tax = *req.Tax
			}
			updateData["subtotal"] = price - discount + tax
		}

		// Perform the update if there is data to update
		if len(updateData) > 0 {
			if _, err := database.Query[tables.Product](ps.db).Where("id", productID).Update(ctx, updateData); err != nil {
				return fmt.Errorf("failed to update product: %w", err)
			}
		}

		// Invalidate product caches asynchronously
		go func() {
			if err := ps.cacheService.InvalidateProductCaches(productID); err != nil {
				ps.logger.Warn("Failed to invalidate product caches after update",
					gecho.Field("error", err),
					gecho.Field("product_id", productID),
				)
			}
		}()

		return nil
	})
}
