package handlers

import (
	"bytes"
	"net/http/httptest"
	"testing"

)

func TestGetPipelineReportIsTenantScoped(t *testing.T) {
	testDB := setupIsolationTestDB(t)
	if _, err := testDB.Exec(`DELETE FROM candidates WHERE tenant_id IN ('tenant_a', 'tenant_b')`); err != nil {
		t.Fatalf("candidate cleanup failed: %v", err)
	}

	if _, err := testDB.Exec(`
		INSERT INTO candidates (name, email, status, pipeline_stage, tenant_id, company_name)
		VALUES
			('A Screening', 'a-screening@test.com', 'active', 'screening', 'tenant_a', 'Tenant A Co'),
			('A Interview', 'a-interview@test.com', 'active', 'interview', 'tenant_a', 'Tenant A Co'),
			('B Rejected', 'b-rejected@test.com', 'active', 'rejected', 'tenant_b', 'Tenant B Co')
	`); err != nil {
		t.Fatalf("candidate seed failed: %v", err)
	}

	req := isoCtx(httptest.NewRequest("GET", "/api/reports/pipeline", nil), "tenant_a")
	rec := httptest.NewRecorder()
	GetPipelineReport(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200; body = %s", rec.Code, rec.Body.String())
	}

	body := rec.Body.Bytes()
	if !bytes.Contains(body, []byte(`"stage":"screening","count":1`)) {
		t.Errorf("screening count missing or incorrect: %s", body)
	}
	if !bytes.Contains(body, []byte(`"stage":"interview","count":1`)) {
		t.Errorf("interview count missing or incorrect: %s", body)
	}
	if !bytes.Contains(body, []byte(`"stage":"rejected","count":0`)) {
		t.Errorf("cross-tenant rejected count leaked or is incorrect: %s", body)
	}
}
