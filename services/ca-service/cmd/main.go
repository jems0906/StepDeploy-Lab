package main

import (
	"net/http"
	"os"

	"github.com/example/stepdeploy-lab/services/ca-service/handlers"
	"github.com/example/stepdeploy-lab/shared/logging"
	"github.com/gorilla/mux"
)

var logger = logging.New("CA-Service")

func main() {
	if err := handlers.Init(); err != nil {
		logger.Error("Failed to initialize CA: %v", err)
		os.Exit(1)
	}

	router := mux.NewRouter()

	router.HandleFunc("/certs/sign", handlers.SignHandler).Methods("POST")
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	router.HandleFunc("/certs/root", handlers.RootCertHandler).Methods("GET")
	router.HandleFunc("/test/inject-failure", handlers.InjectFailureHandler).Methods("POST")
	router.HandleFunc("/test/reset-failures", handlers.ResetFailuresHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	logger.Info("Starting CA Service on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("Server error: %v", err)
	}
}
