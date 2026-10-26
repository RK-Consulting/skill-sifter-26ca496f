package db

import (
	"database/sql"
	"testing"
	"time"
)

func TestRetirementStatus(t *testing.T) {
	now := time.Date(2026, time.October, 26, 9, 0, 0, 0, time.UTC)
	expired := sql.NullTime{Time: now.Add(-24 * time.Hour), Valid: true}
	future := sql.NullTime{Time: now.Add(24 * time.Hour), Valid: true}

	tests := []struct {
		name, accountStatus, subscriptionStatus string
		trialExpiresAt                          sql.NullTime
		want                                    string
	}{
		{name: "cancelled subscription is discontinued", accountStatus: "ACTIVE", subscriptionStatus: "CANCELLED", trialExpiresAt: future, want: "DISCONTINUED"},
		{name: "expired subscription is trial expired", accountStatus: "ACTIVE", subscriptionStatus: "EXPIRED", trialExpiresAt: future, want: "TRIAL_EXPIRED"},
		{name: "past trial expiry is trial expired", accountStatus: "ACTIVE", subscriptionStatus: "TRIAL", trialExpiresAt: expired, want: "TRIAL_EXPIRED"},
		{name: "expired account is trial expired", accountStatus: "EXPIRED", subscriptionStatus: "", want: "TRIAL_EXPIRED"},
		{name: "terminated account is deleted", accountStatus: "TERMINATED", subscriptionStatus: "", trialExpiresAt: future, want: "ACCOUNT_DELETED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retirementStatus(tt.accountStatus, tt.subscriptionStatus, tt.trialExpiresAt, now)
			if got != tt.want {
				t.Fatalf("retirementStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
