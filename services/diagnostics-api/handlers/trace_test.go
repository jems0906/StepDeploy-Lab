package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal"
)

func TestTraceReportAndList(t *testing.T) {
	trace := &internal.Trace{TraceID: "trace-report-test", Status: "success"}
	body, _ := json.Marshal(trace)

	reportReq := httptest.NewRequest(http.MethodPost, "/traces/report", bytes.NewReader(body))
	reportRec := httptest.NewRecorder()
	TraceReportHandler(reportRec, reportReq)

	if reportRec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", reportRec.Code, reportRec.Body.String())
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/traces/detail?trace_id=trace-report-test", nil)
	detailRec := httptest.NewRecorder()
	TraceDetailHandler(detailRec, detailReq)

	if detailRec.Code != http.StatusOK {
		t.Fatalf("expected 200 fetching trace detail, got %d", detailRec.Code)
	}
	var got internal.Trace
	if err := json.NewDecoder(detailRec.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "success" {
		t.Fatalf("expected status success, got %q", got.Status)
	}
}

func TestTraceReportRejectsMissingTraceID(t *testing.T) {
	body, _ := json.Marshal(&internal.Trace{})
	req := httptest.NewRequest(http.MethodPost, "/traces/report", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	TraceReportHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing trace_id, got %d", rec.Code)
	}
}
