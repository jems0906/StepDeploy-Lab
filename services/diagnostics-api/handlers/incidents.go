package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal/health"
	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal/incidents"
)

func IncidentsHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Incidents request from %s", r.RemoteAddr)

	tracker := incidents.GetTracker()
	list := tracker.ListIncidents()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"incidents": list,
		"count":     len(list),
	})
}

func IncidentDetailHandler(w http.ResponseWriter, r *http.Request) {
	incidentID := r.URL.Query().Get("incident_id")
	if incidentID == "" {
		http.Error(w, "incident_id query parameter required", http.StatusBadRequest)
		return
	}

	logger.Debug("Incident detail request for: %s", incidentID)

	tracker := incidents.GetTracker()
	incident, err := tracker.GetIncident(incidentID)
	if err != nil {
		logger.Warn("Incident not found: %s", incidentID)
		http.Error(w, fmt.Sprintf("Incident not found: %s", incidentID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

func CreateIncidentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	tracker := incidents.GetTracker()
	incident := tracker.CreateIncident(
		req["trace_id"],
		req["severity"],
		req["title"],
		req["description"],
		req["root_cause"],
		req["runbook"],
	)

	logger.Info("Incident created: %s", incident.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(incident)
}

func ResolveIncidentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	incidentID := r.URL.Query().Get("incident_id")
	if incidentID == "" {
		http.Error(w, "incident_id query parameter required", http.StatusBadRequest)
		return
	}

	tracker := incidents.GetTracker()
	err := tracker.ResolveIncident(incidentID)
	if err != nil {
		logger.Warn("Failed to resolve incident: %v", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	logger.Info("Incident resolved: %s", incidentID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":      "resolved",
		"incident_id": incidentID,
	})
}

func EnrollHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Enrollment test request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp, err := http.Post(health.EnrollmentAgentURL()+"/enroll", "application/json", nil)
	if err != nil {
		logger.Error("Failed to call enrollment agent: %v", err)
		http.Error(w, fmt.Sprintf("Enrollment failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var enrollmentResult map[string]interface{}
	json.Unmarshal(body, &enrollmentResult)

	// Create incident if enrollment failed
	if status, ok := enrollmentResult["status"].(string); ok && status != "success" {
		if traceID, ok := enrollmentResult["trace_id"].(string); ok {
			tracker := incidents.GetTracker()
			incident := tracker.CreateIncident(
				traceID,
				"critical",
				"Enrollment Failed",
				fmt.Sprintf("Device enrollment failed: %v", enrollmentResult),
				"Enrollment workflow error",
				"docs/escalation_playbook.md",
			)
			logger.Info("Incident created for failed enrollment: %s", incident.ID)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(enrollmentResult)
}
