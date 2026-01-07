-- ============================================================================
-- Orders Table Schema
-- ============================================================================
-- This schema is optimized for the Order struct in Go with proper indexing
-- for high-performance queries and data integrity constraints.
-- ============================================================================

-- Drop existing table if recreating (use with caution in production)
-- DROP TABLE IF EXISTS public.orders CASCADE;

-- ============================================================================
-- ENUMS FOR ORDER AND PAYMENT STATUS
-- ============================================================================

-- Create enum types for order status
DO $$ BEGIN
    CREATE TYPE order_status AS ENUM (
        'pending',
        'paid',
        'processing',
        'shipped',
        'delivered',
        'cancelled',
        'refunded'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Create enum types for payment status
DO $$ BEGIN
    CREATE TYPE payment_status AS ENUM (
        'unpaid',
        'paid'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- ============================================================================
-- ORDERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.orders (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Order Information
    order_number TEXT NOT NULL UNIQUE,

    -- Customer Data
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    phone TEXT NOT NULL,
    note TEXT, -- Customer notes (nullable)

    -- Address Data (Reference to Address table)
    address_id UUID NOT NULL,

    -- Payment Data
    payment_link TEXT, -- Nullable initially, attached later
    payment_status payment_status NOT NULL DEFAULT 'unpaid',

    -- Order Status
    status order_status NOT NULL DEFAULT 'pending',

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE, -- Soft delete support

    -- Constraints
    -- NOTE: Constraints on name, email, phone, and note removed to support encryption
    -- These fields are now encrypted for privacy and cannot be validated with length checks
    CONSTRAINT check_order_number_not_empty CHECK (char_length(order_number) > 0),

    -- Foreign Key Constraint to addresses table
    CONSTRAINT orders_address_id_fkey
        FOREIGN KEY (address_id)
        REFERENCES public.addresses (id)
        ON DELETE RESTRICT
) TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR ORDERS TABLE
-- ============================================================================

-- Primary lookup indexes
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_number
    ON public.orders USING btree (order_number)
    TABLESPACE pg_default;

-- NOTE: Indexes on encrypted fields (email, name, phone) removed
-- Cannot index encrypted data - queries on these fields must be performed in application layer

-- Index for recent orders queries (excluding soft-deleted)
CREATE INDEX IF NOT EXISTS idx_orders_created_at
    ON public.orders USING btree (created_at DESC)
    TABLESPACE pg_default
    WHERE deleted_at IS NULL;

-- Index for address lookups
CREATE INDEX IF NOT EXISTS idx_orders_address_id
    ON public.orders USING btree (address_id)
    TABLESPACE pg_default;

-- Status-based queries (active orders only)
CREATE INDEX IF NOT EXISTS idx_orders_status_created
    ON public.orders USING btree (status, created_at DESC)
    TABLESPACE pg_default
    WHERE deleted_at IS NULL;

-- Payment status queries
CREATE INDEX IF NOT EXISTS idx_orders_payment_status
    ON public.orders USING btree (payment_status, created_at DESC)
    TABLESPACE pg_default
    WHERE deleted_at IS NULL;

-- Composite index for status and payment filtering
CREATE INDEX IF NOT EXISTS idx_orders_status_payment
    ON public.orders USING btree (status, payment_status, created_at DESC)
    TABLESPACE pg_default
    WHERE deleted_at IS NULL;

-- Soft delete support - index for non-deleted records
CREATE INDEX IF NOT EXISTS idx_orders_not_deleted
    ON public.orders USING btree (id, created_at DESC)
    TABLESPACE pg_default
    WHERE deleted_at IS NULL;

-- Index for deleted orders (for recovery/audit)
CREATE INDEX IF NOT EXISTS idx_orders_deleted_at
    ON public.orders USING btree (deleted_at DESC)
    TABLESPACE pg_default
    WHERE deleted_at IS NOT NULL;

-- NOTE: Full-text search index removed - cannot search encrypted fields
-- Customer data (name, email, phone, note) is encrypted
-- Search must be performed in application layer after decryption
-- Only order_number remains searchable
CREATE INDEX IF NOT EXISTS idx_orders_order_number_search
    ON public.orders USING gin (
        to_tsvector('english', COALESCE(order_number, ''))
    )
    TABLESPACE pg_default
    WHERE deleted_at IS NULL;

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

-- Trigger to automatically update updated_at on orders
DROP TRIGGER IF EXISTS set_updated_at ON public.orders;
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON public.orders
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON TABLE public.orders IS
    'Orders table storing customer order information, payment details, and delivery info. Supports both guest and registered users.';

COMMENT ON COLUMN public.orders.id IS
    'Unique identifier for the order (UUID)';

COMMENT ON COLUMN public.orders.order_number IS
    'Human-readable unique order number for customer reference';

COMMENT ON COLUMN public.orders.name IS
    'Customer name associated with the order (ENCRYPTED for privacy)';

COMMENT ON COLUMN public.orders.email IS
    'Customer email address for order notifications (ENCRYPTED for privacy)';

COMMENT ON COLUMN public.orders.phone IS
    'Customer phone number for delivery contact (ENCRYPTED for privacy)';

COMMENT ON COLUMN public.orders.note IS
    'Optional customer notes or special requests for the order (ENCRYPTED for privacy)';

COMMENT ON COLUMN public.orders.address_id IS
    'Foreign key reference to the addresses table for delivery location';

COMMENT ON COLUMN public.orders.payment_link IS
    'Payment gateway link (e.g., Mollie payment URL) - can be null initially';

COMMENT ON COLUMN public.orders.payment_status IS
    'Payment status: unpaid or paid';

COMMENT ON COLUMN public.orders.status IS
    'Order lifecycle status: pending, paid, processing, shipped, delivered, cancelled, or refunded';

COMMENT ON COLUMN public.orders.created_at IS
    'Timestamp when the order was created';

COMMENT ON COLUMN public.orders.updated_at IS
    'Timestamp when the order was last updated';

COMMENT ON COLUMN public.orders.deleted_at IS
    'Soft delete timestamp - NULL for active orders, set to deletion time for deleted orders';

-- ============================================================================
-- ANALYTICS/MONITORING VIEWS (Optional but recommended)
-- ============================================================================

-- View for active orders with basic statistics
CREATE OR REPLACE VIEW v_active_orders AS
SELECT
    o.id,
    o.order_number,
    o.name,
    o.email,
    o.phone,
    o.status,
    o.payment_status,
    o.created_at,
    o.updated_at,
    COUNT(ol.id) as total_items,
    SUM(ol.line_total) as total_amount
FROM public.orders o
LEFT JOIN public.order_lines ol ON o.id = ol.order_id
WHERE o.deleted_at IS NULL
GROUP BY o.id
ORDER BY o.created_at DESC;

COMMENT ON VIEW v_active_orders IS
    'View showing active (non-deleted) orders with item counts and total amounts';

-- View for recent orders (last 30 days)
CREATE OR REPLACE VIEW v_recent_orders AS
SELECT
    o.id,
    o.order_number,
    o.name,
    o.email,
    o.phone,
    o.status,
    o.payment_status,
    o.created_at,
    o.updated_at,
    COUNT(ol.id) as total_items,
    SUM(ol.line_total) as total_amount
FROM public.orders o
LEFT JOIN public.order_lines ol ON o.id = ol.order_id
WHERE o.created_at >= NOW() - INTERVAL '30 days'
  AND o.deleted_at IS NULL
GROUP BY o.id
ORDER BY o.created_at DESC;

COMMENT ON VIEW v_recent_orders IS
    'View showing recent orders (last 30 days) with item counts and total amounts';

-- View for pending payment orders
CREATE OR REPLACE VIEW v_pending_payment_orders AS
SELECT
    o.id,
    o.order_number,
    o.name,
    o.email,
    o.phone,
    o.payment_link,
    o.created_at,
    EXTRACT(HOUR FROM (NOW() - o.created_at)) as hours_pending
FROM public.orders o
WHERE o.payment_status = 'unpaid'
  AND o.status = 'pending'
  AND o.deleted_at IS NULL
ORDER BY o.created_at ASC;

COMMENT ON VIEW v_pending_payment_orders IS
    'Orders awaiting payment with time elapsed since creation';

-- View for order status statistics
CREATE OR REPLACE VIEW v_order_status_stats AS
SELECT
    status,
    payment_status,
    COUNT(*) as order_count,
    SUM(total_amount) as total_revenue
FROM (
    SELECT
        o.id,
        o.status,
        o.payment_status,
        COALESCE(SUM(ol.line_total), 0) as total_amount
    FROM public.orders o
    LEFT JOIN public.order_lines ol ON o.id = ol.order_id
    WHERE o.deleted_at IS NULL
    GROUP BY o.id, o.status, o.payment_status
) order_totals
GROUP BY status, payment_status
ORDER BY status, payment_status;

COMMENT ON VIEW v_order_status_stats IS
    'Statistics of orders grouped by status and payment status';

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
