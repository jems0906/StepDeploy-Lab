package handlers

import (
	"net/http"

	agent "github.com/example/stepdeploy-lab/services/enrollment-agent/internal"
)

func EnrollHandler(w http.ResponseWriter, r *http.Request) {
	agent.EnrollHandler(w, r)
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	agent.StatusHandler(w, r)
}

func InjectFailureHandler(w http.ResponseWriter, r *http.Request) {
	agent.InjectFailureHandler(w, r)
}

func ResetFailuresHandler(w http.ResponseWriter, r *http.Request) {
	agent.ResetFailuresHandler(w, r)
}
