package tracing

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TraceEvent represents a single event in a trace
type TraceEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	Service     string                 `json:"service"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Trace represents a distributed trace
type Trace struct {
	TraceID   string       `json:"trace_id"`
	Events    []TraceEvent `json:"events"`
	Status    string       `json:"status"`
	RootCause string       `json:"root_cause,omitempty"`
	mu        sync.Mutex
}

var traces = make(map[string]*Trace)
var tracesMu sync.Mutex

// storePath enables file-based persistence when set via Init. Empty by
// default so existing in-memory-only callers (and tests) are unaffected.
var storePath string

// Init enables file-based persistence at path and loads any existing traces.
// Persistence is best-effort: a load/save failure is returned but never
// prevents the in-memory store from working.
func Init(path string) error {
	storePath = path
	return load()
}

func load() error {
	if storePath == "" {
		return nil
	}
	// #nosec G304 -- storePath is configured by trusted service env in this lab.
	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read trace store: %w", err)
	}

	var loaded map[string]*Trace
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("failed to parse trace store: %w", err)
	}

	tracesMu.Lock()
	defer tracesMu.Unlock()
	traces = loaded
	return nil
}

func persist() error {
	if storePath == "" {
		return nil
	}

	tracesMu.Lock()
	data, err := json.Marshal(traces)
	tracesMu.Unlock()
	if err != nil {
		return fmt.Errorf("failed to marshal trace store: %w", err)
	}

	if dir := filepath.Dir(storePath); dir != "." {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create trace store directory: %w", err)
		}
	}
	// #nosec G304 -- storePath is configured by trusted service env in this lab.
	if err := os.WriteFile(storePath, data, 0o600); err != nil {
		return fmt.Errorf("failed to write trace store: %w", err)
	}
	return nil
}

// GenerateTraceID creates a new unique trace ID
func GenerateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// NewTrace creates a new trace
func NewTrace(traceID string) *Trace {
	return &Trace{
		TraceID: traceID,
		Events:  []TraceEvent{},
		Status:  "in-progress",
	}
}

// AddEvent adds an event to the trace
func (t *Trace) AddEvent(service, eventType, description string, metadata map[string]interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	event := TraceEvent{
		Timestamp:   time.Now(),
		Service:     service,
		EventType:   eventType,
		Description: description,
		Metadata:    metadata,
	}

	t.Events = append(t.Events, event)
}

// SetStatus sets the final status of the trace
func (t *Trace) SetStatus(status, rootCause string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Status = status
	t.RootCause = rootCause
}

// Store stores a trace in memory and persists it if Init was called.
func Store(trace *Trace) {
	tracesMu.Lock()
	traces[trace.TraceID] = trace
	tracesMu.Unlock()

	persist() //nolint:errcheck // best-effort; in-memory store remains authoritative
}

// Get retrieves a trace by ID
func Get(traceID string) (*Trace, error) {
	tracesMu.Lock()
	defer tracesMu.Unlock()

	trace, exists := traces[traceID]
	if !exists {
		return nil, fmt.Errorf("trace not found: %s", traceID)
	}

	return trace, nil
}

// ListAll returns all stored traces
func ListAll() []*Trace {
	tracesMu.Lock()
	defer tracesMu.Unlock()

	result := make([]*Trace, 0, len(traces))
	for _, trace := range traces {
		result = append(result, trace)
	}

	return result
}

// Clear removes all traces (useful for testing)
func Clear() {
	tracesMu.Lock()
	defer tracesMu.Unlock()
	traces = make(map[string]*Trace)
}
