-- ============================================================================
-- Order Lines Table Schema
-- ============================================================================
-- This schema is optimized for the OrderLine struct in Go with proper indexing
-- for high-performance queries and data integrity constraints.
-- Includes pricing snapshots to preserve historical order data.
-- ============================================================================

-- ============================================================================
-- ORDER LINES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.order_lines (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Keys
    order_id UUID NOT NULL,
    product_id UUID NOT NULL,

    -- Quantity
    quantity INTEGER NOT NULL,

    -- Snapshot of pricing at time of order (stored in cents)
    unit_price BIGINT NOT NULL CHECK (unit_price >= 0),
    unit_discount BIGINT NOT NULL DEFAULT 0 CHECK (unit_discount >= 0),
    unit_tax BIGINT NOT NULL DEFAULT 0 CHECK (unit_tax >= 0),
    unit_subtotal BIGINT NOT NULL CHECK (unit_subtotal >= 0),
    line_total BIGINT NOT NULL CHECK (line_total >= 0),

    -- Snapshot of product details at time of order
    product_name TEXT NOT NULL,
    product_sku TEXT NOT NULL,

    -- Constraints
    CONSTRAINT check_quantity_positive CHECK (quantity > 0),
    CONSTRAINT check_unit_subtotal_calculation CHECK (
        unit_subtotal = (unit_price - unit_discount + unit_tax)
    ),
    CONSTRAINT check_line_total_calculation CHECK (
        line_total = (unit_subtotal * quantity)
    ),

    -- Foreign Key Constraints
    CONSTRAINT order_lines_order_id_fkey
        FOREIGN KEY (order_id)
        REFERENCES public.orders (id)
        ON DELETE CASCADE,

    CONSTRAINT order_lines_product_id_fkey
        FOREIGN KEY (product_id)
        REFERENCES public.products (id)
        ON DELETE RESTRICT,

    -- Ensure unique product per order (prevent duplicate line items)
    CONSTRAINT order_lines_order_product_unique
        UNIQUE (order_id, product_id)
) TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR ORDER LINES TABLE
-- ============================================================================

-- Foreign key index for order lookups (critical for JOIN performance)
CREATE INDEX IF NOT EXISTS idx_order_lines_order_id
    ON public.order_lines USING btree (order_id)
    TABLESPACE pg_default;

-- Foreign key index for product lookups
CREATE INDEX IF NOT EXISTS idx_order_lines_product_id
    ON public.order_lines USING btree (product_id)
    TABLESPACE pg_default;

-- Composite index for order-product queries
CREATE INDEX IF NOT EXISTS idx_order_lines_order_product
    ON public.order_lines USING btree (order_id, product_id)
    TABLESPACE pg_default;

-- Index for quantity-based queries and analytics
CREATE INDEX IF NOT EXISTS idx_order_lines_quantity
    ON public.order_lines USING btree (quantity DESC)
    TABLESPACE pg_default;

-- Index for line total analytics
CREATE INDEX IF NOT EXISTS idx_order_lines_line_total
    ON public.order_lines USING btree (line_total DESC)
    TABLESPACE pg_default;

-- Index for product SKU lookups in order history
CREATE INDEX IF NOT EXISTS idx_order_lines_product_sku
    ON public.order_lines USING btree (product_sku)
    TABLESPACE pg_default;

-- Full-text search index for product name and SKU
CREATE INDEX IF NOT EXISTS idx_order_lines_search
    ON public.order_lines USING gin (
        to_tsvector('english', product_name || ' ' || product_sku)
    )
    TABLESPACE pg_default;

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON TABLE public.order_lines IS
    'Order line items linking orders to products with quantities and pricing snapshots. Preserves historical pricing even if product prices change.';

COMMENT ON COLUMN public.order_lines.id IS
    'Unique identifier for the order line item (UUID)';

COMMENT ON COLUMN public.order_lines.order_id IS
    'Foreign key reference to the orders table';

COMMENT ON COLUMN public.order_lines.product_id IS
    'Foreign key reference to the products table';

COMMENT ON COLUMN public.order_lines.quantity IS
    'Quantity of the product ordered (must be positive)';

COMMENT ON COLUMN public.order_lines.unit_price IS
    'Product price at time of order in cents (snapshot)';

COMMENT ON COLUMN public.order_lines.unit_discount IS
    'Discount amount at time of order in cents (snapshot)';

COMMENT ON COLUMN public.order_lines.unit_tax IS
    'Tax amount at time of order in cents (snapshot)';

COMMENT ON COLUMN public.order_lines.unit_subtotal IS
    'Unit subtotal at time of order in cents: unit_price - unit_discount + unit_tax (snapshot)';

COMMENT ON COLUMN public.order_lines.line_total IS
    'Total for this line item in cents: unit_subtotal * quantity';

COMMENT ON COLUMN public.order_lines.product_name IS
    'Product name at time of order (snapshot, preserves history even if product is renamed)';

COMMENT ON COLUMN public.order_lines.product_sku IS
    'Product SKU at time of order (snapshot, preserves history even if SKU changes)';

COMMENT ON CONSTRAINT order_lines_order_product_unique ON public.order_lines IS
    'Ensures each product appears only once per order (update quantity instead of adding duplicates)';

-- ============================================================================
-- ANALYTICS/MONITORING VIEWS (Optional but recommended)
-- ============================================================================

-- View for order details with complete information
CREATE OR REPLACE VIEW v_order_details AS
SELECT
    o.id as order_id,
    o.order_number,
    o.name as customer_name,
    o.email as customer_email,
    o.phone as customer_phone,
    o.status as order_status,
    o.payment_status,
    o.created_at as order_date,
    ol.id as line_item_id,
    ol.product_id,
    ol.product_name,
    ol.product_sku,
    ol.quantity,
    ol.unit_price,
    ol.unit_discount,
    ol.unit_tax,
    ol.unit_subtotal,
    ol.line_total
FROM public.orders o
INNER JOIN public.order_lines ol ON o.id = ol.order_id
WHERE o.deleted_at IS NULL;

COMMENT ON VIEW v_order_details IS
    'Comprehensive view of active orders with complete line item details including pricing snapshots';

-- View for product sales statistics
CREATE OR REPLACE VIEW v_product_sales_stats AS
SELECT
    ol.product_id,
    ol.product_name,
    ol.product_sku,
    COUNT(DISTINCT ol.order_id) as times_ordered,
    SUM(ol.quantity) as total_quantity_sold,
    SUM(ol.line_total) as total_revenue,
    AVG(ol.unit_subtotal) as average_unit_price,
    MIN(ol.unit_subtotal) as lowest_unit_price,
    MAX(ol.unit_subtotal) as highest_unit_price
FROM public.order_lines ol
INNER JOIN public.orders o ON ol.order_id = o.id
WHERE o.deleted_at IS NULL
  AND o.payment_status = 'paid'
GROUP BY ol.product_id, ol.product_name, ol.product_sku
ORDER BY total_quantity_sold DESC NULLS LAST;

COMMENT ON VIEW v_product_sales_stats IS
    'Sales statistics per product including order count, quantity sold, and revenue (paid orders only)';

-- View for top selling products (last 30 days)
CREATE OR REPLACE VIEW v_top_selling_products_30d AS
SELECT
    ol.product_id,
    ol.product_name,
    ol.product_sku,
    COUNT(DISTINCT ol.order_id) as order_count,
    SUM(ol.quantity) as total_quantity,
    SUM(ol.line_total) as total_revenue
FROM public.order_lines ol
INNER JOIN public.orders o ON ol.order_id = o.id
WHERE o.created_at >= NOW() - INTERVAL '30 days'
  AND o.deleted_at IS NULL
  AND o.payment_status = 'paid'
GROUP BY ol.product_id, ol.product_name, ol.product_sku
ORDER BY total_quantity DESC
LIMIT 20;

COMMENT ON VIEW v_top_selling_products_30d IS
    'Top 20 best-selling products in the last 30 days based on quantity sold';

-- View for revenue analysis
CREATE OR REPLACE VIEW v_revenue_by_product AS
SELECT
    ol.product_id,
    ol.product_name,
    ol.product_sku,
    SUM(ol.line_total) as total_revenue,
    SUM(ol.quantity * ol.unit_price) as gross_revenue,
    SUM(ol.quantity * ol.unit_discount) as total_discounts,
    SUM(ol.quantity * ol.unit_tax) as total_tax,
    COUNT(DISTINCT ol.order_id) as number_of_orders,
    SUM(ol.quantity) as units_sold
FROM public.order_lines ol
INNER JOIN public.orders o ON ol.order_id = o.id
WHERE o.deleted_at IS NULL
  AND o.payment_status = 'paid'
GROUP BY ol.product_id, ol.product_name, ol.product_sku
ORDER BY total_revenue DESC;

COMMENT ON VIEW v_revenue_by_product IS
    'Detailed revenue breakdown by product including discounts and tax (paid orders only)';

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
