package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/example/stepdeploy-lab/shared/auth"
	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("IdP-Mock")

var injectedFailures = make(map[string]bool)

// issuedTokens backs the /introspect endpoint so ca-service can validate
// tokens against the IdP instead of trusting the bearer string blindly.
var (
	issuedTokens   = make(map[string]*auth.Token)
	issuedTokensMu sync.Mutex
)

func storeToken(token *auth.Token) {
	issuedTokensMu.Lock()
	defer issuedTokensMu.Unlock()
	issuedTokens[token.AccessToken] = token
}

func lookupToken(accessToken string) (*auth.Token, bool) {
	issuedTokensMu.Lock()
	defer issuedTokensMu.Unlock()
	token, ok := issuedTokens[accessToken]
	return token, ok
}

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

func TokenHandler(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Token request received from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tokenReq TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&tokenReq); err != nil {
		logger.Error("Failed to decode token request: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Check for injected failures. The fake tokens below are stored just like
	// real ones so introspection reports them as expired/audience-mismatched.
	if injectedFailures["expired-oauth-token"] {
		logger.Warn("Injected failure: returning expired token")
		expired := &auth.Token{
			AccessToken: "expired-token-xyz",
			TokenType:   "Bearer",
			ExpiresIn:   -1,
			IssuedAt:    time.Now(),
			Scope:       "openid profile email",
			Audience:    tokenReq.Audience,
			Subject:     tokenReq.Username,
		}
		storeToken(expired)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: expired.AccessToken,
			TokenType:   expired.TokenType,
			ExpiresIn:   expired.ExpiresIn,
			Scope:       expired.Scope,
		})
		return
	}

	if injectedFailures["wrong-oauth-audience"] {
		logger.Warn("Injected failure: returning token with wrong audience")
		wrongAudience := &auth.Token{
			AccessToken: "wrong-audience-token",
			TokenType:   "Bearer",
			ExpiresIn:   3600,
			IssuedAt:    time.Now(),
			Scope:       "openid profile email",
			Audience:    "wrong-service",
			Subject:     tokenReq.Username,
		}
		storeToken(wrongAudience)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: wrongAudience.AccessToken,
			TokenType:   wrongAudience.TokenType,
			ExpiresIn:   wrongAudience.ExpiresIn,
			Scope:       wrongAudience.Scope,
			Audience:    wrongAudience.Audience,
		})
		return
	}

	// Generate valid token
	token, err := auth.GenerateToken(tokenReq.Username, tokenReq.Audience, 3600)
	if err != nil {
		logger.Error("Failed to generate token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	storeToken(token)

	logger.Info("Token issued for user: %s, audience: %s", token.Subject, token.Audience)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		ExpiresIn:   token.ExpiresIn,
		Scope:       token.Scope,
		Audience:    token.Audience,
	})
}
