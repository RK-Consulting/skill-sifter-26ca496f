package handlers

import (
	"testing"
	"time"
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

func TestJoiningRequestDateContract(t *testing.T) {
	date := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	if date.Format("2006-01-02") != "2026-10-16" {
		t.Fatal("unexpected date contract")
	}
}
