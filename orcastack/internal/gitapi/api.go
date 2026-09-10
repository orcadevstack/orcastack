package gitapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	platformconfig "github.com/orcastack/orcastackapi/internal/platform/config"
)

type verifyTokenRequest struct {
	Token   string `json:"token"`
	Command string `json:"command"`
}

type tokenClaims struct {
	Issuer         string `json:"iss"`
	Subject        string `json:"sub"`
	Repo           string `json:"repo"`
	Action         string `json:"action"`
	Source         string `json:"source"`
	KeyFingerprint string `json:"key_fingerprint"`
	Command        string `json:"command"`
	ExpiresAt      int64  `json:"exp"`
}

func Register(mux *http.ServeMux) {
	mux.HandleFunc("/repos", handleRepos)
	mux.HandleFunc("/auth/verify", handleVerifyToken)
}

func handleRepos(w http.ResponseWriter, _ *http.Request) {
	repoRoot := repoRoot()
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"repos": []string{}, "repo_root": repoRoot})
		return
	}
	repos := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			repos = append(repos, entry.Name())
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"repos": repos, "repo_root": repoRoot})
}

func handleVerifyToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	defer r.Body.Close()

	var request verifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	claims, err := verifyToken(request.Token)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": err.Error()})
		return
	}

	if request.Command != "" {
		action, err := commandAction(request.Command)
		if err != nil {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
			return
		}
		if action != claims.Action {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "token action does not match SSH command"})
			return
		}
		if claims.Command != "" && normalizeCommand(request.Command) != normalizeCommand(claims.Command) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "token command does not match requested SSH command"})
			return
		}
	}

	repoPath := filepath.Join(repoRoot(), claims.Repo+".git")
	writeJSON(w, http.StatusOK, map[string]any{
		"allowed":         true,
		"repo":            claims.Repo,
		"repo_path":       repoPath,
		"action":          claims.Action,
		"subject":         claims.Subject,
		"source":          claims.Source,
		"key_fingerprint": claims.KeyFingerprint,
	})
}

func verifyToken(token string) (tokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return tokenClaims{}, errors.New("invalid token format")
	}

	signature := computeSignature(parts[0])
	if !hmac.Equal([]byte(signature), []byte(parts[1])) {
		return tokenClaims{}, errors.New("invalid token signature")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return tokenClaims{}, errors.New("invalid token payload")
	}

	var claims tokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return tokenClaims{}, errors.New("invalid token claims")
	}
	if claims.ExpiresAt <= time.Now().UTC().Unix() {
		return tokenClaims{}, errors.New("token expired")
	}
	if claims.Repo == "" || claims.Action == "" || claims.Subject == "" {
		return tokenClaims{}, errors.New("token missing required claims")
	}
	return claims, nil
}

func commandAction(command string) (string, error) {
	normalized := normalizeCommand(command)
	switch {
	case strings.HasPrefix(normalized, "git-upload-pack "):
		return "read", nil
	case strings.HasPrefix(normalized, "git-upload-archive "):
		return "archive", nil
	case strings.HasPrefix(normalized, "git-receive-pack "):
		return "write", nil
	default:
		return "", fmt.Errorf("unsupported SSH command")
	}
}

func normalizeCommand(command string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(command)), " ")
}

func computeSignature(payload string) string {
	mac := hmac.New(sha256.New, []byte(platformconfig.String("ORCASTACK_GIT_TOKEN_SECRET", "orcastack-dev-secret")))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func repoRoot() string {
	return platformconfig.String("ORCASTACK_REPO_ROOT", "/srv/git")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}