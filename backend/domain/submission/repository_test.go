package submission

import (
	"database/sql"
	"testing"
	"time"
)

type submissionTestScanner struct {
	values []interface{}
}

func (s *submissionTestScanner) Scan(dest ...interface{}) error {
	for i, value := range s.values {
		switch d := dest[i].(type) {
		case *int:
			*d = value.(int)
		case *string:
			*d = value.(string)
		case *sql.NullInt64:
			*d = value.(sql.NullInt64)
		case *sql.NullString:
			*d = value.(sql.NullString)
		case *[]byte:
			*d = value.([]byte)
		case *time.Time:
			*d = value.(time.Time)
		}
	}
	return nil
}

func TestScanSubmissionHandlesNullableFields(t *testing.T) {
	now := time.Now()
	scanner := &submissionTestScanner{values: []interface{}{
		42,
		"e2e_smoke_tenant",
		10,
		20,
		1,
		"client",
		sql.NullInt64{Int64: 7, Valid: true},
		sql.NullString{String: "Smoke Client", Valid: true},
		sql.NullString{},
		sql.NullString{},
		sql.NullString{},
		[]byte(`{"id":10}`),
		[]byte(`{"id":20}`),
		now,
		now,
	}}

	record, err := scanSubmission(scanner)
	if err != nil {
		t.Fatalf("scanSubmission returned error: %v", err)
	}

	if record.ID != 42 || record.TenantID != "e2e_smoke_tenant" {
		t.Fatalf("unexpected identity fields: %+v", record)
	}
	if record.RecipientClientID == nil || *record.RecipientClientID != 7 {
		t.Fatalf("expected recipient client ID 7, got %v", record.RecipientClientID)
	}
	if record.RecipientName != "Smoke Client" {
		t.Fatalf("expected recipient name, got %q", record.RecipientName)
	}
	if record.RecipientEmail != "" || record.SubmissionContext != "" || record.RecruiterNotes != "" {
		t.Fatalf("expected NULL optional fields to map to empty strings: %+v", record)
	}
	if string(record.CandidateSnapshot) != `{"id":10}` || string(record.RequirementSnapshot) != `{"id":20}` {
		t.Fatalf("unexpected snapshots: %s / %s", record.CandidateSnapshot, record.RequirementSnapshot)
	}
}
