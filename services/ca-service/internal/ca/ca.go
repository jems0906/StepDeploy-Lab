package ca

import (
	"fmt"

	"github.com/example/stepdeploy-lab/shared/certs"
	"github.com/example/stepdeploy-lab/shared/logging"
)

var logger = logging.New("CA")

// CA manages the certificate authority
type CA struct {
	RootCA         *certs.KeyPair
	IntermediateCA *certs.KeyPair
}

// New creates a new CA with root and intermediate certificates
func New() (*CA, error) {
	logger.Info("Initializing Certificate Authority")

	rootCA, err := certs.GenerateRootCA("StepDeploy Root CA", 3650)
	if err != nil {
		return nil, fmt.Errorf("failed to generate root CA: %w", err)
	}
	logger.Info("Root CA generated")

	intermediateCA, err := certs.GenerateIntermediateCA(rootCA, "StepDeploy Intermediate CA", 1825)
	if err != nil {
		return nil, fmt.Errorf("failed to generate intermediate CA: %w", err)
	}
	logger.Info("Intermediate CA generated")

	return &CA{RootCA: rootCA, IntermediateCA: intermediateCA}, nil
}

// IssueCert issues a client certificate
func (ca *CA) IssueCert(clientCN string, validityDays int) (*certs.KeyPair, error) {
	logger.Info("Issuing certificate for: %s", clientCN)

	clientCert, err := certs.GenerateClientCert(ca.IntermediateCA, clientCN, validityDays)
	if err != nil {
		return nil, fmt.Errorf("failed to issue client certificate: %w", err)
	}

	logger.Info("Certificate issued with fingerprint: %s", certs.GetCertFingerprint(clientCert.CertPEM))
	return clientCert, nil
}

// VerifyClientCert verifies a client certificate
func (ca *CA) VerifyClientCert(clientCertPEM string) error {
	return certs.VerifyClientCert(clientCertPEM, ca.IntermediateCA.CertPEM)
}
