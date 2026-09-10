package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestNewIdentityUsesOrcaPattern(t *testing.T) {
	identity, err := NewIdentity("service")
	if err != nil {
		t.Fatalf("NewIdentity returned error: %v", err)
	}

	if identity.ComponentType != "service" {
		t.Fatalf("expected component type service, got %q", identity.ComponentType)
	}

	if _, err := ParseIdentity(identity.Raw); err != nil {
		t.Fatalf("ParseIdentity rejected generated identity: %v", err)
	}
}

func TestParseIdentityRejectsInvalidFormat(t *testing.T) {
	if _, err := ParseIdentity("orcastack:service:not-a-uuid"); err == nil {
		t.Fatal("expected invalid identity format to fail")
	}
}

func TestAttestationSignAndVerify(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey returned error: %v", err)
	}

	attestation := NewAttestation(
		DefaultRepositoryIdentity,
		"orca:service:550e8400-e29b-41d4-a716-446655440000",
		"orca:process:1b4e28ba-2fa1-41d2-883f-0016d3cca427",
		"sha256:binary",
		"sha256:build",
	)

	signed, err := SignAttestation(Keys{Private: privateKey, Public: publicKey}, attestation)
	if err != nil {
		t.Fatalf("SignAttestation returned error: %v", err)
	}

	if err := VerifyAttestation(Keys{Public: publicKey}, signed); err != nil {
		t.Fatalf("VerifyAttestation returned error: %v", err)
	}
}