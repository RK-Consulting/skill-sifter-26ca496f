package handlers

import (
	"os"
	"testing"
)

func TestOTPHashRequiresSecret(t *testing.T) {
	t.Setenv("OTP_HASH_SECRET", "")
	if _, err := otpHash("123456"); err == nil {
		t.Fatal("expected missing OTP_HASH_SECRET to return an error")
	}
}

func TestOTPHashUsesConfiguredSecret(t *testing.T) {
	t.Setenv("OTP_HASH_SECRET", "test-secret")
	got, err := otpHash("123456")
	if err != nil {
		t.Fatalf("otpHash returned unexpected error: %v", err)
	}
	if got == "" {
		t.Fatal("expected OTP hash")
	}

	if err := os.Unsetenv("OTP_HASH_SECRET"); err != nil {
		t.Fatalf("failed to clear OTP_HASH_SECRET: %v", err)
	}
	if _, err := otpHash("123456"); err == nil {
		t.Fatal("expected otpHash to fail after secret is removed")
	}
}
