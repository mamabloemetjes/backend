-- ============================================================================
-- Products Table Schema
-- ============================================================================
-- This schema is optimized for the Product struct in Go with proper indexing
-- for high-performance queries and data integrity constraints.
-- ============================================================================

-- Drop existing table if recreating (use with caution in production)
-- DROP TABLE IF EXISTS public.product_images CASCADE;
-- DROP TABLE IF EXISTS public.products CASCADE;

-- ============================================================================
-- PRODUCTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.products (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Product Information
    name TEXT NOT NULL,
    sku TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,

    -- Pricing (stored in cents for precision)
    price BIGINT NOT NULL CHECK (price >= 0),
    discount BIGINT NOT NULL DEFAULT 0 CHECK (discount >= 0),
    tax BIGINT NOT NULL DEFAULT 0 CHECK (tax >= 0),
    subtotal BIGINT NOT NULL DEFAULT 0 CHECK (subtotal >= 0),

    -- Product Attributes
    size TEXT CHECK (size IN ('small', 'medium', 'large', '')),
    colors TEXT[] NOT NULL DEFAULT '{}',
    product_type TEXT CHECK (product_type IN ('flower', 'bouquet', '')),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraint: Ensure pricing calculation is valid
    CONSTRAINT check_price_calculation CHECK (
        ABS(price - (subtotal + tax)) < 1
    )
) TABLESPACE pg_default;

-- ============================================================================
-- PRODUCT IMAGES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.product_images (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key to Products
    product_id UUID NOT NULL,

    -- Image Information
    url TEXT NOT NULL,
    alt_text TEXT,
    is_primary BOOLEAN NOT NULL DEFAULT false,

    -- Foreign Key Constraint with CASCADE delete
    CONSTRAINT product_images_product_id_fkey
        FOREIGN KEY (product_id)
        REFERENCES public.products (id)
        ON DELETE CASCADE
) TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR PRODUCTS TABLE
-- ============================================================================

-- Primary lookup indexes
CREATE INDEX IF NOT EXISTS idx_products_sku
    ON public.products USING btree (sku)
    TABLESPACE pg_default;

CREATE INDEX IF NOT EXISTS idx_products_name
    ON public.products USING btree (name)
    TABLESPACE pg_default;

-- Active products composite index (most common query)
CREATE INDEX IF NOT EXISTS idx_products_active_created
    ON public.products USING btree (is_active, created_at DESC)
    TABLESPACE pg_default;

-- Covering index for product listings (includes common columns)
CREATE INDEX IF NOT EXISTS idx_products_listing_cover
    ON public.products USING btree (is_active, created_at DESC)
    INCLUDE (id, name, sku, price, description, stock)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for active products only (partial index)
CREATE INDEX IF NOT EXISTS idx_products_active_only
    ON public.products USING btree (created_at DESC, id)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for inventory management
CREATE INDEX IF NOT EXISTS idx_products_inventory
    ON public.products USING btree (stock, is_active)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for price range queries
CREATE INDEX IF NOT EXISTS idx_products_price
    ON public.products USING btree (price)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for product type filtering
CREATE INDEX IF NOT EXISTS idx_products_type
    ON public.products USING btree (product_type, is_active)
    TABLESPACE pg_default
    WHERE product_type IS NOT NULL AND product_type != '';

-- Index for size filtering
CREATE INDEX IF NOT EXISTS idx_products_size
    ON public.products USING btree (size, is_active)
    TABLESPACE pg_default
    WHERE size IS NOT NULL AND size != '';

-- GIN index for array operations on colors
CREATE INDEX IF NOT EXISTS idx_products_colors_gin
    ON public.products USING gin (colors)
    TABLESPACE pg_default;

-- Full-text search index for name and description
CREATE INDEX IF NOT EXISTS idx_products_search
    ON public.products USING gin (
        to_tsvector('english', name || ' ' || description || ' ' || sku)
    )
    TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR PRODUCT IMAGES TABLE
-- ============================================================================

-- Foreign key index (critical for JOIN performance)
CREATE INDEX IF NOT EXISTS idx_product_images_product_id
    ON public.product_images USING btree (product_id)
    TABLESPACE pg_default;

-- Composite index for finding primary images
CREATE INDEX IF NOT EXISTS idx_product_images_product_primary
    ON public.product_images USING btree (product_id, is_primary DESC)
    TABLESPACE pg_default;

-- Index for primary images only (partial index)
CREATE INDEX IF NOT EXISTS idx_product_images_primary_only
    ON public.product_images USING btree (product_id)
    TABLESPACE pg_default
    WHERE is_primary = true;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on products
DROP TRIGGER IF EXISTS set_updated_at ON public.products;
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON public.products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to ensure only one primary image per product
CREATE OR REPLACE FUNCTION ensure_single_primary_image()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_primary = true THEN
        -- Set all other images for this product to non-primary
        UPDATE public.product_images
        SET is_primary = false
        WHERE product_id = NEW.product_id
          AND id != NEW.id
          AND is_primary = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to ensure only one primary image per product
DROP TRIGGER IF EXISTS ensure_primary_image ON public.product_images;
CREATE TRIGGER ensure_primary_image
    BEFORE INSERT OR UPDATE ON public.product_images
    FOR EACH ROW
    WHEN (NEW.is_primary = true)
    EXECUTE FUNCTION ensure_single_primary_image();

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON TABLE public.products IS
    'Main products table storing all product information including flowers and bouquets';

COMMENT ON COLUMN public.products.price IS
    'Product price stored in cents for precision (e.g., $19.99 = 1999)';

COMMENT ON COLUMN public.products.discount IS
    'Discount amount in cents';

COMMENT ON COLUMN public.products.tax IS
    'Tax amount in cents';

COMMENT ON COLUMN public.products.subtotal IS
    'Subtotal in cents (price - discount + tax)';

COMMENT ON COLUMN public.products.sku IS
    'Stock Keeping Unit - unique identifier for inventory management';

COMMENT ON COLUMN public.products.colors IS
    'Array of color values for the product';

COMMENT ON COLUMN public.products.stock IS
    'Current inventory stock level';

COMMENT ON TABLE public.product_images IS
    'Product images with support for multiple images per product';

COMMENT ON COLUMN public.product_images.is_primary IS
    'Indicates if this is the primary/featured image for the product';

-- ============================================================================
-- GRANTS (Adjust based on your user roles)
-- ============================================================================

-- Example: Grant permissions to application user
-- GRANT SELECT, INSERT, UPDATE, DELETE ON public.products TO your_app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON public.product_images TO your_app_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO your_app_user;

-- ============================================================================
-- ANALYTICS/MONITORING VIEWS (Optional but recommended)
-- ============================================================================

-- View for active products with image count
CREATE OR REPLACE VIEW v_products_with_image_count AS
SELECT
    p.*,
    COUNT(pi.id) as image_count,
    MAX(CASE WHEN pi.is_primary THEN pi.url END) as primary_image_url
FROM public.products p
LEFT JOIN public.product_images pi ON p.id = pi.product_id
GROUP BY p.id;

COMMENT ON VIEW v_products_with_image_count IS
    'Convenience view showing products with their image counts and primary image URL';

-- View for low stock alerts
CREATE OR REPLACE VIEW v_low_stock_products AS
SELECT
    id,
    name,
    sku,
    stock,
    is_active
FROM public.products
WHERE is_active = true
  AND stock <= 10
ORDER BY stock ASC, name ASC;

COMMENT ON VIEW v_low_stock_products IS
    'Products with low inventory (10 or fewer items in stock)';

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
