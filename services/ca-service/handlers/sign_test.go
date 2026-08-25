package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func startStubIdP(t *testing.T, response map[string]interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestSignHandlerSucceedsWithActiveToken(t *testing.T) {
	stub := startStubIdP(t, map[string]interface{}{
		"active": true, "sub": "device-1", "audience": "ca-service", "expires_in": 3600,
	})
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	idpURL = stub.URL

	body, _ := json.Marshal(SignRequest{Token: "any-token", ClientCN: "device-1.local", ValidityDays: 1})
	req := httptest.NewRequest(http.MethodPost, "/certs/sign", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	SignHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp SignResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Certificate == "" || resp.PrivateKey == "" || resp.CAChain == "" {
		t.Fatalf("expected populated certificate response, got %+v", resp)
	}
}

func TestSignHandlerRejectsInactiveToken(t *testing.T) {
	stub := startStubIdP(t, map[string]interface{}{"active": false})
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	idpURL = stub.URL

	body, _ := json.Marshal(SignRequest{Token: "expired-token", ClientCN: "device-1.local", ValidityDays: 1})
	req := httptest.NewRequest(http.MethodPost, "/certs/sign", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	SignHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for inactive token, got %d", rec.Code)
	}
}

func TestSignHandlerRejectsBadClientCertInjection(t *testing.T) {
	stub := startStubIdP(t, map[string]interface{}{
		"active": true, "sub": "device-1", "audience": "ca-service", "expires_in": 3600,
	})
	if err := Init(); err != nil {
		t.Fatal(err)
	}
	idpURL = stub.URL

	injectedFailures["bad-client-cert"] = true
	defer delete(injectedFailures, "bad-client-cert")

	body, _ := json.Marshal(SignRequest{Token: "any-token", ClientCN: "device-1.local", ValidityDays: 1})
	req := httptest.NewRequest(http.MethodPost, "/certs/sign", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	SignHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for injected bad-client-cert failure, got %d", rec.Code)
	}
}
