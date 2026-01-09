package products

import (
	"mamabloemetjes_server/handling"
	"mamabloemetjes_server/lib"
	"net/http"
	"strconv"

	"github.com/MonkyMars/gecho"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// FetchAllProducts handles GET /products with comprehensive filtering, pagination, and sorting
func (p *ProductRoutesManager) FetchAllProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters into options
	opts, err := handling.ParseProductListOptions(r)
	if err != nil {
		p.logger.Warn("Invalid query parameters", "error", err)
		gecho.BadRequest(w,
			gecho.WithMessage("error.invalidQueryParameters"),
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
			gecho.WithMessage("error.products.failedToFetch"),
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
	idStr := chi.URLParam(r, "id")

	// Validate and parse ID
	id, err := uuid.Parse(idStr)
	if err != nil {
		p.logger.Warn("Invalid product ID format", "id", idStr, "error", err)
		gecho.BadRequest(w,
			gecho.WithMessage("error.products.invalidProductId"),
			gecho.Send(),
		)
		return
	}

	// Check if ID is empty (zero UUID)
	if id == uuid.Nil {
		p.logger.Warn("Product ID not provided")
		gecho.BadRequest(w,
			gecho.WithMessage("error.products.productIdRequired"),
			gecho.Send(),
		)
		return
	}

	// Check if images should be included
	includeImages := r.URL.Query().Get("include_images") == "true"

	// Fetch product using the service
	product, err := p.productService.GetProductByID(ctx, id, includeImages)
	if err != nil {
		if err.Error() == "product not found" {
			gecho.NotFound(w,
				gecho.WithMessage("error.products.notFound"),
				gecho.Send(),
			)
			return
		}

		p.logger.Error("Failed to fetch product by ID", "id", id, "error", err)
		gecho.InternalServerError(w,
			gecho.WithMessage("error.products.failedToFetchOne"),
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

	if pageStr := lib.SanitizeString(r.URL.Query().Get("page"), true, false); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil && val > 0 {
			page = val
		}
	}

	if pageSizeStr := lib.SanitizeString(r.URL.Query().Get("page_size"), true, false); pageSizeStr != "" {
		if val, err := strconv.Atoi(pageSizeStr); err == nil && val > 0 {
			pageSize = val
		}
	}

	productType := lib.SanitizeString(r.URL.Query().Get("product_type"), true, false)

	// Check if images should be included
	includeImages := lib.SanitizeString(r.URL.Query().Get("include_images"), true, false) == "true"

	// Fetch active products using the service
	result, err := p.productService.GetActiveProducts(ctx, page, pageSize, includeImages, productType)
	if err != nil {
		p.logger.Error("Failed to fetch active products", "error", err)
		gecho.InternalServerError(w,
			gecho.WithMessage("error.products.failedToFetchActive"),
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
	opts, err := handling.ParseProductListOptions(r)
	if err != nil {
		p.logger.Warn("Invalid query parameters", "error", err)
		gecho.BadRequest(w,
			gecho.WithMessage("error.invalidQueryParameters"),
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
			gecho.WithMessage("error.products.failedToCount"),
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
