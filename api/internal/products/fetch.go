package products

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"mamabloemetjes_server/services"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
)

// FetchAllProducts handles GET /products with comprehensive filtering, pagination, and sorting
func (p *ProductRoutesManager) FetchAllProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters into options
	opts, err := p.parseProductListOptions(r)
	if err != nil {
		p.logger.Warn("Invalid query parameters", "error", err)
		gecho.BadRequest(w,
			gecho.WithMessage("Invalid query parameters"),
			gecho.WithData(err.Error()),
			gecho.Send(),
		)
		return
	}

	// Log the request
	p.logger.Debug("Fetching products",
		gecho.Field("include_images", opts.IncludeImages),
		gecho.Field("page", opts.Page),
		gecho.Field("page_size", opts.PageSize),
	)

	// Fetch products using the service
	result, err := p.productService.GetAllProducts(ctx, opts)
	if err != nil {
		p.logger.Error("Failed to fetch products", "error", err)
		gecho.InternalServerError(w,
			gecho.WithMessage("Failed to fetch products"),
			gecho.WithData(err.Error()),
			gecho.Send(),
		)
		return
	}

	// Return successful response with metadata
	gecho.Success(w,
		gecho.WithData(map[string]any{
			"products":   result.Products,
			"pagination": result.Pagination,
			"filters":    result.Filters,
			"meta": map[string]any{
				"query_time_ms": result.QueryTime.Milliseconds(),
				"count":         len(result.Products),
			},
		}),
		gecho.Send(),
	)
}

// FetchProductByID handles GET /products/{id} to fetch a single product
func (p *ProductRoutesManager) FetchProductByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get ID from URL parameter using chi
	id := chi.URLParam(r, "id")

	if id == "" {
		p.logger.Warn("Product ID not provided")
		gecho.BadRequest(w,
			gecho.WithMessage("Product ID is required"),
			gecho.Send(),
		)
		return
	}

	// Check if images should be included
	includeImages := r.URL.Query().Get("include_images") == "true"

	p.logger.Debug("Fetching product by ID", "id", id, "includeImages", includeImages)

	// Fetch product using the service
	product, err := p.productService.GetProductByID(ctx, id, includeImages)
	if err != nil {
		if err.Error() == "product not found" {
			gecho.NotFound(w,
				gecho.WithMessage("Product not found"),
				gecho.Send(),
			)
			return
		}

		p.logger.Error("Failed to fetch product by ID", "id", id, "error", err)
		gecho.InternalServerError(w,
			gecho.WithMessage("Failed to fetch product"),
			gecho.WithData(err.Error()),
			gecho.Send(),
		)
		return
	}

	// Return successful response
	gecho.Success(w,
		gecho.WithData(map[string]any{
			"product": product,
		}),
		gecho.Send(),
	)
}

// FetchActiveProducts handles GET /products/active to fetch only active products
func (p *ProductRoutesManager) FetchActiveProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 {
			pageSize = val
		}
	}

	// Check if images should be included
	includeImages := r.URL.Query().Get("include_images") == "true"

	// Fetch active products using the service
	result, err := p.productService.GetActiveProducts(ctx, page, pageSize, includeImages)
	if err != nil {
		p.logger.Error("Failed to fetch active products", "error", err)
		gecho.InternalServerError(w,
			gecho.WithMessage("Failed to fetch active products"),
			gecho.WithData(err.Error()),
			gecho.Send(),
		)
		return
	}

	// Return successful response with metadata
	gecho.Success(w,
		gecho.WithData(map[string]any{
			"products":   result.Products,
			"pagination": result.Pagination,
			"filters":    result.Filters,
			"meta": map[string]any{
				"query_time_ms": result.QueryTime.Milliseconds(),
				"count":         len(result.Products),
			},
		}),
		gecho.Send(),
	)
}

// GetProductCount handles GET /products/count to get total count of products
func (p *ProductRoutesManager) GetProductCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters into options (for filtering)
	opts, err := p.parseProductListOptions(r)
	if err != nil {
		p.logger.Warn("Invalid query parameters", "error", err)
		gecho.BadRequest(w,
			gecho.WithMessage("Invalid query parameters"),
			gecho.WithData(err.Error()),
			gecho.Send(),
		)
		return
	}

	// Get count using the service
	count, err := p.productService.GetProductCount(ctx, opts)
	if err != nil {
		p.logger.Error("Failed to count products", "error", err)
		gecho.InternalServerError(w,
			gecho.WithMessage("Failed to count products"),
			gecho.WithData(err.Error()),
			gecho.Send(),
		)
		return
	}

	// Return successful response
	gecho.Success(w,
		gecho.WithData(map[string]any{
			"count":   count,
			"filters": opts,
		}),
		gecho.Send(),
	)
}

// parseProductListOptions parses HTTP query parameters into ProductListOptions
func (p *ProductRoutesManager) parseProductListOptions(r *http.Request) (*services.ProductListOptions, error) {
	query := r.URL.Query()

	// Early return if no query params
	if len(query) == 0 {
		return &services.ProductListOptions{}, nil
	}

	opts := &services.ProductListOptions{}
	var err error
	var val64 uint64
	var valInt int
	var valBool bool

	// Parse pagination parameters
	if page := query.Get("page"); page != "" {
		if valInt, err = strconv.Atoi(page); err != nil {
			return nil, err
		}
		opts.Page = valInt
	}

	if pageSize := query.Get("page_size"); pageSize != "" {
		if valInt, err = strconv.Atoi(pageSize); err != nil {
			return nil, err
		}
		opts.PageSize = valInt
	}

	// Parse boolean filters
	if isActive := query.Get("is_active"); isActive != "" {
		if valBool, err = strconv.ParseBool(isActive); err != nil {
			return nil, err
		}
		opts.IsActive = &valBool
	}

	if inStock := query.Get("in_stock"); inStock != "" {
		if valBool, err = strconv.ParseBool(inStock); err != nil {
			return nil, err
		}
		opts.InStock = &valBool
	}

	// Parse string filters (no allocation needed)
	if productType := query.Get("product_type"); productType != "" {
		opts.ProductType = productType
	}

	if size := query.Get("size"); size != "" {
		opts.Size = size
	}

	if searchTerm := query.Get("search"); searchTerm != "" {
		opts.SearchTerm = searchTerm
	}

	// Parse price filters
	if minPrice := query.Get("min_price"); minPrice != "" {
		if val64, err = strconv.ParseUint(minPrice, 10, 64); err != nil {
			return nil, err
		}
		opts.MinPrice = &val64
	}

	if maxPrice := query.Get("max_price"); maxPrice != "" {
		if val64, err = strconv.ParseUint(maxPrice, 10, 64); err != nil {
			return nil, err
		}
		opts.MaxPrice = &val64
	}

	// Parse comma-separated lists
	if colors := query.Get("colors"); colors != "" {
		opts.Colors = splitAndTrim(colors)
	}

	if skus := query.Get("skus"); skus != "" {
		opts.SKUs = splitAndTrim(skus)
	}

	if excludeSKUs := query.Get("exclude_skus"); excludeSKUs != "" {
		opts.ExcludeSKUs = splitAndTrim(excludeSKUs)
	}

	// Parse date filters
	if createdAfter := query.Get("created_after"); createdAfter != "" {
		t, err := time.Parse(time.RFC3339, createdAfter)
		if err != nil {
			return nil, err
		}
		opts.CreatedAfter = &t
	}

	if createdBefore := query.Get("created_before"); createdBefore != "" {
		t, err := time.Parse(time.RFC3339, createdBefore)
		if err != nil {
			return nil, err
		}
		opts.CreatedBefore = &t
	}

	// Parse sorting parameters
	if sortBy := query.Get("sort_by"); sortBy != "" {
		opts.SortBy = sortBy
	}

	if sortDirection := query.Get("sort_direction"); sortDirection != "" {
		// Avoid allocation by converting in-place if needed
		opts.SortDirection = strings.ToUpper(sortDirection)
	}

	// Parse include_images flag
	if includeImages := query.Get("include_images"); includeImages != "" {
		if valBool, err = strconv.ParseBool(includeImages); err != nil {
			return nil, err
		}
		opts.IncludeImages = valBool
	}

	return opts, nil
}

// splitAndTrim splits a comma-separated string and trims whitespace efficiently
func splitAndTrim(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	// Trim in place to avoid extra allocations
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
