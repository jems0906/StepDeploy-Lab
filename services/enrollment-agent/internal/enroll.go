package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/example/stepdeploy-lab/shared/logging"
	"github.com/example/stepdeploy-lab/shared/tracing"
)

var logger = logging.New("Enrollment-Agent")

type TokenRequest struct {
	GrantType string `json:"grant_type"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Audience  string `json:"audience"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	Audience    string `json:"audience"`
}

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

type EnrollmentStatus struct {
	Status   string `json:"status"`
	TokenOK  bool   `json:"token_ok"`
	CertOK   bool   `json:"cert_ok"`
	AccessOK bool   `json:"access_ok"`
	TraceID  string `json:"trace_id"`
	Message  string `json:"message"`
}

var (
	idpURL              string
	caURL               string
	protectedServiceURL string
	diagnosticsURL      string
)

var injectedFailures = make(map[string]bool)

// Init reads service URLs from the environment. Must be called before serving requests.
func Init() {
	idpURL = os.Getenv("IDP_URL")
	if idpURL == "" {
		idpURL = "http://localhost:8001"
	}

	caURL = os.Getenv("CA_URL")
	if caURL == "" {
		caURL = "http://localhost:8002"
	}

	protectedServiceURL = os.Getenv("PROTECTED_SERVICE_URL")
	if protectedServiceURL == "" {
		protectedServiceURL = "https://localhost:8004"
	}

	diagnosticsURL = os.Getenv("DIAGNOSTICS_URL")
	if diagnosticsURL == "" {
		diagnosticsURL = "http://localhost:8080"
	}

	logger.Info("Enrollment Agent initialized")
	logger.Info("IdP URL: %s", idpURL)
	logger.Info("CA URL: %s", caURL)
	logger.Info("Protected Service URL: %s", protectedServiceURL)
}

func EnrollHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Enrollment request started")

	traceID := tracing.GenerateTraceID()
	trace := tracing.NewTrace(traceID)
	trace.AddEvent("enrollment-agent", "start", "Enrollment process started", map[string]interface{}{
		"timestamp": time.Now(),
	})

	status := EnrollmentStatus{
		Status:   "failed",
		TraceID:  traceID,
		TokenOK:  false,
		CertOK:   false,
		AccessOK: false,
	}
	defer func() {
		tracing.Store(trace)
		reportTrace(trace)
	}()

	// Step 1: Get OAuth token from IdP
	logger.Info("Step 1: Requesting OAuth token from IdP")
	trace.AddEvent("enrollment-agent", "token-request", "Requesting OAuth token", nil)

	tokenReq := TokenRequest{
		GrantType: "password",
		Username:  "device-123",
		Password:  "device-password",
		Audience:  "ca-service",
	}

	tokenReqBody, _ := json.Marshal(tokenReq)
	resp, err := http.Post(idpURL+"/token", "application/json", bytes.NewBuffer(tokenReqBody))
	if err != nil {
		logger.Error("Failed to get token: %v", err)
		trace.AddEvent("enrollment-agent", "token-failed", fmt.Sprintf("Token request failed: %v", err), nil)
		trace.SetStatus("failed", "IdP unreachable")
		status.Message = fmt.Sprintf("Token request failed: %v", err)
		respondWithStatus(w, status)
		return
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		logger.Error("Failed to decode token response: %v", err)
		trace.AddEvent("enrollment-agent", "token-decode-failed", "Failed to decode token response", nil)
		trace.SetStatus("failed", "Token response invalid")
		status.Message = fmt.Sprintf("Token decode failed: %v", err)
		respondWithStatus(w, status)
		return
	}

	// Check for injected failures, and validate the actual token response so
	// injecting the failure into the IdP (not just this agent) also works.
	if injectedFailures["expired-oauth-token"] || injectedFailures["clock-skew"] || tokenResp.ExpiresIn <= 0 {
		logger.Warn("Token is expired")
		trace.AddEvent("enrollment-agent", "token-expired", "Token validation failed due to expiration or clock skew", nil)
		trace.SetStatus("failed", "Token expired or clock skew detected")
		status.Message = "OAuth token is expired or clocks are out of sync"
		respondWithStatus(w, status)
		return
	}

	if injectedFailures["wrong-oauth-audience"] || (tokenResp.Audience != "" && tokenResp.Audience != tokenReq.Audience) {
		logger.Warn("Token audience mismatch")
		trace.AddEvent("enrollment-agent", "audience-mismatch", "Token validation: audience mismatch", nil)
		trace.SetStatus("failed", "Audience mismatch")
		status.Message = "Token audience does not match"
		respondWithStatus(w, status)
		return
	}

	status.TokenOK = true
	trace.AddEvent("enrollment-agent", "token-ok", "OAuth token obtained successfully", map[string]interface{}{
		"expires_in": tokenResp.ExpiresIn,
	})
	logger.Info("Step 1 OK: Token obtained (expires in %d seconds)", tokenResp.ExpiresIn)

	// Step 2: Request certificate from CA
	logger.Info("Step 2: Requesting certificate from CA")
	trace.AddEvent("enrollment-agent", "cert-request", "Requesting certificate", nil)

	signReq := SignRequest{
		Token:        tokenResp.AccessToken,
		ClientCN:     "device-123.stepdeploy.local",
		ValidityDays: 1,
	}

	signReqBody, _ := json.Marshal(signReq)
	resp, err = http.Post(caURL+"/certs/sign", "application/json", bytes.NewBuffer(signReqBody))
	if err != nil {
		logger.Error("Failed to get certificate: %v", err)
		trace.AddEvent("enrollment-agent", "cert-request-failed", fmt.Sprintf("Certificate request failed: %v", err), nil)
		trace.SetStatus("failed", "CA unreachable")
		status.Message = fmt.Sprintf("Certificate request failed: %v", err)
		respondWithStatus(w, status)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Certificate request failed with status %d: %s", resp.StatusCode, string(body))
		trace.AddEvent("enrollment-agent", "cert-request-error", fmt.Sprintf("Certificate request error: %s", string(body)), nil)
		trace.SetStatus("failed", "Certificate signing failed")
		status.Message = fmt.Sprintf("Certificate signing failed: %s", string(body))
		respondWithStatus(w, status)
		return
	}

	var signResp SignResponse
	if err := json.NewDecoder(resp.Body).Decode(&signResp); err != nil {
		logger.Error("Failed to decode certificate response: %v", err)
		trace.AddEvent("enrollment-agent", "cert-decode-failed", "Failed to decode certificate response", nil)
		trace.SetStatus("failed", "Certificate response invalid")
		status.Message = fmt.Sprintf("Certificate decode failed: %v", err)
		respondWithStatus(w, status)
		return
	}

	status.CertOK = true
	trace.AddEvent("enrollment-agent", "cert-ok", "Certificate obtained successfully", nil)
	logger.Info("Step 2 OK: Certificate obtained")

	// Step 3: Try to access protected service
	logger.Info("Step 3: Accessing protected service with client certificate")
	trace.AddEvent("enrollment-agent", "protected-access", "Attempting to access protected resource", nil)

	protectedResp, err := connectToProtectedService(signResp)
	if err != nil {
		logger.Error("Failed to access protected service: %v", err)
		trace.AddEvent("enrollment-agent", "protected-access-failed", fmt.Sprintf("Protected service access failed: %v", err), nil)
		trace.SetStatus("failed", "Protected service unreachable")
		status.Message = fmt.Sprintf("Protected service access failed: %v", err)
		respondWithStatus(w, status)
		return
	}
	defer protectedResp.Body.Close()

	if injectedFailures["service-unavailable"] {
		logger.Warn("Injected failure: protected service unavailable")
		trace.AddEvent("enrollment-agent", "service-unavailable", "Protected service returned 503", nil)
		trace.SetStatus("failed", "Protected service unavailable")
		status.Message = "Protected service unavailable"
		respondWithStatus(w, status)
		return
	}

	if protectedResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(protectedResp.Body)
		logger.Error("Protected service returned status %d: %s", protectedResp.StatusCode, string(body))
		trace.AddEvent("enrollment-agent", "protected-access-error", fmt.Sprintf("Protected service error: %d", protectedResp.StatusCode), nil)
		trace.SetStatus("failed", "Protected service access denied")
		status.Message = fmt.Sprintf("Protected service access denied: %d", protectedResp.StatusCode)
		respondWithStatus(w, status)
		return
	}

	status.AccessOK = true
	status.Status = "success"
	trace.AddEvent("enrollment-agent", "success", "Enrollment completed successfully", nil)
	trace.SetStatus("success", "")
	logger.Info("Step 3 OK: Protected resource accessed")

	respondWithStatus(w, status)
}

func respondWithStatus(w http.ResponseWriter, status EnrollmentStatus) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "enrollment-agent",
	})
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
