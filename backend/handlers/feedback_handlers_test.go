package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeFeedbackRequest(t *testing.T) {
	req := httptest.NewRequest(
		"POST",
		"/api/v1/submissions/10/feedback",
		strings.NewReader(`{"outcome":"shortlist","reasonCode":"profile_mismatch","comments":"Proceed to technical round","nextAction":"Schedule interview"}`),
	)
	req = req.WithContext(context.WithValue(req.Context(), "userID", 42))

	input, err := decodeFeedbackRequest(req, 10, 42)
	if err != nil {
		t.Fatalf("decodeFeedbackRequest returned error: %v", err)
	}
	if input.SubmissionID != 10 || input.FeedbackByUserID != 42 {
		t.Fatalf("submission/actor = %d/%d, want 10/42", input.SubmissionID, input.FeedbackByUserID)
	}
	if string(input.Outcome) != "shortlist" {
		t.Fatalf("outcome = %q, want shortlist", input.Outcome)
	}
	if input.ReasonCode != "profile_mismatch" {
		t.Fatalf("reasonCode = %q, want profile_mismatch", input.ReasonCode)
	}
	if input.NextAction != "Schedule interview" {
		t.Fatalf("nextAction = %q, want Schedule interview", input.NextAction)
	}
}

func TestDecodeFeedbackRequest_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(
		"POST",
		"/api/v1/submissions/10/feedback",
		strings.NewReader(`{"outcome":`),
	)

	_, err := decodeFeedbackRequest(req, 10, 42)
	if err == nil {
		t.Fatal("expected invalid request payload error")
	}
}
