package provisioner

import (
	"fmt"
	"sync"
	"time"

	"github.com/example/stepdeploy-lab/services/ca-service/internal/ca"
	"github.com/example/stepdeploy-lab/shared/auth"
	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("Provisioner")

type OAuthProvisioner struct {
	ca             *ca.CA
	oauthValidator func(token string) (*auth.Token, error)
	mu             sync.Mutex
	issuedCerts    map[string]time.Time
}

// New creates a new OAuth provisioner
func New(ca *ca.CA, oauthValidator func(token string) (*auth.Token, error)) *OAuthProvisioner {
	return &OAuthProvisioner{ca: ca, oauthValidator: oauthValidator, issuedCerts: make(map[string]time.Time)}
}

// ProvisionCert provisions a certificate using an OAuth token
func (p *OAuthProvisioner) ProvisionCert(oauthToken string, clientCN string, validityDays int) (string, string, error) {
	logger.Info("Certificate provisioning request for: %s", clientCN)

	token, err := p.oauthValidator(oauthToken)
	if err != nil {
		logger.Error("Token validation failed: %v", err)
		return "", "", fmt.Errorf("token validation failed: %w", err)
	}
	if token.IsExpired() {
		logger.Warn("Attempted to use expired token for %s", clientCN)
		return "", "", fmt.Errorf("token has expired")
	}

	certKeyPair, err := p.ca.IssueCert(clientCN, validityDays)
	if err != nil {
		logger.Error("Certificate issuance failed: %v", err)
		return "", "", fmt.Errorf("certificate issuance failed: %w", err)
	}

	p.mu.Lock()
	p.issuedCerts[clientCN] = time.Now()
	p.mu.Unlock()

	logger.Info("Certificate provisioned successfully for: %s", clientCN)
	return certKeyPair.CertPEM, certKeyPair.KeyPEM, nil
}

// GetIssuedCertCount returns the number of issued certificates
func (p *OAuthProvisioner) GetIssuedCertCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.issuedCerts)
}
