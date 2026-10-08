package auth

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/RK-Consulting/skill-sifter/db"
	"github.com/RK-Consulting/skill-sifter/domain/platformaccess"
	"github.com/RK-Consulting/skill-sifter/models"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key, read from the JWT_SECRET environment variable.
// Falls back to a well-known dev-only value with a loud warning so it
// is never silently used in a real deployment.
var JwtKey = loadJwtKey()

func loadJwtKey() []byte {
	secret := db.GetEnv("JWT_SECRET", "")
	if secret == "" {
		if db.GetEnv("SKILLSIFTER_ENV", "") == "production" {
			log.Fatal("JWT_SECRET must be configured in production")
		}
		log.Println("⚠️  WARNING: JWT_SECRET is not set. Using an insecure default key for non-production use.")
		secret = "dev_only_insecure_default_key"
	}
	return []byte(secret)
}

// Claims for JWT.
//
// TenantID is the authoritative tenant identity per ADR 0001 (Stage C):
// tenant-scoped operations must be scoped to TenantID, not CompanyName.
// CompanyName is retained on the claims for display/compatibility with
// existing clients during the migration window (ADR 0001 explicitly lists
// "existing JWTs containing companyName" as a compatibility concern), but it
// must never be used as an authorization or isolation key going forward.
type Claims struct {
	UserID      int    `json:"userId"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	TenantID    string `json:"tenantId"`
	CompanyName string `json:"companyName"`
	Recovery    bool   `json:"recovery,omitempty"`
	jwt.RegisteredClaims
}

// TenantDBMiddleware resolves the authenticated tenant to its READY database.
// Control-plane authentication remains on db.DB; tenant-owned handlers use
// db.RequestDB(r) to access this request-scoped connection pool.
func TenantDBMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/account") || r.URL.Path == "/api/admin/tenant/provision" {
			next.ServeHTTP(w, r)
			return
		}

		tenantID, ok := r.Context().Value("tenantID").(string)
		if !ok || tenantID == "" {
			http.Error(w, "Tenant context missing", http.StatusUnauthorized)
			return
		}

		tenantDB, err := db.TenantDB(db.DB, tenantID)
		if err != nil {
			http.Error(w, "Tenant database is not ready", http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r.WithContext(db.WithTenantDB(r.Context(), tenantDB)))
	})
}

// AuthMiddleware authenticates JWT tokens
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":false,"message":"Authorization header is required. Please include a valid Bearer token in your request headers."}`))
			return
		}

		// Check if the header has the format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Authorization header format must be 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return JwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Reject tokens issued before tenant_id existed on the claims,
		// rather than silently propagating an empty tenant identity into
		// request context, where an empty string could otherwise be
		// mistaken for "no tenant restriction" by a future bug. Existing
		// users are unaffected in practice: JWTs expire after 24h
		// (auth.GenerateToken), and every token issued after this change
		// carries TenantID, so this only ever rejects tokens from before
		// deployment, which would have expired anyway within one day.
		if claims.TenantID == "" {
			http.Error(w, "Token is missing tenant identity; please log in again", http.StatusUnauthorized)
			return
		}

		// Re-resolve the control-plane access context on every protected
		// request. This prevents a suspended/expired subscription from
		// remaining usable until JWT expiry and makes the platform layer
		// authoritative for tenant + RBAC.
		var access platformaccess.LoginAccess
		if r.URL.Path == "/api/admin/tenant/provision" && claims.Recovery {
			if claims.UserID != 0 || claims.Role != "admin" {
				http.Error(w, "Invalid provisioning recovery token", http.StatusForbidden)
				return
			}
			var accountStatus, provisioningStatus, companyName string
			if err := db.DB.QueryRow(
				`SELECT account_status, provisioning_status, company_name
				 FROM platform_tenants WHERE tenant_id=$1`,
				claims.TenantID,
			).Scan(&accountStatus, &provisioningStatus, &companyName); err != nil ||
				accountStatus == "TERMINATED" ||
				(provisioningStatus != "PENDING" && provisioningStatus != "PROVISIONING" && provisioningStatus != "FAILED" && provisioningStatus != "READY") {
				http.Error(w, "Tenant provisioning recovery is not available", http.StatusForbidden)
				return
			}
			access = platformaccess.LoginAccess{
				TenantID: claims.TenantID, Role: "admin",
				AccountStatus: accountStatus, ProvisioningStatus: provisioningStatus,
			}
		} else {
			if r.URL.Path == "/api/admin/tenant/provision" {
				access, err = platformaccess.ResolveProvisioningAccess(db.DB, claims.UserID, claims.TenantID)
			} else {
				access, err = platformaccess.ResolveLoginAccess(db.DB, claims.UserID, claims.TenantID)
			}
			if err != nil {
				http.Error(w, "Tenant subscription or access is not active", http.StatusForbidden)
				return
			}
		}

		if access.SubscriptionStatus == "EXPIRED" {
			path := r.URL.Path
			if !strings.HasPrefix(path, "/api/account") && path != "/api/admin/tenant/provision" {
				http.Error(w, "Trial expired. Please subscribe to continue.", http.StatusPaymentRequired)
				return
			}
		}

		// Store trusted control-plane context for downstream handlers.
		ctx := r.Context()
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "provisioningRecovery", claims.Recovery)
		ctx = context.WithValue(ctx, "role", access.Role)
		ctx = context.WithValue(ctx, "tenantID", access.TenantID)
		ctx = context.WithValue(ctx, "companyName", claims.CompanyName)
		ctx = context.WithValue(ctx, "subscriptionStatus", access.SubscriptionStatus)
		ctx = context.WithValue(ctx, "planCode", access.PlanCode)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware restricts access based on user role
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the role resolved from the trusted control-plane context.
			role, ok := r.Context().Value("role").(string)
			if !ok || role == "" {
				http.Error(w, "Missing authorization context", http.StatusForbidden)
				return
			}

			// Check if the role is allowed
			allowed := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}


// GenerateProvisioningRecoveryToken creates a short-lived token that can only
// be used to retry provisioning for a verified registration.
func GenerateProvisioningRecoveryToken(tenantID, email, companyName string) (string, error) {
	claims := &Claims{
		UserID: 0, Email: email, Role: "admin", TenantID: tenantID,
		CompanyName: companyName, Recovery: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtKey)
}

// GenerateToken creates a JWT token for a user. user.TenantID must be the
// authoritative tenant_id for the user's tenant (ADR 0001) — callers must
// never pass a client-supplied value here.
func GenerateToken(user models.User, roleName string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		TenantID:    user.TenantID,
		CompanyName: user.CompanyName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(JwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
