package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/example/stepdeploy-lab/services/ca-service/internal/ca"
	"github.com/example/stepdeploy-lab/services/ca-service/internal/provisioner"
	"github.com/example/stepdeploy-lab/shared/auth"
	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("CA-Service")
var caInstance *ca.CA
var provisionerInstance *provisioner.OAuthProvisioner
var idpURL string

var injectedFailures = make(map[string]bool)

type SignRequest struct {
	Token        string `json:"token"`
	ClientCN     string `json:"client_cn"`
	ValidityDays int    `json:"validity_days"`
}

type SignResponse struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
	CAChain     string `json:"ca_chain"`
}

// Init creates the CA and OAuth provisioner. Must be called before serving requests.
func Init() error {
	logger.Info("CA Service initializing")

	var err error
	caInstance, err = ca.New()
	if err != nil {
		return fmt.Errorf("failed to initialize CA: %w", err)
	}

	idpURL = os.Getenv("IDP_URL")
	if idpURL == "" {
		idpURL = "http://localhost:8001"
	}

	provisionerInstance = provisioner.New(caInstance, validateOAuthToken)
	return nil
}

func validateOAuthToken(tokenString string) (*auth.Token, error) {
	// Check for injected failures
	if injectedFailures["wrong-oauth-audience"] {
		logger.Warn("Injected failure: rejecting token due to audience mismatch")
		return nil, fmt.Errorf("token audience mismatch")
	}

	// Validate the token against the IdP's introspection endpoint rather than
	// trusting the bearer string.
	reqBody, _ := json.Marshal(map[string]string{"token": tokenString})
	resp, err := http.Post(idpURL+"/introspect", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to reach IdP for introspection: %w", err)
	}
	defer resp.Body.Close()

	var introspection struct {
		Active    bool   `json:"active"`
		Subject   string `json:"sub"`
		Audience  string `json:"audience"`
		ExpiresIn int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&introspection); err != nil {
		return nil, fmt.Errorf("failed to decode introspection response: %w", err)
	}

	if !introspection.Active {
		return nil, fmt.Errorf("token is not active (expired or unknown to IdP)")
	}
	if introspection.Audience != "ca-service" {
		return nil, fmt.Errorf("token audience mismatch: expected ca-service, got %s", introspection.Audience)
	}

	return &auth.Token{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   introspection.ExpiresIn,
		IssuedAt:    time.Now(),
		Audience:    introspection.Audience,
		Subject:     introspection.Subject,
	}, nil
}

func SignHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Sign request received from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var signReq SignRequest
	if err := json.NewDecoder(r.Body).Decode(&signReq); err != nil {
		logger.Error("Failed to decode sign request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if signReq.ValidityDays == 0 {
		signReq.ValidityDays = 1
	}

	// Check for injected failures
	if injectedFailures["bad-client-cert"] {
		logger.Warn("Injected failure: refusing to sign certificate")
		http.Error(w, "Certificate signing failed", http.StatusBadRequest)
		return
	}

	// Provision certificate
	certPEM, keyPEM, err := provisionerInstance.ProvisionCert(signReq.Token, signReq.ClientCN, signReq.ValidityDays)
	if err != nil {
		logger.Error("Provisioning failed: %v", err)
		http.Error(w, fmt.Sprintf("Provisioning failed: %v", err), http.StatusBadRequest)
		return
	}

	// Build CA chain (root + intermediate)
	caChain := caInstance.IntermediateCA.CertPEM + "\n" + caInstance.RootCA.CertPEM

	logger.Info("Certificate signed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignResponse{
		Certificate: certPEM,
		PrivateKey:  keyPEM,
		CAChain:     caChain,
	})
}
