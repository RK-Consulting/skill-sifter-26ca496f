-- 035_control_plane_subscription.sql
-- Phase 9A: introduce the platform/control-plane account layer.
--
-- This is the compatibility bridge from the current shared database model.
-- The platform_* tables are the authoritative control-plane contract for
-- tenant/account/subscription access. Physical separation into a dedicated
-- control-plane database is Phase 9B/9C and must not change this contract.

CREATE TABLE IF NOT EXISTS platform_tenants (
    tenant_id VARCHAR(255) PRIMARY KEY,
    company_name VARCHAR(255) NOT NULL UNIQUE,
    account_status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    provisioning_status VARCHAR(30) NOT NULL DEFAULT 'READY',
    tenant_database VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_tenants_account_status_chk
        CHECK (account_status IN ('ACTIVE', 'SUSPENDED', 'EXPIRED', 'TERMINATED')),
    CONSTRAINT platform_tenants_provisioning_status_chk
        CHECK (provisioning_status IN ('PENDING', 'PROVISIONING', 'READY', 'FAILED'))
);

CREATE TABLE IF NOT EXISTS platform_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES platform_tenants(tenant_id),
    plan_code VARCHAR(100) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'ACTIVE',
    starts_at TIMESTAMP NOT NULL DEFAULT NOW(),
    ends_at TIMESTAMP,
    provider VARCHAR(100),
    provider_subscription_ref VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT platform_subscriptions_status_chk
        CHECK (status IN ('TRIAL', 'ACTIVE', 'PAST_DUE', 'SUSPENDED', 'CANCELLED', 'EXPIRED')),
    CONSTRAINT platform_subscriptions_dates_chk
        CHECK (ends_at IS NULL OR ends_at >= starts_at)
);

CREATE INDEX IF NOT EXISTS idx_platform_subscriptions_tenant
    ON platform_subscriptions(tenant_id);

CREATE INDEX IF NOT EXISTS idx_platform_subscriptions_access
    ON platform_subscriptions(tenant_id, status);

CREATE UNIQUE INDEX IF NOT EXISTS uq_platform_subscriptions_provider_ref
    ON platform_subscriptions(provider, provider_subscription_ref)
    WHERE provider IS NOT NULL AND provider_subscription_ref IS NOT NULL;

CREATE TABLE IF NOT EXISTS platform_user_accounts (
    user_id INTEGER PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    tenant_id VARCHAR(255) NOT NULL REFERENCES platform_tenants(tenant_id),
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_platform_user_accounts_tenant
    ON platform_user_accounts(tenant_id);

CREATE INDEX IF NOT EXISTS idx_platform_user_accounts_email
    ON platform_user_accounts(email);

-- Backfill the control-plane tenant record from the existing authoritative
-- companies table. Existing companies remain accessible during migration.
INSERT INTO platform_tenants (
    tenant_id,
    company_name,
    account_status,
    provisioning_status,
    created_at,
    updated_at
)
SELECT
    c.id,
    c.name,
    'ACTIVE',
    'READY',
    c.created_at,
    NOW()
FROM companies c
ON CONFLICT (tenant_id) DO UPDATE
SET company_name = EXCLUDED.company_name,
    updated_at = NOW();

-- Existing customers need a subscription record so the new login gate does
-- not lock out currently valid tenants. "legacy" is a compatibility plan;
-- it does not represent a payment provider or a financial transaction.
INSERT INTO platform_subscriptions (
    tenant_id,
    plan_code,
    status,
    starts_at,
    created_at,
    updated_at
)
SELECT
    t.tenant_id,
    'legacy',
    'ACTIVE',
    t.created_at,
    NOW(),
    NOW()
FROM platform_tenants t
WHERE NOT EXISTS (
    SELECT 1
    FROM platform_subscriptions s
    WHERE s.tenant_id = t.tenant_id
      AND s.status IN ('TRIAL', 'ACTIVE')
);

-- Mirror existing tenant users into the platform account index. The users
-- table remains the password source during this compatibility stage.
INSERT INTO platform_user_accounts (
    user_id,
    tenant_id,
    email,
    role,
    created_at,
    updated_at
)
SELECT
    u.id,
    u.tenant_id,
    u.email,
    u.role,
    u.created_at,
    NOW()
FROM users u
WHERE u.tenant_id IS NOT NULL
ON CONFLICT (user_id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    email = EXCLUDED.email,
    role = EXCLUDED.role,
    updated_at = NOW();

-- Validate that every existing user has a control-plane account and that the
-- account points at the same authoritative tenant identity.
DO $$
DECLARE
    unresolved INTEGER;
BEGIN
    SELECT COUNT(*)
    INTO unresolved
    FROM users u
    LEFT JOIN platform_user_accounts pua ON pua.user_id = u.id
    WHERE u.tenant_id IS NULL OR pua.user_id IS NULL OR pua.tenant_id <> u.tenant_id;

    IF unresolved > 0 THEN
        RAISE EXCEPTION 'control-plane user backfill incomplete: % user(s) are not mapped to their authoritative tenant', unresolved;
    END IF;
END $$;
