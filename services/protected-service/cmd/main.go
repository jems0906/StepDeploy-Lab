package main

import (
	"net/http"
	"os"

	"github.com/example/stepdeploy-lab/services/protected-service/handlers"
	tlsconfig "github.com/example/stepdeploy-lab/services/protected-service/internal"
	"github.com/example/stepdeploy-lab/shared/logging"
	"github.com/gorilla/mux"
)

var logger = logging.New("Protected-Service")

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/api/resource", handlers.ResourceHandler).Methods("GET")
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	router.HandleFunc("/test/inject-failure", handlers.InjectFailureHandler).Methods("POST")
	router.HandleFunc("/test/reset-failures", handlers.ResetFailuresHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8004"
	}

	logger.Info("Starting Protected Service on port %s (mTLS enabled)", port)

	tlsCfg, err := tlsconfig.Create()
	if err != nil {
		logger.Error("Failed to initialize TLS: %v", err)
		return
	}

	server := &http.Server{
		Addr:      ":" + port,
		Handler:   router,
		TLSConfig: tlsCfg,
	}

	if err := server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
		logger.Error("Server error: %v", err)
	}
}
