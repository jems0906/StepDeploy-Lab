package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenHandlerIssuesValidToken(t *testing.T) {
	body, _ := json.Marshal(TokenRequest{GrantType: "password", Username: "device-1", Audience: "ca-service"})
	req := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	TokenHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp TokenResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.AccessToken == "" || resp.Audience != "ca-service" || resp.ExpiresIn <= 0 {
		t.Fatalf("unexpected token response: %+v", resp)
	}
}

func TestTokenHandlerRejectsNonPost(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/token", nil)
	rec := httptest.NewRecorder()

	TokenHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestIntrospectRoundTrip(t *testing.T) {
	body, _ := json.Marshal(TokenRequest{GrantType: "password", Username: "device-1", Audience: "ca-service"})
	tokenReq := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
	tokenRec := httptest.NewRecorder()
	TokenHandler(tokenRec, tokenReq)

	var tokenResp TokenResponse
	json.NewDecoder(tokenRec.Body).Decode(&tokenResp)

	introspectBody, _ := json.Marshal(IntrospectRequest{Token: tokenResp.AccessToken})
	introspectReq := httptest.NewRequest(http.MethodPost, "/introspect", bytes.NewReader(introspectBody))
	introspectRec := httptest.NewRecorder()
	IntrospectHandler(introspectRec, introspectReq)

	var introspectResp IntrospectResponse
	if err := json.NewDecoder(introspectRec.Body).Decode(&introspectResp); err != nil {
		t.Fatal(err)
	}
	if !introspectResp.Active || introspectResp.Audience != "ca-service" {
		t.Fatalf("expected active token with matching audience, got %+v", introspectResp)
	}
}

func TestIntrospectUnknownTokenIsInactive(t *testing.T) {
	body, _ := json.Marshal(IntrospectRequest{Token: "never-issued"})
	req := httptest.NewRequest(http.MethodPost, "/introspect", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	IntrospectHandler(rec, req)

	var resp IntrospectResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Active {
		t.Fatal("expected unknown token to be reported inactive")
	}
}

func TestTokenHandlerExpiredInjectionIsReportedInactive(t *testing.T) {
	injectedFailures["expired-oauth-token"] = true
	defer delete(injectedFailures, "expired-oauth-token")

	body, _ := json.Marshal(TokenRequest{GrantType: "password", Username: "device-1", Audience: "ca-service"})
	tokenReq := httptest.NewRequest(http.MethodPost, "/token", bytes.NewReader(body))
	tokenRec := httptest.NewRecorder()
	TokenHandler(tokenRec, tokenReq)

	var tokenResp TokenResponse
	json.NewDecoder(tokenRec.Body).Decode(&tokenResp)
	if tokenResp.ExpiresIn >= 0 {
		t.Fatalf("expected injected expired token, got expires_in=%d", tokenResp.ExpiresIn)
	}

	introspectBody, _ := json.Marshal(IntrospectRequest{Token: tokenResp.AccessToken})
	introspectReq := httptest.NewRequest(http.MethodPost, "/introspect", bytes.NewReader(introspectBody))
	introspectRec := httptest.NewRecorder()
	IntrospectHandler(introspectRec, introspectReq)

	var introspectResp IntrospectResponse
	json.NewDecoder(introspectRec.Body).Decode(&introspectResp)
	if introspectResp.Active {
		t.Fatal("expected injected expired token to introspect as inactive")
	}
}
