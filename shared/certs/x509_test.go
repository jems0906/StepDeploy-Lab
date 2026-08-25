package certs

import "testing"

func TestGenerateAndVerifyCertificateChain(t *testing.T) {
	root, err := GenerateRootCA("root", 365)
	if err != nil {
		t.Fatal(err)
	}
	intermediate, err := GenerateIntermediateCA(root, "intermediate", 365)
	if err != nil {
		t.Fatal(err)
	}
	client, err := GenerateClientCert(intermediate, "device", 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyClientCert(client.CertPEM, intermediate.CertPEM); err != nil {
		t.Fatalf("VerifyClientCert returned error: %v", err)
	}
	if GetCertFingerprint(client.CertPEM) == "" {
		t.Fatal("expected certificate fingerprint")
	}
}
