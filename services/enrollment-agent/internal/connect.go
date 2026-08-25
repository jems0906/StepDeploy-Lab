package agent

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"

	"github.com/example/stepdeploy-lab/shared/tracing"
)

// connectToProtectedService builds an mTLS client from the issued certificate
// and requests the protected resource, honoring any injected connection failures.
func connectToProtectedService(signResp SignResponse) (*http.Response, error) {
	clientCert, err := tls.X509KeyPair([]byte(signResp.Certificate), []byte(signResp.PrivateKey))
	if err != nil {
		return nil, err
	}

	// Include the intermediate CA in the client chain for the mTLS handshake.
	if block, _ := pem.Decode([]byte(signResp.CAChain)); block != nil {
		if intermediate, parseErr := x509.ParseCertificate(block.Bytes); parseErr == nil && intermediate.IsCA {
			clientCert.Certificate = append(clientCert.Certificate, intermediate.Raw)
		}
	}

	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: true,
	}
	if injectedFailures["missing-ca-trust"] {
		tlsConfig.InsecureSkipVerify = false
		tlsConfig.RootCAs = x509.NewCertPool()
	}

	requestURL := protectedServiceURL
	if injectedFailures["dns-mismatch"] {
		requestURL = "https://stepdeploy-invalid.local:8004"
	}

	mtlsClient := &http.Client{Transport: &http.Transport{TLSClientConfig: tlsConfig}}
	protectedReq, err := http.NewRequest("GET", requestURL+"/api/resource", nil)
	if err != nil {
		return nil, err
	}
	return mtlsClient.Do(protectedReq)
}

// reportTrace forwards the trace to diagnostics-api, since each service is a
// separate process with its own in-memory trace store.
func reportTrace(trace *tracing.Trace) {
	body, err := json.Marshal(trace)
	if err != nil {
		logger.Error("Failed to marshal trace for reporting: %v", err)
		return
	}
	resp, err := http.Post(diagnosticsURL+"/traces/report", "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Warn("Failed to report trace to diagnostics-api: %v", err)
		return
	}
	resp.Body.Close()
}
