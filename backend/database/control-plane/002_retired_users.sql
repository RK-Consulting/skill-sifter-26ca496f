-- Permanent lifecycle ledger for tenant users whose accounts are retired.
-- User IDs are nullable only for legacy registrations whose tenant DB was already removed.
CREATE TABLE IF NOT EXISTS platform_retired_users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL,
    user_id BIGINT,
    username VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    status VARCHAR(30) NOT NULL,
    retired_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_retired_users_status_valid
        CHECK (status IN ('TRIAL_EXPIRED', 'ACCOUNT_DELETED', 'DISCONTINUED', 'LEGACY_UNKNOWN')),
    CONSTRAINT platform_retired_users_tenant_email_unique UNIQUE (tenant_id, email)
);

CREATE INDEX IF NOT EXISTS idx_platform_retired_users_status
    ON platform_retired_users(status, retired_at DESC);

CREATE INDEX IF NOT EXISTS idx_platform_retired_users_email
    ON platform_retired_users(lower(email));

-- Preserve the identities of registrations whose tenant rows were already
-- removed before this ledger existed. Their former tenant-local user IDs and
-- usernames cannot be reconstructed, so make that limitation explicit.
INSERT INTO platform_retired_users (tenant_id, username, email, status, retired_at)
SELECT rr.last_tenant_id, rr.email_id, rr.email_id, 'LEGACY_UNKNOWN',
       COALESCE(rr.first_registered, NOW())
FROM platform_registration_registry rr
LEFT JOIN platform_tenants pt ON pt.tenant_id = rr.last_tenant_id
WHERE rr.last_tenant_id IS NOT NULL
  AND pt.tenant_id IS NULL
ON CONFLICT (tenant_id, email) DO NOTHING;
