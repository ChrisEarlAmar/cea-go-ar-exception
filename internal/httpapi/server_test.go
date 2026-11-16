package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/chrisamar/cea-go-ar-exception/internal/report"
)

type fakeRunner struct {
	filter report.ChargedSalesFilter
	result report.Result
	err    error
}

func (r *fakeRunner) ChargedSales(_ context.Context, filter report.ChargedSalesFilter) (report.Result, error) {
	r.filter = filter
	return r.result, r.err
}

func TestChargedSalesReturnsReport(t *testing.T) {
	runner := &fakeRunner{result: report.Result{
		Columns: []string{"ID", "COMPANY"},
		Rows:    []map[string]any{{"ID": int64(1), "COMPANY": "Personal Demo"}},
		Count:   1,
	}}

	request := httptest.NewRequest(http.MethodPost, "/api/reports/charged-sales", strings.NewReader(`{"dateFrom":"2026-01-01","dateTo":"2026-01-31","companyID":10,"storeCode":"001"}`))
	response := httptest.NewRecorder()
	New(runner).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if runner.filter.CompanyID == nil || *runner.filter.CompanyID != 10 {
		t.Fatalf("company filter = %#v, want 10", runner.filter.CompanyID)
	}
	if runner.filter.StoreCode == nil || *runner.filter.StoreCode != "001" {
		t.Fatalf("store filter = %#v, want 001", runner.filter.StoreCode)
	}
	if !runner.filter.DateFrom.Equal(time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("date from = %v", runner.filter.DateFrom)
	}

	var body report.Result
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Count != 1 || len(body.Rows) != 1 {
		t.Fatalf("response = %#v", body)
	}
}

func TestChargedSalesRejectsBadDateRange(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/api/reports/charged-sales", strings.NewReader(`{"dateFrom":"2026-02-01","dateTo":"2026-01-01"}`))
	response := httptest.NewRecorder()
	New(&fakeRunner{}).ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestNewWithFrontendServesBuiltApp(t *testing.T) {
	frontendDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(frontendDir, "index.html"), []byte("<h1>AR Exception</h1>"), 0o600); err != nil {
		t.Fatalf("write frontend: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	NewWithFrontend(&fakeRunner{}, frontendDir).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.Contains(response.Body.String(), "AR Exception") {
		t.Fatalf("frontend response = %q", response.Body.String())
	}
}
