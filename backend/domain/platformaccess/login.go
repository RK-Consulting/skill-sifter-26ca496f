package platformaccess

import (
	"database/sql"
	"fmt"
)

type LoginAccess struct {
	TenantID           string
	ProvisioningStatus string
	AccountStatus      string
	SubscriptionStatus string
	PlanCode           string
	Role               string
}

func resolveAccess(dbConn *sql.DB, userID int, tenantID string, requireReady bool) (LoginAccess, error) {
	var access LoginAccess
	err := dbConn.QueryRow(`
		SELECT
			pua.tenant_id,
			pt.provisioning_status,
			pt.account_status,
			COALESCE(active_sub.status, ''),
			COALESCE(active_sub.plan_code, ''),
			pua.role
		FROM platform_user_accounts pua
		JOIN platform_tenants pt ON pt.tenant_id = pua.tenant_id
		LEFT JOIN LATERAL (
			SELECT status, plan_code
			FROM platform_subscriptions
			WHERE tenant_id = pua.tenant_id
			  AND status IN ('TRIAL', 'ACTIVE', 'EXPIRED', 'PAST_DUE')
			ORDER BY starts_at DESC, id DESC
			LIMIT 1
		) active_sub ON TRUE
		WHERE pua.user_id = $1
		  AND pua.tenant_id = $2
	`, userID, tenantID).Scan(
		&access.TenantID,
		&access.ProvisioningStatus,
		&access.AccountStatus,
		&access.SubscriptionStatus,
		&access.PlanCode,
		&access.Role,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return LoginAccess{}, fmt.Errorf("tenant access context not found")
		}
		return LoginAccess{}, fmt.Errorf("resolve tenant access: %w", err)
	}

	if requireReady && access.ProvisioningStatus != "READY" {
		return LoginAccess{}, fmt.Errorf("tenant database is not ready")
	}
	if access.AccountStatus != "ACTIVE" {
		return LoginAccess{}, fmt.Errorf("tenant account is %s", access.AccountStatus)
	}
	if access.SubscriptionStatus == "" {
		return LoginAccess{}, fmt.Errorf("tenant subscription is not active")
	}
	return access, nil
}

func ResolveLoginAccess(dbConn *sql.DB, userID int, tenantID string) (LoginAccess, error) {
	return resolveAccess(dbConn, userID, tenantID, true)
}

// ResolveProvisioningAccess authenticates an active tenant even when its
// tenant database is PENDING or FAILED, so the administrator can recover
// provisioning through the dedicated control-plane endpoint.
func ResolveProvisioningAccess(dbConn *sql.DB, userID int, tenantID string) (LoginAccess, error) {
	return resolveAccess(dbConn, userID, tenantID, false)
}
