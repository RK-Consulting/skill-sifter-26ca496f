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
		name                   string
		accountStatus          string
		subscriptionStatus     string
		subscriptionProvider   string
		trialExpiresAt         sql.NullTime
		want                   string
	}{
		{name: "cancelled subscription is discontinued", accountStatus: "ACTIVE", subscriptionStatus: "CANCELLED", subscriptionProvider: "razorpay", trialExpiresAt: future, want: "DISCONTINUED"},
		{name: "expired trial subscription is trial expired", accountStatus: "ACTIVE", subscriptionStatus: "EXPIRED", subscriptionProvider: "", trialExpiresAt: future, want: "TRIAL_EXPIRED"},
		{name: "expired paid subscription is not mislabeled as trial", accountStatus: "EXPIRED", subscriptionStatus: "EXPIRED", subscriptionProvider: "razorpay", trialExpiresAt: expired, want: "ACCOUNT_DELETED"},
		{name: "past trial expiry is trial expired", accountStatus: "ACTIVE", subscriptionStatus: "TRIAL", subscriptionProvider: "", trialExpiresAt: expired, want: "TRIAL_EXPIRED"},
		{name: "terminated trial is trial expired", accountStatus: "EXPIRED", subscriptionStatus: "EXPIRED", subscriptionProvider: "", trialExpiresAt: expired, want: "TRIAL_EXPIRED"},
		{name: "terminated account is deleted even if trial date is old", accountStatus: "TERMINATED", subscriptionStatus: "", subscriptionProvider: "", trialExpiresAt: expired, want: "ACCOUNT_DELETED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retirementStatus(tt.accountStatus, tt.subscriptionStatus, tt.subscriptionProvider, tt.trialExpiresAt, now)
			if got != tt.want {
				t.Fatalf("retirementStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}
