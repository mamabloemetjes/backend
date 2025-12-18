-- ============================================================================
-- EMAIL VERIFICATIONS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS public.email_verifications (
    -- Primary Key
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Foreign Key to Users
    user_id UUID NOT NULL,

    -- Verification Token
    token TEXT NOT NULL UNIQUE,

    -- Expiration & Status
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN NOT NULL DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Foreign Key Constraint with CASCADE delete
    CONSTRAINT email_verifications_user_id_fkey
        FOREIGN KEY (user_id)
        REFERENCES public.users (id)
        ON DELETE CASCADE,

    -- Ensure expiration is in the future when created
    CONSTRAINT check_expires_in_future
        CHECK (expires_at > created_at)
) TABLESPACE pg_default;

-- ============================================================================
-- INDEXES FOR EMAIL VERIFICATIONS TABLE
-- ============================================================================

-- Primary lookup index for token verification
CREATE UNIQUE INDEX IF NOT EXISTS idx_email_verifications_token
    ON public.email_verifications USING btree (token)
    TABLESPACE pg_default;

-- Foreign key index (critical for JOIN performance and CASCADE deletes)
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id
    ON public.email_verifications USING btree (user_id)
    TABLESPACE pg_default;

-- Composite index for finding valid tokens (most common query)
CREATE INDEX IF NOT EXISTS idx_email_verifications_valid
    ON public.email_verifications USING btree (used, expires_at DESC, token)
    TABLESPACE pg_default
    WHERE used = false;

-- Index for cleanup of expired tokens
CREATE INDEX IF NOT EXISTS idx_email_verifications_expires
    ON public.email_verifications USING btree (expires_at)
    TABLESPACE pg_default;

-- Index for user verification history
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_created
    ON public.email_verifications USING btree (user_id, created_at DESC)
    TABLESPACE pg_default;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Function to prevent reuse of verification tokens
CREATE OR REPLACE FUNCTION prevent_token_reuse()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.used = true THEN
        RAISE EXCEPTION 'Cannot modify a used verification token';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to prevent modifying used tokens
DROP TRIGGER IF EXISTS prevent_used_token_modification ON public.email_verifications;
CREATE TRIGGER prevent_used_token_modification
    BEFORE UPDATE ON public.email_verifications
    FOR EACH ROW
    WHEN (OLD.used = true)
    EXECUTE FUNCTION prevent_token_reuse();

-- ============================================================================
-- FUNCTIONS FOR TOKEN MANAGEMENT
-- ============================================================================

-- Function to clean up expired verification tokens
CREATE OR REPLACE FUNCTION cleanup_expired_verifications()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM public.email_verifications
    WHERE expires_at < NOW() - INTERVAL '7 days'
      OR (used = true AND created_at < NOW() - INTERVAL '30 days');

    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- COMMENTS (Documentation)
-- ============================================================================

COMMENT ON TABLE public.email_verifications IS
    'Stores email verification tokens for user registration and email change workflows';

COMMENT ON COLUMN public.email_verifications.token IS
    'Unique cryptographically secure token for email verification (recommend 32+ character random string)';

COMMENT ON COLUMN public.email_verifications.expires_at IS
    'Token expiration timestamp - typically 24-48 hours from creation';

COMMENT ON COLUMN public.email_verifications.used IS
    'Whether this token has been used - prevents replay attacks';

COMMENT ON FUNCTION mark_verification_token_used(TEXT) IS
    'Safely marks a verification token as used if it is still valid';

COMMENT ON FUNCTION cleanup_expired_verifications() IS
    'Removes expired (7+ days old) and used (30+ days old) verification tokens';

COMMENT ON FUNCTION verify_email_token(TEXT) IS
    'Validates a verification token and returns its status and associated user_id';

-- ============================================================================
-- ANALYTICS/MONITORING VIEWS (Optional but recommended)
-- ============================================================================

-- View for monitoring pending verifications
CREATE OR REPLACE VIEW v_pending_verifications AS
SELECT
    ev.id,
    ev.user_id,
    u.username,
    u.email,
    ev.created_at,
    ev.expires_at,
    EXTRACT(EPOCH FROM (ev.expires_at - NOW())) / 3600 as hours_until_expiry,
    CASE
        WHEN ev.expires_at < NOW() THEN 'Expired'
        WHEN ev.expires_at < NOW() + INTERVAL '6 hours' THEN 'Expiring Soon'
        ELSE 'Active'
    END as status
FROM public.email_verifications ev
JOIN public.users u ON ev.user_id = u.id
WHERE ev.used = false
ORDER BY ev.created_at DESC;

COMMENT ON VIEW v_pending_verifications IS
    'Shows all pending (unused) verification tokens with expiry information';

-- View for verification statistics
CREATE OR REPLACE VIEW v_verification_stats AS
SELECT
    COUNT(*) FILTER (WHERE used = false AND expires_at > NOW()) as pending_valid,
    COUNT(*) FILTER (WHERE used = false AND expires_at <= NOW()) as pending_expired,
    COUNT(*) FILTER (WHERE used = true) as used_total,
    COUNT(*) FILTER (WHERE used = true AND created_at > NOW() - INTERVAL '24 hours') as used_last_24h,
    COUNT(*) FILTER (WHERE created_at > NOW() - INTERVAL '24 hours') as created_last_24h
FROM public.email_verifications;

COMMENT ON VIEW v_verification_stats IS
    'Statistics about email verification tokens including pending, used, and timing metrics';

-- ============================================================================
-- MAINTENANCE
-- ============================================================================
-- Schedule cleanup_expired_verifications() to run daily via cron or pg_cron:
SELECT cron.schedule('cleanup-expired-verifications', '0 2 * * *',
    'SELECT cleanup_expired_verifications()');

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
