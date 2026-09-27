package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubmissionRequestContract(t *testing.T) {
	req := httptest.NewRequest(
		"POST",
		"/api/v1/assignments/10/submissions",
		strings.NewReader(`{"recipientType":"client","recipientClientId":7,"recipientName":"Acme","recipientEmail":"hm@acme.example","submissionContext":"Backend opening","recruiterNotes":"Screened and interested"}`),
	)
	var payload submissionRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		t.Fatalf("decode returned error: %v", err)
	}
	if payload.RecipientType != "client" || payload.RecipientClientID == nil || *payload.RecipientClientID != 7 {
		t.Fatalf("unexpected recipient: %#v", payload)
	}
	if payload.SubmissionContext != "Backend opening" {
		t.Fatalf("unexpected context: %q", payload.SubmissionContext)
	}
}

func TestSubmissionRequestRejectsMissingPayload(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/assignments/10/submissions", strings.NewReader("{"))
	var payload submissionRequest
	if err := json.NewDecoder(req.Body).Decode(&payload); err == nil {
		t.Fatal("expected malformed JSON error")
	}
}
