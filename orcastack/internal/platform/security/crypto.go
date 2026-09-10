package security

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
	"time"
)

const (
	SigningAlgorithm     = "ed25519"
	VerificationAlgorithm = "ed25519"
	HashAlgorithm        = "sha256"
)

type Keys struct {
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
}

type Attestation struct {
	RepositoryIdentity string `json:"repository_identity"`
	ComponentIdentity  string `json:"component_identity"`
	ProcessIdentity    string `json:"process_identity"`
	BinaryHash         string `json:"binary_hash"`
	BuildHash          string `json:"build_hash"`
	SignedAt           string `json:"signed_at"`
	Signature          string `json:"signature,omitempty"`
}

func LoadKeys(privateKeyPath, publicKeyPath string) (Keys, error) {
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return Keys{}, err
	}

	publicKey, err := loadPublicKey(publicKeyPath)
	if err != nil {
		return Keys{}, err
	}

	return Keys{Private: privateKey, Public: publicKey}, nil
}

func HashBytes(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("sha256:%x", sum[:])
}

func HashFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %s: %w", path, err)
	}
	return HashBytes(content), nil
}

func NewAttestation(repositoryIdentity, componentIdentity, processIdentity, binaryHash, buildHash string) Attestation {
	return Attestation{
		RepositoryIdentity: repositoryIdentity,
		ComponentIdentity:  componentIdentity,
		ProcessIdentity:    processIdentity,
		BinaryHash:         binaryHash,
		BuildHash:          buildHash,
		SignedAt:           time.Now().UTC().Format(time.RFC3339),
	}
}

func (a Attestation) CanonicalPayload() string {
	parts := []string{
		"repo=" + a.RepositoryIdentity,
		"component=" + a.ComponentIdentity,
		"process=" + a.ProcessIdentity,
		"binary=" + a.BinaryHash,
		"build=" + a.BuildHash,
		"signed_at=" + a.SignedAt,
	}
	return strings.Join(parts, "\n")
}

func SignAttestation(keys Keys, attestation Attestation) (Attestation, error) {
	if len(keys.Private) == 0 {
		return Attestation{}, fmt.Errorf("private signing key is required")
	}

	copy := attestation
	signature := ed25519.Sign(keys.Private, []byte(copy.CanonicalPayload()))
	copy.Signature = base64.StdEncoding.EncodeToString(signature)
	return copy, nil
}

func VerifyAttestation(keys Keys, attestation Attestation) error {
	if len(keys.Public) == 0 {
		return fmt.Errorf("public verification key is required")
	}
	if attestation.Signature == "" {
		return fmt.Errorf("attestation signature is required")
	}

	signature, err := base64.StdEncoding.DecodeString(attestation.Signature)
	if err != nil {
		return fmt.Errorf("decode attestation signature: %w", err)
	}

	if !ed25519.Verify(keys.Public, []byte(attestation.CanonicalPayload()), signature) {
		return fmt.Errorf("attestation verification failed")
	}

	return nil
}

func loadPrivateKey(path string) (ed25519.PrivateKey, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read private key %s: %w", path, err)
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, fmt.Errorf("private key %s is not PEM encoded", path)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key %s: %w", path, err)
	}

	edKey, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key %s is not ed25519", path)
	}

	return edKey, nil
}

func loadPublicKey(path string) (ed25519.PublicKey, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read public key %s: %w", path, err)
	}

	block, _ := pem.Decode(content)
	if block == nil {
		return nil, fmt.Errorf("public key %s is not PEM encoded", path)
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key %s: %w", path, err)
	}

	edKey, ok := key.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key %s is not ed25519", path)
	}

	return edKey, nil
}