package main

import (
	"net/http"
	"os"

	"github.com/example/stepdeploy-lab/services/enrollment-agent/handlers"
	internal "github.com/example/stepdeploy-lab/services/enrollment-agent/internal"
	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("Enrollment-Agent")

func main() {
	internal.Init()

	http.HandleFunc("/enroll", handlers.EnrollHandler)
	http.HandleFunc("/health", handlers.HealthHandler)
	http.HandleFunc("/test/inject-failure", handlers.InjectFailureHandler)
	http.HandleFunc("/test/reset-failures", handlers.ResetFailuresHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8003"
	}

	logger.Info("Starting Enrollment Agent on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		logger.Error("Server error: %v", err)
	}
}
