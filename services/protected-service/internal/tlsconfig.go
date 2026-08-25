package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/example/stepdeploy-lab/shared/certs"
)

// Create generates the server's demo certificate and configures mTLS against
// the deployment CA's trusted root.
func Create() (*tls.Config, error) {
	// Generate a self-signed certificate for the server
	serverCert, err := certs.GenerateRootCA("protected-service", 365)
	if err != nil {
		return nil, fmt.Errorf("failed to generate server cert: %w", err)
	}

	// Parse server certificate
	cert, err := tls.X509KeyPair([]byte(serverCert.CertPEM), []byte(serverCert.KeyPEM))
	if err != nil {
		return nil, fmt.Errorf("failed to parse server cert: %w", err)
	}

	caCertPool, err := loadTrustedRoot()
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// VerifyClientCertIfGiven (not Require) so /health can be probed without a
		// client cert; the resource handler still enforces the cert requirement itself.
		ClientAuth: tls.VerifyClientCertIfGiven,
		ClientCAs:  caCertPool,
		MinVersion: tls.VersionTLS12,
	}

	return tlsConfig, nil
}

func loadTrustedRoot() (*x509.CertPool, error) {
	caURL := os.Getenv("CA_URL")
	if caURL == "" {
		caURL = "http://localhost:8002"
	}
	rootURL := strings.TrimRight(caURL, "/") + "/certs/root"
	client := &http.Client{Timeout: 5 * time.Second}
	var lastErr error
	for attempt := 1; attempt <= 10; attempt++ {
		resp, err := client.Get(rootURL)
		if err == nil {
			rootPEM, readErr := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && readErr == nil {
				pool := x509.NewCertPool()
				if pool.AppendCertsFromPEM(rootPEM) {
					return pool, nil
				}
				lastErr = fmt.Errorf("failed to parse CA root certificate")
			} else if readErr != nil {
				lastErr = fmt.Errorf("failed to read CA root: %w", readErr)
			} else {
				lastErr = fmt.Errorf("CA root request returned status %d", resp.StatusCode)
			}
		} else {
			lastErr = fmt.Errorf("failed to fetch CA root: %w", err)
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("CA root unavailable after retries: %w", lastErr)
}
