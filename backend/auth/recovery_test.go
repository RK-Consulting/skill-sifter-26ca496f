package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateProvisioningRecoveryToken(t *testing.T) {
	oldKey := JwtKey
	JwtKey = []byte("test-recovery-key")
	defer func() { JwtKey = oldKey }()

	tokenString, err := GenerateProvisioningRecoveryToken(
		"tenant_test_recovery",
		"admin@example.com",
		"Recovery Company",
	)
	if err != nil {
		t.Fatalf("GenerateProvisioningRecoveryToken failed: %v", err)
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return JwtKey, nil
	})
	if err != nil || !token.Valid {
		t.Fatalf("recovery token is invalid: %v", err)
	}
	if claims.UserID != 0 {
		t.Fatalf("recovery token user ID = %d, want 0", claims.UserID)
	}
	if claims.TenantID != "tenant_test_recovery" {
		t.Fatalf("recovery token tenant ID = %q", claims.TenantID)
	}
	if claims.Email != "admin@example.com" {
		t.Fatalf("recovery token email = %q", claims.Email)
	}
	if claims.Role != "admin" {
		t.Fatalf("recovery token role = %q, want admin", claims.Role)
	}
	if !claims.Recovery {
		t.Fatal("recovery token does not carry recovery marker")
	}
	if claims.ExpiresAt == nil || claims.ExpiresAt.Time.IsZero() {
		t.Fatal("recovery token has no expiry")
	}
}
