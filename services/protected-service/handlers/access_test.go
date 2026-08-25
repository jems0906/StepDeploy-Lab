package handlers

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/stepdeploy-lab/shared/certs"
)

func TestResourceHandlerRequiresClientCert(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rec := httptest.NewRecorder()

	ResourceHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without client cert, got %d", rec.Code)
	}
}

func TestResourceHandlerSucceedsWithClientCert(t *testing.T) {
	root, err := certs.GenerateRootCA("test-root", 1)
	if err != nil {
		t.Fatal(err)
	}
	client, err := certs.GenerateClientCert(root, "device-123", 1)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	req.TLS = &tls.ConnectionState{PeerCertificates: []*x509.Certificate{client.Certificate}}
	rec := httptest.NewRecorder()

	ResourceHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with client cert, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp ProtectedResource
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.ClientDN == "" || resp.ClientDN == "unknown" {
		t.Fatalf("expected client DN to be extracted from cert, got %q", resp.ClientDN)
	}
}

func TestResourceHandlerServiceUnavailableInjection(t *testing.T) {
	injectedFailures["service-unavailable"] = true
	defer delete(injectedFailures, "service-unavailable")

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	rec := httptest.NewRecorder()

	ResourceHandler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 for injected failure, got %d", rec.Code)
	}
}
