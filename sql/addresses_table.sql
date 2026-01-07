-- ============================================================================
-- Addresses Table Schema
-- ============================================================================
-- This schema is optimized for the Address struct in Go with proper indexing
-- for high-performance queries and data integrity constraints.
-- Supports both user-linked addresses and guest order addresses.
-- ============================================================================

-- Drop existing table if recreating (use with caution in production)
-- DROP TABLE IF EXISTS public.addresses CASCADE;

-- ============================================================================
-- ADDRESSES TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.addresses (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key to users (nullable for guest orders)
    user_id UUID,

    -- Address Information
    street TEXT NOT NULL,
    house_no TEXT NOT NULL,
    postal_code TEXT NOT NULL,
    city TEXT NOT NULL,
    country TEXT NOT NULL DEFAULT 'NL',

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- NOTE: Field constraints removed to support encryption
    -- street, house_no, postal_code, and city are now encrypted for privacy
    -- char_length checks cannot be performed on encrypted data

    -- Foreign Key Constraint to users table (nullable for guest checkouts)
    CONSTRAINT addresses_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES public.users (id)
        ON DELETE CASCADE
) TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR ADDRESSES TABLE
-- ============================================================================

-- Foreign key index for user lookups
CREATE INDEX IF NOT EXISTS idx_addresses_user_id
    ON public.addresses USING btree (user_id)
    TABLESPACE pg_default
    WHERE user_id IS NOT NULL;

-- NOTE: Indexes on encrypted fields (postal_code, city) removed
-- Cannot index encrypted data - queries on these fields will be slower
-- Consider using separate non-encrypted fields if searching is required

-- Index for country lookups (country is NOT encrypted)
CREATE INDEX IF NOT EXISTS idx_addresses_country
    ON public.addresses USING btree (country)
    TABLESPACE pg_default;

-- Index for user addresses ordered by creation
CREATE INDEX IF NOT EXISTS idx_addresses_user_created
    ON public.addresses USING btree (user_id, created_at DESC)
    TABLESPACE pg_default
    WHERE user_id IS NOT NULL;

-- Index for guest addresses (no user_id)
CREATE INDEX IF NOT EXISTS idx_addresses_guest
    ON public.addresses USING btree (created_at DESC)
    TABLESPACE pg_default
    WHERE user_id IS NULL;

-- NOTE: Full-text search index removed - cannot search encrypted fields
-- Address search must be performed in application layer after decryption

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_addresses_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on addresses
DROP TRIGGER IF EXISTS set_updated_at ON public.addresses;
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON public.addresses
    FOR EACH ROW
    EXECUTE FUNCTION update_addresses_updated_at_column();

-- NOTE: normalize_address_fields function and trigger removed
-- Address fields are now encrypted - normalization happens in application layer before encryption
-- Only country field can still be normalized (it remains unencrypted for regional statistics)
CREATE OR REPLACE FUNCTION normalize_country_field()
RETURNS TRIGGER AS $$
BEGIN
    NEW.country = UPPER(TRIM(NEW.country));
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to normalize country field only
DROP TRIGGER IF EXISTS normalize_fields ON public.addresses;
CREATE TRIGGER normalize_country
    BEFORE INSERT OR UPDATE ON public.addresses
    FOR EACH ROW
    EXECUTE FUNCTION normalize_country_field();

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON TABLE public.addresses IS
    'Addresses table storing delivery addresses for both registered users and guest orders. user_id is nullable to support guest checkout.';

COMMENT ON COLUMN public.addresses.id IS
    'Unique identifier for the address (UUID)';

COMMENT ON COLUMN public.addresses.user_id IS
    'Foreign key reference to users table - NULL for guest order addresses';

COMMENT ON COLUMN public.addresses.street IS
    'Street name';

COMMENT ON COLUMN public.addresses.house_no IS
    'House number (can include additions like 12A, 34-36, etc.)';

COMMENT ON COLUMN public.addresses.postal_code IS
    'Postal code / ZIP code - normalized to uppercase without spaces for NL addresses';

COMMENT ON COLUMN public.addresses.city IS
    'City name';

COMMENT ON COLUMN public.addresses.country IS
    'Country code (e.g., NL, BE, DE) - stored in uppercase';

COMMENT ON COLUMN public.addresses.created_at IS
    'Timestamp when the address was created';

COMMENT ON COLUMN public.addresses.updated_at IS
    'Timestamp when the address was last updated';

-- ============================================================================
-- ANALYTICS/MONITORING VIEWS (Optional but recommended)
-- ============================================================================

-- View for user addresses
CREATE OR REPLACE VIEW v_user_addresses AS
SELECT
    a.id,
    a.user_id,
    u.username,
    u.email,
    a.street,
    a.house_no,
    a.postal_code,
    a.city,
    a.country,
    a.created_at,
    a.updated_at,
    COUNT(o.id) as times_used_in_orders
FROM public.addresses a
INNER JOIN public.users u ON a.user_id = u.id
LEFT JOIN public.orders o ON a.id = o.address_id
WHERE a.user_id IS NOT NULL
GROUP BY a.id, u.username, u.email
ORDER BY a.created_at DESC;

COMMENT ON VIEW v_user_addresses IS
    'View showing addresses linked to registered users with order usage count';

-- View for guest addresses (used in orders)
CREATE OR REPLACE VIEW v_guest_addresses AS
SELECT
    a.id,
    a.street,
    a.house_no,
    a.postal_code,
    a.city,
    a.country,
    a.created_at,
    COUNT(o.id) as order_count,
    MAX(o.created_at) as last_order_date
FROM public.addresses a
INNER JOIN public.orders o ON a.id = o.address_id
WHERE a.user_id IS NULL
GROUP BY a.id
ORDER BY a.created_at DESC;

COMMENT ON VIEW v_guest_addresses IS
    'View showing addresses used by guest orders';

-- View for delivery zones (cities served)
CREATE OR REPLACE VIEW v_delivery_zones AS
SELECT
    a.city,
    a.country,
    COUNT(DISTINCT a.postal_code) as unique_postal_codes,
    COUNT(DISTINCT o.id) as total_orders,
    COUNT(DISTINCT CASE WHEN o.status = 'delivered' THEN o.id END) as delivered_orders,
    MAX(o.created_at) as last_order_date
FROM public.addresses a
INNER JOIN public.orders o ON a.id = o.address_id
WHERE o.deleted_at IS NULL
GROUP BY a.city, a.country
ORDER BY total_orders DESC;

COMMENT ON VIEW v_delivery_zones IS
    'Statistics about delivery zones showing which cities and postal codes are served';

-- View for most popular delivery addresses
CREATE OR REPLACE VIEW v_popular_delivery_addresses AS
SELECT
    a.street,
    a.house_no,
    a.postal_code,
    a.city,
    a.country,
    COUNT(o.id) as order_count,
    MAX(o.created_at) as last_delivery_date,
    ARRAY_AGG(DISTINCT o.name ORDER BY o.name) as customer_names
FROM public.addresses a
INNER JOIN public.orders o ON a.id = o.address_id
WHERE o.deleted_at IS NULL
GROUP BY a.street, a.house_no, a.postal_code, a.city, a.country
HAVING COUNT(o.id) > 1
ORDER BY order_count DESC
LIMIT 50;

COMMENT ON VIEW v_popular_delivery_addresses IS
    'Top 50 addresses that have received multiple deliveries';

-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Function to format address as single line string
CREATE OR REPLACE FUNCTION format_address(address_id UUID)
RETURNS TEXT AS $$
DECLARE
    result TEXT;
BEGIN
    SELECT
        street || ' ' || house_no || ', ' || postal_code || ' ' || city || ', ' || country
    INTO result
    FROM public.addresses
    WHERE id = address_id;

    RETURN result;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION format_address(UUID) IS
    'Formats an address as a single-line string for display purposes';

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
