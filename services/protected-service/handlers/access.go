package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("Protected-Service")

var injectedFailures = make(map[string]bool)

type ProtectedResource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Secret   string `json:"secret"`
	ClientDN string `json:"client_dn"`
}

func ResourceHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Resource request from %s", r.RemoteAddr)

	// Check for injected service-unavailable failure
	if injectedFailures["service-unavailable"] {
		logger.Warn("Injected failure: service unavailable")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	// Get client certificate info
	clientDN := "unknown"
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		cert := r.TLS.PeerCertificates[0]
		clientDN = cert.Subject.String()
		logger.Info("Request authenticated with client cert: %s", clientDN)
	} else {
		logger.Warn("Request without valid client certificate")
		http.Error(w, "Client certificate required", http.StatusForbidden)
		return
	}

	resource := ProtectedResource{
		ID:       "secret-123",
		Name:     "Confidential Data",
		Secret:   "This is classified information",
		ClientDN: clientDN,
	}

	logger.Info("Protected resource accessed by: %s", clientDN)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}
