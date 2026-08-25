package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal/incidents"
)

func TestIncidentLifecycleViaHandlers(t *testing.T) {
	createBody, _ := json.Marshal(map[string]string{
		"trace_id":    "trace-1",
		"severity":    "critical",
		"title":       "Test Incident",
		"description": "something broke",
		"root_cause":  "unknown",
		"runbook":     "docs/escalation_playbook.md",
	})
	req := httptest.NewRequest(http.MethodPost, "/incidents", bytes.NewReader(createBody))
	rec := httptest.NewRecorder()
	CreateIncidentHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var created incidents.Incident
	if err := json.NewDecoder(rec.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	if created.ID == "" {
		t.Fatal("expected incident ID to be set")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/incidents", nil)
	listRec := httptest.NewRecorder()
	IncidentsHandler(listRec, listReq)

	var listResp struct {
		Count int `json:"count"`
	}
	json.NewDecoder(listRec.Body).Decode(&listResp)
	if listResp.Count < 1 {
		t.Fatalf("expected at least one incident, got %d", listResp.Count)
	}

	resolveReq := httptest.NewRequest(http.MethodPost, "/incidents/resolve?incident_id="+created.ID, nil)
	resolveRec := httptest.NewRecorder()
	ResolveIncidentHandler(resolveRec, resolveReq)
	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected 200 resolving incident, got %d: %s", resolveRec.Code, resolveRec.Body.String())
	}
}
