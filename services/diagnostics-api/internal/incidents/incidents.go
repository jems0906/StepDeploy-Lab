package incidents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Incident struct {
	ID          string        `json:"id"`
	TraceID     string        `json:"trace_id"`
	Severity    string        `json:"severity"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	RootCause   string        `json:"root_cause"`
	CreatedAt   time.Time     `json:"created_at"`
	ResolvedAt  *time.Time    `json:"resolved_at,omitempty"`
	SLATarget   time.Duration `json:"sla_target_minutes"`
	Runbook     string        `json:"runbook"`
}

type IncidentTracker struct {
	incidents map[string]*Incident
	mu        sync.Mutex
}

var tracker = &IncidentTracker{
	incidents: make(map[string]*Incident),
}

// storePath enables file-based persistence when set via Init. Empty by
// default so existing in-memory-only callers (and tests) are unaffected.
var storePath string

// persistedIncident is the on-disk shape. It stores SLATarget as nanoseconds
// explicitly, independent of Incident's API-facing MarshalJSON.
type persistedIncident struct {
	ID          string     `json:"id"`
	TraceID     string     `json:"trace_id"`
	Severity    string     `json:"severity"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	RootCause   string     `json:"root_cause"`
	CreatedAt   time.Time  `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	SLATargetNs int64      `json:"sla_target_ns"`
	Runbook     string     `json:"runbook"`
}

// Init enables file-based persistence at path and loads any existing incidents.
// Persistence is best-effort: a load/save failure is returned but never
// prevents the in-memory tracker from working.
func Init(path string) error {
	storePath = path
	return load()
}

func load() error {
	if storePath == "" {
		return nil
	}
	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read incident store: %w", err)
	}

	var loaded map[string]persistedIncident
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to parse incident store: %w", err)
	}

	restored := make(map[string]*Incident, len(loaded))
	for id, p := range loaded {
		restored[id] = &Incident{
			ID:          p.ID,
			TraceID:     p.TraceID,
			Severity:    p.Severity,
			Title:       p.Title,
			Description: p.Description,
			RootCause:   p.RootCause,
			CreatedAt:   p.CreatedAt,
			ResolvedAt:  p.ResolvedAt,
			SLATarget:   time.Duration(p.SLATargetNs),
			Runbook:     p.Runbook,
		}
	}

	tracker.mu.Lock()
	defer tracker.mu.Unlock()
	tracker.incidents = restored
	return nil
}

func persist() error {
	if storePath == "" {
		return nil
	}

	tracker.mu.Lock()
	snapshot := make(map[string]persistedIncident, len(tracker.incidents))
	for id, i := range tracker.incidents {
		snapshot[id] = persistedIncident{
			ID:          i.ID,
			TraceID:     i.TraceID,
			Severity:    i.Severity,
			Title:       i.Title,
			Description: i.Description,
			RootCause:   i.RootCause,
			CreatedAt:   i.CreatedAt,
			ResolvedAt:  i.ResolvedAt,
			SLATargetNs: int64(i.SLATarget),
			Runbook:     i.Runbook,
		}
	}
	tracker.mu.Unlock()

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("failed to marshal incident store: %w", err)
	}

	if dir := filepath.Dir(storePath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create incident store directory: %w", err)
		}
	}
	if err := os.WriteFile(storePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write incident store: %w", err)
	}
	return nil
}

func (t *IncidentTracker) CreateIncident(traceID, severity, title, description, rootCause, runbook string) *Incident {
	t.mu.Lock()

	incident := &Incident{
		ID:          fmt.Sprintf("INC-%d", time.Now().UnixNano()),
		TraceID:     traceID,
		Severity:    severity,
		Title:       title,
		Description: description,
		RootCause:   rootCause,
		CreatedAt:   time.Now(),
		SLATarget:   15 * time.Minute, // 15 min SLA for errors
		Runbook:     runbook,
	}

	if severity == "critical" {
		incident.SLATarget = 5 * time.Minute
	}

	t.incidents[incident.ID] = incident
	t.mu.Unlock()

	persist() //nolint:errcheck // best-effort; in-memory tracker remains authoritative
	return incident
}

func (t *IncidentTracker) ResolveIncident(incidentID string) error {
	t.mu.Lock()

	incident, exists := t.incidents[incidentID]
	if !exists {
		t.mu.Unlock()
		return fmt.Errorf("incident not found: %s", incidentID)
	}

	now := time.Now()
	incident.ResolvedAt = &now
	t.mu.Unlock()

	persist() //nolint:errcheck // best-effort; in-memory tracker remains authoritative
	return nil
}

func (t *IncidentTracker) GetIncident(incidentID string) (*Incident, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	incident, exists := t.incidents[incidentID]
	if !exists {
		return nil, fmt.Errorf("incident not found: %s", incidentID)
	}

	return incident, nil
}

func (t *IncidentTracker) ListIncidents() []*Incident {
	t.mu.Lock()
	defer t.mu.Unlock()

	incidents := make([]*Incident, 0, len(t.incidents))
	for _, incident := range t.incidents {
		incidents = append(incidents, incident)
	}

	return incidents
}

func (t *IncidentTracker) IsSLABreached(incident *Incident) bool {
	if incident.ResolvedAt != nil {
		duration := incident.ResolvedAt.Sub(incident.CreatedAt)
		return duration > incident.SLATarget
	}

	// For unresolved incidents, check against current time
	duration := time.Since(incident.CreatedAt)
	return duration > incident.SLATarget
}

func (i *Incident) MarshalJSON() ([]byte, error) {
	type Alias Incident
	return json.Marshal(&struct {
		SLABreached      bool    `json:"sla_breached"`
		TimeToResolve    string  `json:"time_to_resolve,omitempty"`
		SLATargetMinutes float64 `json:"sla_target_minutes"`
		*Alias
	}{
		SLABreached:      tracker.IsSLABreached(i),
		TimeToResolve:    formatDuration(i.CreatedAt, i.ResolvedAt),
		SLATargetMinutes: i.SLATarget.Minutes(),
		Alias:            (*Alias)(i),
	})
}

func formatDuration(createdAt time.Time, resolvedAt *time.Time) string {
	if resolvedAt == nil {
		return ""
	}
	duration := resolvedAt.Sub(createdAt)
	return fmt.Sprintf("%.1f minutes", duration.Minutes())
}

func GetTracker() *IncidentTracker {
	return tracker
}
