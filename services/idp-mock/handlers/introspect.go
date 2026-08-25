package handlers

import (
	"encoding/json"
	"net/http"
)

type IntrospectRequest struct {
	Token string `json:"token"`
}

type IntrospectResponse struct {
	Active    bool   `json:"active"`
	Subject   string `json:"sub,omitempty"`
	Audience  string `json:"audience,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
}

// IntrospectHandler lets other services (ca-service) validate a bearer token
// against the IdP instead of trusting the token string blindly.
func IntrospectHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req IntrospectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	token, ok := lookupToken(req.Token)
	if !ok || token.IsExpired() {
		json.NewEncoder(w).Encode(IntrospectResponse{Active: false})
		return
	}

	json.NewEncoder(w).Encode(IntrospectResponse{
		Active:    true,
		Subject:   token.Subject,
		Audience:  token.Audience,
		ExpiresIn: token.ExpiresIn,
	})
}
