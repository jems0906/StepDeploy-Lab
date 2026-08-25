package agent

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/stepdeploy-lab/shared/certs"
)

func startStub(t *testing.T, method, path string, respond func(w http.ResponseWriter)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method || r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		respond(w)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestEnrollHandlerFullHappyPath(t *testing.T) {
	root, err := certs.GenerateRootCA("test-root", 1)
	if err != nil {
		t.Fatal(err)
	}
	client, err := certs.GenerateClientCert(root, "device-under-test", 1)
	if err != nil {
		t.Fatal(err)
	}

	idpStub := startStub(t, http.MethodPost, "/token", func(w http.ResponseWriter) {
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "test-token", TokenType: "Bearer", ExpiresIn: 3600, Audience: "ca-service",
		})
	})
	caStub := startStub(t, http.MethodPost, "/certs/sign", func(w http.ResponseWriter) {
		json.NewEncoder(w).Encode(SignResponse{
			Certificate: client.CertPEM, PrivateKey: client.KeyPEM, CAChain: root.CertPEM,
		})
	})
	protectedStub := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/resource" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": "secret-123"})
	}))
	t.Cleanup(protectedStub.Close)
	diagnosticsStub := startStub(t, http.MethodPost, "/traces/report", func(w http.ResponseWriter) {
		json.NewEncoder(w).Encode(map[string]string{"status": "stored"})
	})

	idpURL = idpStub.URL
	caURL = caStub.URL
	protectedServiceURL = protectedStub.URL
	diagnosticsURL = diagnosticsStub.URL

	req := httptest.NewRequest(http.MethodPost, "/enroll", nil)
	rec := httptest.NewRecorder()
	EnrollHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var status EnrollmentStatus
	if err := json.NewDecoder(rec.Body).Decode(&status); err != nil {
		t.Fatal(err)
	}
	if status.Status != "success" || !status.TokenOK || !status.CertOK || !status.AccessOK {
		t.Fatalf("expected full success, got %+v", status)
	}
}

func TestEnrollHandlerFailsOnExpiredToken(t *testing.T) {
	idpStub := startStub(t, http.MethodPost, "/token", func(w http.ResponseWriter) {
		json.NewEncoder(w).Encode(TokenResponse{AccessToken: "expired", ExpiresIn: -1})
	})
	idpURL = idpStub.URL

	req := httptest.NewRequest(http.MethodPost, "/enroll", nil)
	rec := httptest.NewRecorder()
	EnrollHandler(rec, req)

	var status EnrollmentStatus
	json.NewDecoder(rec.Body).Decode(&status)
	if status.Status == "success" || status.TokenOK {
		t.Fatalf("expected enrollment to fail on expired token, got %+v", status)
	}
}
