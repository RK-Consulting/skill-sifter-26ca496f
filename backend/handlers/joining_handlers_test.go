package handlers

import (
	"encoding/json"
	"testing"
)

func TestParseJoiningDate(t *testing.T) {
	got, err := parseJoiningDate("2026-10-16")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected parsed date")
	}
	if got.Format("2006-01-02") != "2026-10-16" {
		t.Fatalf("got %s, want 2026-10-16", got.Format("2006-01-02"))
	}
}

func TestParseJoiningDateEmpty(t *testing.T) {
	got, err := parseJoiningDate("")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil date")
	}
}

func TestParseJoiningDateRejectsInvalidFormat(t *testing.T) {
	for _, value := range []string{"16-10-2026", "2026/10/16", "2026-02-30", "not-a-date"} {
		if _, err := parseJoiningDate(value); err == nil {
			t.Fatalf("expected error for %q", value)
		}
	}
}

func TestJoiningRequestDecodesDateOnlyValue(t *testing.T) {
	var req joiningRequest
	if err := json.Unmarshal([]byte(`{"joiningDate":"2026-10-16","joined":true}`), &req); err != nil {
		t.Fatal(err)
	}
	got, err := parseJoiningDate(req.JoiningDate)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Format("2006-01-02") != "2026-10-16" {
		t.Fatalf("unexpected parsed date: %v", got)
	}
	if !req.Joined {
		t.Fatal("expected joined=true")
	}
}
