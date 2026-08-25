package main

import (
	"net/http"
	"os"

	"github.com/example/stepdeploy-lab/services/diagnostics-api/handlers"
	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal"
	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal/health"
	"github.com/example/stepdeploy-lab/services/diagnostics-api/internal/incidents"
	"github.com/example/stepdeploy-lab/shared/logging"
	"github.com/gorilla/mux"
)

var logger = logging.New("Diagnostics-API")

func main() {
	health.Start()

	traceStorePath := os.Getenv("TRACE_STORE_PATH")
	if traceStorePath == "" {
		traceStorePath = "data/traces.json"
	}
	if err := internal.Init(traceStorePath); err != nil {
		logger.Warn("Trace persistence disabled: %v", err)
	}

	incidentStorePath := os.Getenv("INCIDENT_STORE_PATH")
	if incidentStorePath == "" {
		incidentStorePath = "data/incidents.json"
	}
	if err := incidents.Init(incidentStorePath); err != nil {
		logger.Warn("Incident persistence disabled: %v", err)
	}

	router := mux.NewRouter()

	router.HandleFunc("/health", handlers.StatusHandler).Methods("GET")
	router.HandleFunc("/dependencies", handlers.HealthCheckHandler).Methods("GET")
	router.HandleFunc("/traces", handlers.TracesHandler).Methods("GET")
	router.HandleFunc("/traces/detail", handlers.TraceDetailHandler).Methods("GET")
	router.HandleFunc("/traces/report", handlers.TraceReportHandler).Methods("POST")
	router.HandleFunc("/incidents", handlers.IncidentsHandler).Methods("GET")
	router.HandleFunc("/incidents/detail", handlers.IncidentDetailHandler).Methods("GET")
	router.HandleFunc("/incidents", handlers.CreateIncidentHandler).Methods("POST")
	router.HandleFunc("/incidents/resolve", handlers.ResolveIncidentHandler).Methods("POST")
	router.HandleFunc("/enroll", handlers.EnrollHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Add CORS headers
	router.Use(corsMiddleware)

	logger.Info("Starting Diagnostics API on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("Server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
