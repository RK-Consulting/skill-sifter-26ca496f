package handlers

import (
	"context"\n\t"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSafeResumeName(t *testing.T) {
	cases := map[string]string{
		"../candidate.pdf": "candidate.pdf",
		"John Doe CV.pdf": "John_Doe_CV.pdf",
		"resume;rm -rf.txt": "resume_rm_-rf.txt",
		"": "resume.bin",
	}
	for input, want := range cases {
		if got := safeResumeName(input); got != want {
			t.Errorf("safeResumeName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExtractResumeText(t *testing.T) {
	if got := extractResumeText([]byte("Go\nPostgreSQL\nReact"), "resume.txt"); got != "Go\nPostgreSQL\nReact" {
		t.Fatalf("text extraction mismatch: %q", got)
	}
	if got := extractResumeText([]byte("plain text"), "resume.exe"); got != "" {
		t.Fatalf("unsupported format returned %q", got)
	}
	if got := extractResumeText([]byte("not a zip"), "resume.docx"); got != "" {
		t.Fatal("malformed DOCX should return empty text")
	}
	pdf := []byte("BT (Harish Nagaraju) Tj (Go PostgreSQL) Tj ET")
	if got := extractResumeText(pdf, "resume.pdf"); !strings.Contains(got, "Harish Nagaraju") || !strings.Contains(got, "Go PostgreSQL") {
		t.Fatalf("PDF extraction did not recover expected text: %q", got)
	}
}

func TestCallOllamaSuccessAndMalformedJSON(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/api/generate" {
			t.Errorf("path = %s, want /api/generate", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if requests == 1 {
			io.WriteString(w, "{\"response\":\"{\\\"name\\\":\\\"Ada\\\",\\\"email\\\":\\\"ada@example.com\\\",\\\"phone\\\":\\\"123\\\",\\\"skills\\\":[\\\"Go\\\",\\\"PostgreSQL\\\"]}\"}")
			return
		}
		io.WriteString(w, "{\"response\":\"not-json\"}")
	}))
	defer server.Close()
	t.Setenv("OLLAMA_URL", server.URL)

	got, errText := callOllama("Ada resume")
	if errText != "" || got.Name != "Ada" || len(got.Skills) != 2 {
		t.Fatalf("successful Ollama parse = %+v, err=%q", got, errText)
	}
	_, errText = callOllama("Ada resume")
	if !strings.Contains(errText, "Ollama JSON parse failed") {
		t.Fatalf("malformed Ollama response error = %q", errText)
	}
}

func TestCallOllamaHTTPErrorAndUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()
	t.Setenv("OLLAMA_URL", server.URL)
	_, errText := callOllama("resume")
	if !strings.Contains(errText, "Ollama returned 503") {
		t.Fatalf("HTTP error = %q", errText)
	}

	t.Setenv("OLLAMA_URL", "http://127.0.0.1:1")
	_, errText = callOllama("resume")
	if !strings.Contains(errText, "Ollama unavailable") {
		t.Fatalf("unavailable error = %q", errText)
	}
}

func TestUploadResumesRequiresTenant(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/resume-ai/upload", nil)
	rec := httptest.NewRecorder()
	UploadResumes(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestSearchResumesRequiresTenant(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/resume-ai/search?q=Go", nil)
	rec := httptest.NewRecorder()
	SearchResumes(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
}

func TestSearchResumesRequiresQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/resume-ai/search", nil)
	req = req.WithContext(contextWithResumeTenant("tenant_a"))
	rec := httptest.NewRecorder()
	SearchResumes(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("missing q status = %d, want 400", rec.Code)
	}
}

func contextWithResumeTenant(tenant string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "tenantID", tenant)
	return ctx
}
