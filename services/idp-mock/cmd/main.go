package main

import (
	"net/http"
	"os"

	"github.com/example/stepdeploy-lab/services/idp-mock/handlers"
	"github.com/example/stepdeploy-lab/shared/logging"
	"github.com/gorilla/mux"
)

var logger = logging.New("IdP-Mock")

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/token", handlers.TokenHandler).Methods("POST")
	router.HandleFunc("/userinfo", handlers.UserInfoHandler).Methods("GET")
	router.HandleFunc("/introspect", handlers.IntrospectHandler).Methods("POST")
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET")
	router.HandleFunc("/test/inject-failure", handlers.InjectFailureHandler).Methods("POST")
	router.HandleFunc("/test/reset-failures", handlers.ResetFailuresHandler).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8001"
	}

	logger.Info("Starting IdP Mock service on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		logger.Error("Server error: %v", err)
	}
}
