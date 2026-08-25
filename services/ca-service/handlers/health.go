package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "ok",
		"service":      "ca-service",
		"issued_certs": provisionerInstance.GetIssuedCertCount(),
	})
}

func RootCertHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Root cert request from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "text/plain")
	io.WriteString(w, caInstance.RootCA.CertPEM)
}

func InjectFailureHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	failureType := req["failure_type"]
	injectedFailures[failureType] = true

	logger.Warn("Failure injected: %s", failureType)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":       "failure injected",
		"failure_type": failureType,
	})
}

func ResetFailuresHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	injectedFailures = make(map[string]bool)
	logger.Info("All failures reset")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "failures reset",
	})
}
