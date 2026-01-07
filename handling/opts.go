package handling

import (
	"mamabloemetjes_server/services"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// parseProductListOptions parses HTTP query parameters into ProductListOptions
func ParseProductListOptions(r *http.Request) (*services.ProductListOptions, error) {
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
