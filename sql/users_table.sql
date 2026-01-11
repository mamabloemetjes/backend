-- ============================================================================
-- Users Table Schema
-- ============================================================================
-- This schema is optimized for the User struct in Go with proper indexing
-- for authentication, authorization, and user management.
-- ============================================================================

-- Drop existing table if recreating (use with caution in production)
-- DROP TABLE IF EXISTS public.users CASCADE;

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.users (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- User Credentials
    username TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,

    -- Authorization
    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin', 'moderator')),

    -- Account Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    last_login TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()

    -- NOTE: Constraints removed to support encryption
    -- Email and username are now encrypted for privacy and cannot be validated with regex/length checks
) TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR USERS TABLE
-- ============================================================================

-- Primary lookup indexes
-- Username is non-unique (full names can be identical)
CREATE INDEX IF NOT EXISTS idx_users_username
    ON public.users USING btree (username)
    TABLESPACE pg_default;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email
    ON public.users USING btree (email)
    TABLESPACE pg_default;

-- NOTE: Case-insensitive indexes removed - fields are encrypted
-- LOWER() cannot be used on encrypted data

-- Active users composite index
CREATE INDEX IF NOT EXISTS idx_users_active_created
    ON public.users USING btree (is_active, created_at DESC)
    TABLESPACE pg_default;

-- Index for role-based queries
CREATE INDEX IF NOT EXISTS idx_users_role
    ON public.users USING btree (role, is_active)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for active users only (partial index)
CREATE INDEX IF NOT EXISTS idx_users_active_only
    ON public.users USING btree (created_at DESC)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for email verification status
CREATE INDEX IF NOT EXISTS idx_users_email_verified
    ON public.users USING btree (email_verified, created_at DESC)
    TABLESPACE pg_default
    WHERE is_active = true;

-- Index for last login tracking
CREATE INDEX IF NOT EXISTS idx_users_last_login
    ON public.users USING btree (last_login DESC NULLS LAST)
    TABLESPACE pg_default
    WHERE is_active = true;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to update updated_at timestamp automatically
CREATE OR REPLACE FUNCTION update_users_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on users
DROP TRIGGER IF EXISTS set_updated_at ON public.users;
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON public.users
    FOR EACH ROW
    EXECUTE FUNCTION update_users_updated_at_column();

-- NOTE: normalize_user_fields function and trigger removed
-- Fields are now encrypted and normalization happens in application layer before encryption

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON TABLE public.users IS
    'Main users table storing authentication and authorization information';

COMMENT ON COLUMN public.users.username IS
    'User full name for display purposes (2-100 characters, non-unique)';

COMMENT ON COLUMN public.users.email IS
    'Unique email address for authentication and communication';

COMMENT ON COLUMN public.users.password_hash IS
    'Argon hashed password - never store plain text passwords';

COMMENT ON COLUMN public.users.role IS
    'User role for authorization (user, admin, moderator)';

COMMENT ON COLUMN public.users.is_active IS
    'Account active status - false for deactivated/banned accounts';

COMMENT ON COLUMN public.users.email_verified IS
    'Whether the user has verified their email address';

COMMENT ON COLUMN public.users.last_login IS
    'Timestamp of the user''s last successful login';

-- ============================================================================
-- GRANTS (Adjust based on your user roles)
-- ============================================================================

-- Example: Grant permissions to application user
-- GRANT SELECT, INSERT, UPDATE, DELETE ON public.users TO your_app_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO your_app_user;

-- ============================================================================
-- ANALYTICS/MONITORING VIEWS (Optional but recommended)
-- ============================================================================

-- View for active users statistics
CREATE OR REPLACE VIEW v_active_users_stats AS
SELECT
    role,
    COUNT(*) as user_count,
    COUNT(CASE WHEN email_verified THEN 1 END) as verified_count,
    COUNT(CASE WHEN last_login IS NOT NULL THEN 1 END) as logged_in_count,
    COUNT(CASE WHEN last_login > NOW() - INTERVAL '30 days' THEN 1 END) as active_last_30_days
FROM public.users
WHERE is_active = true
GROUP BY role;

COMMENT ON VIEW v_active_users_stats IS
    'Statistics about active users grouped by role';

-- View for user activity monitoring
CREATE OR REPLACE VIEW v_user_activity AS
SELECT
    id,
    username,
    email,
    role,
    email_verified,
    last_login,
    CASE
        WHEN last_login IS NULL THEN 'Never logged in'
        WHEN last_login > NOW() - INTERVAL '7 days' THEN 'Active (7 days)'
        WHEN last_login > NOW() - INTERVAL '30 days' THEN 'Active (30 days)'
        WHEN last_login > NOW() - INTERVAL '90 days' THEN 'Inactive (90 days)'
        ELSE 'Inactive (90+ days)'
    END as activity_status,
    created_at
FROM public.users
WHERE is_active = true
ORDER BY last_login DESC NULLS LAST;

COMMENT ON VIEW v_user_activity IS
    'User activity overview with categorized activity status';

-- View for unverified emails
CREATE OR REPLACE VIEW v_unverified_users AS
SELECT
    id,
    username,
    email,
    created_at,
    EXTRACT(DAY FROM (NOW() - created_at)) as days_since_registration
FROM public.users
WHERE is_active = true
  AND email_verified = false
ORDER BY created_at ASC;

COMMENT ON VIEW v_unverified_users IS
    'Users who have not yet verified their email addresses';

-- ============================================================================
-- SECURITY NOTES
-- ============================================================================
-- 1. Always hash passwords using bcrypt or argon2 before storing
-- 2. Never expose password_hash in API responses (use json:"-" tag)
-- 3. Implement rate limiting on login attempts
-- 4. Consider adding a failed_login_attempts column for brute force protection
-- 5. Consider adding a password_reset_token and password_reset_expires for password recovery
-- 6. Use HTTPS/TLS for all authentication requests
-- 7. Implement proper session management and token expiration

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
