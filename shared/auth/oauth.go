package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// Token represents an OAuth/OIDC token
type Token struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"`
	Scope       string    `json:"scope"`
	IssuedAt    time.Time `json:"issued_at"`
	Audience    string    `json:"audience"`
	Subject     string    `json:"sub"`
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	expireTime := t.IssuedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
	return time.Now().After(expireTime)
}

// GenerateToken creates a new OAuth token
func GenerateToken(subject, audience string, expiresIn int) (*Token, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &Token{
		AccessToken: base64.URLEncoding.EncodeToString(tokenBytes),
		TokenType:   "Bearer",
		ExpiresIn:   expiresIn,
		Scope:       "openid profile email",
		IssuedAt:    time.Now(),
		Subject:     subject,
		Audience:    audience,
	}, nil
}

// ValidateToken checks if a token is valid
func ValidateToken(token *Token, expectedAudience string) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}
	if token.IsExpired() {
		return fmt.Errorf("token has expired")
	}
	if token.Audience != expectedAudience {
		return fmt.Errorf("token audience mismatch: expected %s, got %s", expectedAudience, token.Audience)
	}
	return nil
}

// UserInfo represents user information from the IdP
type UserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
