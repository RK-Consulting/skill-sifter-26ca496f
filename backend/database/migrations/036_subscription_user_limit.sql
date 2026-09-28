-- 036_subscription_user_limit.sql
-- Phase 9: subscription controls the maximum number of tenant users.
--
-- One tenant always starts with exactly one administrator. Additional users
-- consume subscription seats and may only use the fixed operational roles:
-- manager, recruiter, team_leader.

ALTER TABLE platform_subscriptions
    ADD COLUMN IF NOT EXISTS user_limit INTEGER NOT NULL DEFAULT 1;

ALTER TABLE platform_subscriptions
    DROP CONSTRAINT IF EXISTS platform_subscriptions_user_limit_chk;

ALTER TABLE platform_subscriptions
    ADD CONSTRAINT platform_subscriptions_user_limit_chk
    CHECK (user_limit > 0);

-- Preserve access for existing tenants during the compatibility stage by
-- setting their seat limit to at least their current user count.
UPDATE platform_subscriptions s
SET user_limit = GREATEST(
    1,
    COALESCE((
        SELECT COUNT(*)
        FROM users u
        WHERE u.tenant_id = s.tenant_id
    ), 0)
)
WHERE s.status IN ('TRIAL', 'ACTIVE');
