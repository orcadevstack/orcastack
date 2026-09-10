package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"help"}
	}

	switch args[0] {
	case "help", "--help", "-h":
		printHelp()

	case "version", "--version", "-v":
		fmt.Printf("orcastack %s\ncommit: %s\nbuilt: %s\n", version, commit, date)

	case "serve":
		if err := runGateway(args[1:]); err != nil {
			fmt.Fprintf(os.Stderr, "orcastack serve: %v\n", err)
			os.Exit(1)
		}

	case "serve-all":
		if err := runAllServices(); err != nil {
			fmt.Fprintf(os.Stderr, "orcastack serve-all: %v\n", err)
			os.Exit(1)
		}
		select {} // keep running

	case "healthcheck":
		if err := healthcheck(); err != nil {
			fmt.Fprintf(os.Stderr, "orcastack healthcheck: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", args[0])
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("ORCASTACK launcher")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  orcastack serve                     Start only the gateway service")
	fmt.Println("  orcastack serve-all                 Start ALL microservices")
	fmt.Println("  orcastack healthcheck               Check the configured gateway health endpoint")
	fmt.Println("  orcastack version                   Print build version")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  ORCASTACK_GATEWAY_BINARY            Override the packaged orcastack-gateway path")
	fmt.Println("  ORCASTACK_GATEWAY_BASE              Override the healthcheck base URL (default http://127.0.0.1:8080)")
}

func runAllServices() error {
	services := []string{
		"orcastack-analytics-service",
		"orcastack-device-orch",
		"orcastack-hw-automation",
		"orcastack-secctl",
		"orcastack-cd-service",
		"orcastack-gateway",
		"orcastack-review-service",
		"orcastack-sw-automation",
		"orcastack-ci-service",
		"orcastack-git-service",
		"orcastack-runner",
	}

	for _, svc := range services {
		go func(name string) {
			fmt.Printf("Starting %s...\n", name)
			cmd := exec.Command(name)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			cmd.Env = os.Environ()
			if err := cmd.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to start %s: %v\n", name, err)
			}
		}(svc)
	}

	return nil
}

func runGateway(args []string) error {
	binaryPath, err := locateGatewayBinary()
	if err != nil {
		return err
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Env = os.Environ()

	return cmd.Run()
}

func locateGatewayBinary() (string, error) {
	candidates := make([]string, 0, 6)
	if override := os.Getenv("ORCASTACK_GATEWAY_BINARY"); override != "" {
		candidates = append(candidates, override)
	}

	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "orcastack-gateway"),
			filepath.Join(exeDir, "..", "lib", "orcastack", "orcastack-gateway"),
		)
	}

	candidates = append(candidates,
		"/usr/lib/orcastack/orcastack-gateway",
		"/usr/local/lib/orcastack/orcastack-gateway",
		"orcastack-gateway",
	)

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		resolved := candidate
		if !filepath.IsAbs(candidate) {
			pathCandidate, err := exec.LookPath(candidate)
			if err != nil {
				continue
			}
			resolved = pathCandidate
		}

		info, err := os.Stat(resolved)
		if err == nil && !info.IsDir() {
			if runtime.GOOS == "windows" || info.Mode()&0o111 != 0 {
				return resolved, nil
			}
		}
	}

	return "", errors.New("packaged orcastack-gateway binary not found")
}

func healthcheck() error {
	base := os.Getenv("ORCASTACK_GATEWAY_BASE")
	if base == "" {
		base = "http://127.0.0.1:8080"
	}

	base = strings.TrimRight(base, "/")
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(base + "/healthz")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("unexpected status %s", resp.Status)
	}

	fmt.Println("gateway healthy")
	return nil
}
