package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/orcastack/orcastackapi/internal/platform/security"
)

func main() {
	if len(os.Args) < 2 {
		fatalf("expected subcommand: genkeys or attest")
	}

	switch os.Args[1] {
	case "genkeys":
		runGenKeys(os.Args[2:])
	case "attest":
		runAttest(os.Args[2:])
	default:
		fatalf("unknown subcommand %q", os.Args[1])
	}
}

func runGenKeys(args []string) {
	flags := flag.NewFlagSet("genkeys", flag.ExitOnError)
	privateKeyPath := flags.String("private", "", "path to PKCS8 private key PEM")
	publicKeyPath := flags.String("public", "", "path to PKIX public key PEM")
	_ = flags.Parse(args)

	if *privateKeyPath == "" || *publicKeyPath == "" {
		fatalf("both --private and --public are required")
	}

	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fatalf("generate ed25519 key pair: %v", err)
	}

	privateDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		fatalf("marshal private key: %v", err)
	}
	publicDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fatalf("marshal public key: %v", err)
	}

	writePEM(*privateKeyPath, "PRIVATE KEY", privateDER)
	writePEM(*publicKeyPath, "PUBLIC KEY", publicDER)
}

func runAttest(args []string) {
	flags := flag.NewFlagSet("attest", flag.ExitOnError)
	repositoryIdentity := flags.String("repository-identity", security.DefaultRepositoryIdentity, "repository identity")
	componentIdentity := flags.String("component-identity", "", "component identity")
	processIdentity := flags.String("process-identity", "", "process identity")
	binaryPath := flags.String("binary-path", "", "path to binary or artifact to hash")
	buildReference := flags.String("build-reference", "", "string to hash for the build attestation")
	privateKeyPath := flags.String("private-key", "", "path to PKCS8 private key PEM")
	publicKeyPath := flags.String("public-key", "", "path to PKIX public key PEM")
	outputPath := flags.String("output", "", "path to write the attestation json")
	_ = flags.Parse(args)

	if *componentIdentity == "" || *binaryPath == "" || *buildReference == "" || *privateKeyPath == "" || *publicKeyPath == "" || *outputPath == "" {
		fatalf("--component-identity, --binary-path, --build-reference, --private-key, --public-key, and --output are required")
	}

	if _, err := security.ParseIdentity(*repositoryIdentity); err != nil {
		fatalf("repository identity: %v", err)
	}
	if _, err := security.ParseIdentity(*componentIdentity); err != nil {
		fatalf("component identity: %v", err)
	}

	process := *processIdentity
	if process == "" {
		identity, err := security.NewIdentity("process")
		if err != nil {
			fatalf("process identity: %v", err)
		}
		process = identity.Raw
	}
	if _, err := security.ParseIdentity(process); err != nil {
		fatalf("process identity: %v", err)
	}

	binaryHash, err := security.HashFile(*binaryPath)
	if err != nil {
		fatalf("binary hash: %v", err)
	}
	buildHash := security.HashBytes([]byte(*buildReference))

	keys, err := security.LoadKeys(*privateKeyPath, *publicKeyPath)
	if err != nil {
		fatalf("load keys: %v", err)
	}

	attestation, err := security.SignAttestation(keys, security.NewAttestation(*repositoryIdentity, *componentIdentity, process, binaryHash, buildHash))
	if err != nil {
		fatalf("sign attestation: %v", err)
	}
	if err := security.VerifyAttestation(keys, attestation); err != nil {
		fatalf("verify attestation: %v", err)
	}

	writeJSON(*outputPath, attestation)
}

func writePEM(path, blockType string, content []byte) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatalf("create parent directory for %s: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		fatalf("create %s: %v", path, err)
	}
	defer file.Close()

	if err := pem.Encode(file, &pem.Block{Type: blockType, Bytes: content}); err != nil {
		fatalf("write %s: %v", path, err)
	}
}

func writeJSON(path string, value any) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fatalf("create parent directory for %s: %v", path, err)
	}

	file, err := os.Create(path)
	if err != nil {
		fatalf("create %s: %v", path, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fatalf("write %s: %v", path, err)
	}
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}