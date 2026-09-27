package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestResumeRepositoryHandlersRequireTenant(t *testing.T) {
	tests := []struct {
		name string
		handler http.HandlerFunc
		method string
		path string
	}{
		{"detail", GetResumeDetail, http.MethodGet, "/api/resume-ai/resumes/1"},
		{"download", DownloadResume, http.MethodGet, "/api/resume-ai/resumes/1/file"},
		{"retry", RetryResume, http.MethodPost, "/api/resume-ai/resumes/1/retry"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			tc.handler(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
		})
	}
}

func TestResumeIDValidation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/resume-ai/resumes/not-an-id", nil)
	req = req.WithContext(context.WithValue(context.Background(), "tenantID", "tenant_a"))
	req = req.WithContext(context.WithValue(req.Context(), "tenantID", "tenant_a"))
	if id, ok := resumeID(req); ok || id != 0 {
		t.Fatalf("resumeID returned id=%d ok=%v for invalid path without mux vars", id, ok)
	}
}

func TestStoredResumeFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + string(os.PathSeparator) + "resume.txt"
	if err := os.WriteFile(path, []byte("resume"), 0600); err != nil {
		t.Fatal(err)
	}
	if !storedResumeFile(path) {
		t.Fatal("storedResumeFile returned false for a regular file")
	}
	if storedResumeFile(dir) {
		t.Fatal("storedResumeFile returned true for a directory")
	}
	if storedResumeFile("") {
		t.Fatal("storedResumeFile returned true for an empty path")
	}
}
