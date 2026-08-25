package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	tracing "github.com/example/stepdeploy-lab/services/diagnostics-api/internal"
)

func TracesHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Traces request from %s", r.RemoteAddr)

	traces := tracing.ListAll()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"traces": traces,
		"count":  len(traces),
	})
}

func TraceDetailHandler(w http.ResponseWriter, r *http.Request) {
	traceID := r.URL.Query().Get("trace_id")
	if traceID == "" {
		http.Error(w, "trace_id query parameter required", http.StatusBadRequest)
		return
	}

	logger.Debug("Trace detail request for: %s", traceID)

	trace, err := tracing.Get(traceID)
	if err != nil {
		logger.Warn("Trace not found: %s", traceID)
		http.Error(w, fmt.Sprintf("Trace not found: %s", traceID), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trace)
}

// TraceReportHandler lets other services (which run as separate processes
// with their own in-memory trace store) publish a completed trace here.
func TraceReportHandler(w http.ResponseWriter, r *http.Request) {
	var trace tracing.Trace
	if err := json.NewDecoder(r.Body).Decode(&trace); err != nil {
		http.Error(w, "Invalid trace payload", http.StatusBadRequest)
		return
	}
	if trace.TraceID == "" {
		http.Error(w, "trace_id is required", http.StatusBadRequest)
		return
	}

	tracing.Store(&trace)
	logger.Info("Trace reported: %s (status=%s)", trace.TraceID, trace.Status)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "stored"})
}
