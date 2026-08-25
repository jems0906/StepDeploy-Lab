package incidents

import (
	"testing"
	"time"
)

func TestIncidentSLAAndResolution(t *testing.T) {
	tracker := &IncidentTracker{incidents: make(map[string]*Incident)}
	incident := tracker.CreateIncident("trace-1", "critical", "Enrollment failed", "failure", "root cause", "runbook")
	if incident.SLATarget != 5*time.Minute {
		t.Fatalf("expected critical SLA to be 5 minutes, got %v", incident.SLATarget)
	}
	if err := tracker.ResolveIncident(incident.ID); err != nil {
		t.Fatal(err)
	}
	if incident.ResolvedAt == nil || tracker.IsSLABreached(incident) {
		t.Fatal("expected promptly resolved incident to be within SLA")
	}
}
