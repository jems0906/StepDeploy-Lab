package certs

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

// KeyPair represents a certificate and its private key
type KeyPair struct {
	Certificate *x509.Certificate
	PrivateKey  *rsa.PrivateKey
	CertPEM     string
	KeyPEM      string
}

// GenerateRootCA generates a self-signed root CA certificate
func GenerateRootCA(commonName string, validityDays int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, validityDays),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return &KeyPair{
		Certificate: cert,
		PrivateKey:  privateKey,
		CertPEM:     string(certPEM),
		KeyPEM:      string(keyPEM),
	}, nil
}

// GenerateIntermediateCA generates a CA certificate signed by a parent CA.
func GenerateIntermediateCA(parent *KeyPair, commonName string, validityDays int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	template := x509.Certificate{
		SerialNumber:          big.NewInt(time.Now().UnixNano()),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(0, 0, validityDays),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, &template, parent.Certificate, &privateKey.PublicKey, parent.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create intermediate certificate: %w", err)
	}
	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intermediate certificate: %w", err)
	}

	return &KeyPair{
		Certificate: cert,
		PrivateKey:  privateKey,
		CertPEM:     string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})),
		KeyPEM:      string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})),
	}, nil
}

// GenerateClientCert generates a client certificate signed by the CA
func GenerateClientCert(caKeyPair *KeyPair, clientCN string, validityDays int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	serialNumber := big.NewInt(time.Now().UnixNano())

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: clientCN,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(0, 0, validityDays),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		&template,
		caKeyPair.Certificate,
		&privateKey.PublicKey,
		caKeyPair.PrivateKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	return &KeyPair{
		Certificate: cert,
		PrivateKey:  privateKey,
		CertPEM:     string(certPEM),
		KeyPEM:      string(keyPEM),
	}, nil
}

// VerifyClientCert verifies a client certificate against the CA
func VerifyClientCert(clientCertPEM string, caCertPEM string) error {
	// Parse client cert
	clientBlock, _ := pem.Decode([]byte(clientCertPEM))
	if clientBlock == nil {
		return fmt.Errorf("failed to parse client certificate")
	}

	clientCert, err := x509.ParseCertificate(clientBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse client certificate: %w", err)
	}

	// Parse CA cert
	caBlock, _ := pem.Decode([]byte(caCertPEM))
	if caBlock == nil {
		return fmt.Errorf("failed to parse CA certificate")
	}

	caCert, err := x509.ParseCertificate(caBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse CA certificate: %w", err)
	}

	// Check expiration
	if time.Now().After(clientCert.NotAfter) {
		return fmt.Errorf("client certificate has expired")
	}

	// Verify signature
	err = clientCert.CheckSignatureFrom(caCert)
	if err != nil {
		return fmt.Errorf("client certificate signature verification failed: %w", err)
	}

	return nil
}

// GetCertFingerprint returns the SHA256 fingerprint of a certificate
func GetCertFingerprint(certPEM string) string {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return ""
	}
	hash := sha256.Sum256(block.Bytes)
	return fmt.Sprintf("%x", hash)
}
