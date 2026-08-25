package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal/health"
	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("Diagnostics-API")

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health.GetDependencyMap())
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "diagnostics-api",
	})
}
