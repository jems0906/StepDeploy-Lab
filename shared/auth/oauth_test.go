package auth

import "testing"

func TestGenerateAndValidateToken(t *testing.T) {
	token, err := GenerateToken("device-123", "ca-service", 3600)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if err := ValidateToken(token, "ca-service"); err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
}

func TestValidateTokenRejectsExpiredAndWrongAudience(t *testing.T) {
	expired, err := GenerateToken("device-123", "ca-service", -1)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateToken(expired, "ca-service"); err == nil {
		t.Fatal("expected expired token to be rejected")
	}

	valid, err := GenerateToken("device-123", "other-service", 3600)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateToken(valid, "ca-service"); err == nil {
		t.Fatal("expected wrong audience to be rejected")
	}
}
