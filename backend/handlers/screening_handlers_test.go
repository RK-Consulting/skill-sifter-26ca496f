package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeScreeningRequest_ParsesRecruiterEvidence(t *testing.T) {
	req := httptest.NewRequest(
		"POST",
		"/api/v1/assignments/10/screenings",
		strings.NewReader(`{"currentCTC":"12 LPA","expectedCTC":"16 LPA","noticePeriod":"30 days","lastWorkingDay":"2026-10-15","currentLocation":"Bengaluru","willingToRelocate":true,"preferredLocation":"Pune","candidateInterest":"interested","availabilityDate":"2026-10-16"}`),
	)
	req = req.WithContext(context.WithValue(req.Context(), "userID", 42))

	input, err := decodeScreeningRequest(req, 10)
	if err != nil {
		t.Fatalf("decodeScreeningRequest returned error: %v", err)
	}
	if input.AssignmentID != 10 || input.RecruiterUserID != 42 {
		t.Fatalf("assignment/recruiter = %d/%d, want 10/42", input.AssignmentID, input.RecruiterUserID)
	}
	if input.CurrentCTC != "12 LPA" || input.ExpectedCTC != "16 LPA" {
		t.Fatalf("CTC values = %q/%q", input.CurrentCTC, input.ExpectedCTC)
	}
	if input.LastWorkingDay == nil || input.LastWorkingDay.Format("2006-01-02") != "2026-10-15" {
		t.Fatalf("unexpected last working day: %v", input.LastWorkingDay)
	}
	if input.WillingToRelocate == nil || !*input.WillingToRelocate {
		t.Fatal("willingToRelocate should be true")
	}
	if input.AvailabilityDate == nil || input.AvailabilityDate.Format("2006-01-02") != "2026-10-16" {
		t.Fatalf("unexpected availability date: %v", input.AvailabilityDate)
	}
}

func TestDecodeScreeningRequest_InvalidDate(t *testing.T) {
	req := httptest.NewRequest(
		"POST",
		"/api/v1/assignments/10/screenings",
		strings.NewReader(`{"lastWorkingDay":"15-10-2026"}`),
	)
	req = req.WithContext(context.WithValue(req.Context(), "userID", 42))

	_, err := decodeScreeningRequest(req, 10)
	if err == nil {
		t.Fatal("expected invalid date error")
	}
}

func TestParseScreeningDate_EmptyIsNil(t *testing.T) {
	got, err := parseScreeningDate("")
	if err != nil || got != nil {
		t.Fatalf("got %v, err %v; want nil, nil", got, err)
	}
}
