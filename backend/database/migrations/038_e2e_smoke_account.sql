-- 038_e2e_smoke_account.sql
-- Phase 9 production smoke fixture.
--
-- This is a permanent, non-payment-backed account used only by the production
-- browser smoke test. It deliberately uses the normal login/RBAC/subscription
-- path; it does not bypass authentication or subscription checks.
--
-- The password is stored only as a bcrypt hash. The corresponding smoke-test
-- credentials must be configured in GitHub Actions secrets.

INSERT INTO companies (id, name, created_at)
VALUES (
    'e2e_smoke_tenant',
    'SkillSifter E2E Smoke',
    NOW()
)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name;

INSERT INTO users (
    username,
    email,
    password,
    role,
    tenant_id,
    company_name,
    created_at
)
VALUES (
    'e2e-admin',
    'e2e-admin@skillsifter.in',
    '$2y$10$mb5vqkJ9EnJtYjVqMBeSru2Fk/9WwAgsK.GVWTWC8c4gM/tqELa5O',
    'admin',
    'e2e_smoke_tenant',
    'SkillSifter E2E Smoke',
    NOW()
)
ON CONFLICT (email) DO UPDATE
SET username = EXCLUDED.username,
    password = EXCLUDED.password,
    role = EXCLUDED.role,
    tenant_id = EXCLUDED.tenant_id,
    company_name = EXCLUDED.company_name;

INSERT INTO platform_tenants (
    tenant_id,
    company_name,
    account_status,
    provisioning_status,
    tenant_database,
    created_at,
    updated_at
)
VALUES (
    'e2e_smoke_tenant',
    'SkillSifter E2E Smoke',
    'ACTIVE',
    'READY',
    current_database(),
    NOW(),
    NOW()
)
ON CONFLICT (tenant_id) DO UPDATE
SET company_name = EXCLUDED.company_name,
    account_status = 'ACTIVE',
    provisioning_status = 'READY',
    tenant_database = current_database(),
    updated_at = NOW();

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
    'e2e_smoke_tenant',
    u.email,
    'admin',
    NOW(),
    NOW()
FROM users u
WHERE u.email = 'e2e-admin@skillsifter.in'
ON CONFLICT (user_id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    email = EXCLUDED.email,
    role = 'admin',
    updated_at = NOW();

UPDATE platform_subscriptions
SET status = 'CANCELLED',
    ends_at = COALESCE(ends_at, NOW()),
    updated_at = NOW()
WHERE tenant_id = 'e2e_smoke_tenant'
  AND status IN ('TRIAL', 'ACTIVE')
  AND id <> COALESCE((
      SELECT MAX(id)
      FROM platform_subscriptions
      WHERE tenant_id = 'e2e_smoke_tenant'
        AND plan_code = 'e2e-smoke'
  ), -1);

INSERT INTO platform_subscriptions (
    tenant_id,
    plan_code,
    status,
    starts_at,
    ends_at,
    provider,
    provider_subscription_ref,
    created_at,
    updated_at
)
VALUES (
    'e2e_smoke_tenant',
    'e2e-smoke',
    'ACTIVE',
    NOW(),
    NULL,
    'internal',
    'e2e-smoke',
    NOW(),
    NOW()
)
ON CONFLICT (provider, provider_subscription_ref)
WHERE provider IS NOT NULL AND provider_subscription_ref IS NOT NULL
DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    plan_code = EXCLUDED.plan_code,
    status = 'ACTIVE',
    ends_at = NULL,
    updated_at = NOW();
