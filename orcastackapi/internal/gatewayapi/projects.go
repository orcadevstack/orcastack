package gatewayapi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const repositoryMetadataFile = ".orcastack.json"

var repositoryNamePattern = regexp.MustCompile(`[^a-z0-9._-]+`)

type repositoryMetadata struct {
	Name          string `json:"name"`
	Summary       string `json:"summary"`
	DefaultBranch string `json:"default_branch"`
	ProviderID    string `json:"provider_id"`
	SourceURL     string `json:"source_url,omitempty"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	LastAction    string `json:"last_action"`
}

type createRepositoryRequest struct {
	Name          string `json:"name"`
	Summary       string `json:"summary"`
	DefaultBranch string `json:"default_branch"`
}

type importRepositoryRequest struct {
	Name          string `json:"name"`
	Summary       string `json:"summary"`
	DefaultBranch string `json:"default_branch"`
	SourceURL     string `json:"source_url"`
}

type repositoryMutationResponse struct {
	Repository     Repository     `json:"repository"`
	CloneOperation CloneOperation `json:"clone_operation"`
}

func repositoryRoot() string {
	if root := strings.TrimSpace(os.Getenv("ORCASTACK_REPO_ROOT")); root != "" {
		return root
	}

	return "/srv/git"
}

func repositoryPublicRoot() string {
	if root := strings.TrimSpace(os.Getenv("ORCASTACK_REPO_PUBLIC_ROOT")); root != "" {
		return root
	}

	return repositoryRoot()
}

func publicGitBase() string {
	if base := strings.TrimSpace(os.Getenv("ORCASTACK_PUBLIC_GIT_BASE")); base != "" {
		return strings.TrimRight(base, "/")
	}

	return "http://localhost:8080/git"
}

func ensureRepositoryRoot() error {
	return os.MkdirAll(repositoryRoot(), 0o755)
}

func repositoryDir(name string) string {
	return filepath.Join(repositoryRoot(), name+".git")
}

func repositoryMetadataPath(repoDir string) string {
	return filepath.Join(repoDir, repositoryMetadataFile)
}

func sanitizeRepositoryName(raw string) string {
	cleaned := strings.ToLower(strings.TrimSpace(raw))
	cleaned = strings.TrimSuffix(cleaned, ".git")
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = repositoryNamePattern.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-._")
	for strings.Contains(cleaned, "--") {
		cleaned = strings.ReplaceAll(cleaned, "--", "-")
	}
	return cleaned
}

func inferRepositoryName(sourceURL string) string {
	parsed, err := url.Parse(sourceURL)
	if err == nil && parsed.Path != "" {
		return sanitizeRepositoryName(path.Base(parsed.Path))
	}

	parts := strings.Split(strings.TrimSpace(sourceURL), "/")
	if len(parts) == 0 {
		return ""
	}

	return sanitizeRepositoryName(parts[len(parts)-1])
}

func providerForSource(sourceURL string) string {
	parsed, err := url.Parse(sourceURL)
	if err != nil {
		return "external"
	}

	host := strings.ToLower(parsed.Hostname())
	switch {
	case strings.Contains(host, "github"):
		return "github"
	case strings.Contains(host, "gitlab"):
		return "gitlab"
	case strings.Contains(host, "bitbucket"):
		return "bitbucket"
	case strings.Contains(host, "gitea"):
		return "gitea"
	case host == "":
		return "external"
	default:
		return "external"
	}
}

func repositoryCloneURL(name string) string {
	return publicGitBase() + "/" + name + ".git"
}

func runGit(args ...string) (string, error) {
	commandArgs := append([]string{"-c", "safe.directory=*"}, args...)
	cmd := exec.Command("git", commandArgs...)
	output, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(output))
	if err != nil {
		if trimmed == "" {
			return "", err
		}
		return "", fmt.Errorf("%w: %s", err, trimmed)
	}

	return trimmed, nil
}

func enableRepositoryHTTPAccess(repoDir string) error {
	if _, err := runGit("--git-dir", repoDir, "config", "http.receivepack", "true"); err != nil {
		return err
	}

	if _, err := runGit("--git-dir", repoDir, "config", "http.uploadpack", "true"); err != nil {
		return err
	}

	return nil
}

func writeRepositoryMetadata(repoDir string, metadata repositoryMetadata) error {
	payload, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(repositoryMetadataPath(repoDir), payload, 0o644)
}

func readRepositoryMetadata(repoDir string) (repositoryMetadata, error) {
	metaPath := repositoryMetadataPath(repoDir)
	payload, err := os.ReadFile(metaPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			name := strings.TrimSuffix(filepath.Base(repoDir), ".git")
			now := time.Now().UTC().Format(time.RFC3339)
			return repositoryMetadata{
				Name:          name,
				Summary:       "Local repository hosted by orcastack.",
				DefaultBranch: "main",
				ProviderID:    "orcastack",
				CreatedAt:     now,
				UpdatedAt:     now,
				LastAction:    "created",
			}, nil
		}
		return repositoryMetadata{}, err
	}

	var metadata repositoryMetadata
	if err := json.Unmarshal(payload, &metadata); err != nil {
		return repositoryMetadata{}, err
	}

	return metadata, nil
}

func inspectRepository(repoDir, fallbackBranch, fallbackUpdatedAt string) (string, string, string) {
	branch := fallbackBranch
	if output, err := runGit("--git-dir", repoDir, "symbolic-ref", "--short", "HEAD"); err == nil && output != "" {
		branch = output
	}

	commit := "No commits yet"
	if output, err := runGit("--git-dir", repoDir, "rev-parse", "--short", "HEAD"); err == nil && output != "" {
		commit = output
	}

	lastCommitAt := fallbackUpdatedAt
	if output, err := runGit("--git-dir", repoDir, "log", "-1", "--format=%cI"); err == nil && output != "" {
		lastCommitAt = output
	}

	return branch, commit, lastCommitAt
}

func buildRepository(metadata repositoryMetadata) Repository {
	repoDir := repositoryDir(metadata.Name)
	branch, commit, lastCommitAt := inspectRepository(repoDir, metadata.DefaultBranch, metadata.UpdatedAt)
	summary := strings.TrimSpace(metadata.Summary)
	if summary == "" {
		if metadata.SourceURL != "" {
			summary = "Imported from " + metadata.SourceURL
		} else {
			summary = "Local repository hosted by orcastack."
		}
	}

	return Repository{
		ID:            metadata.Name,
		ProviderID:    metadata.ProviderID,
		Name:          metadata.Name,
		Branch:        branch,
		DefaultBranch: metadata.DefaultBranch,
		Commit:        commit,
		LastCommitAt:  lastCommitAt,
		Reviewer:      "gateway",
		Summary:       summary,
		CloneURL:      repositoryCloneURL(metadata.Name),
		Identity:      "orcastack:repo:" + metadata.Name,
		Security: SecurityState{
			LDAPRegistered:    false,
			RBACVerified:      false,
			AttestationSigned: false,
			Verified:          false,
		},
	}
}

func buildCloneOperation(metadata repositoryMetadata, repository Repository) CloneOperation {
	status := "ready"
	if metadata.SourceURL != "" {
		status = "completed"
	}

	return CloneOperation{
		RepositoryID: repository.ID,
		Status:       status,
		CloneURL:     repository.CloneURL,
		Command:      "git clone " + repository.CloneURL,
		UPI:          "orcastack:clone:" + repository.ID,
		UpdatedAt:    metadata.UpdatedAt,
	}
}

func listRepositoryMetadata() ([]repositoryMetadata, error) {
	if err := ensureRepositoryRoot(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(repositoryRoot())
	if err != nil {
		return nil, err
	}

	repositories := make([]repositoryMetadata, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".git") {
			continue
		}

		metadata, err := readRepositoryMetadata(filepath.Join(repositoryRoot(), entry.Name()))
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, metadata)
	}

	sort.Slice(repositories, func(i, j int) bool {
		return repositories[i].Name < repositories[j].Name
	})

	return repositories, nil
}

func providerName(providerID string) string {
	switch providerID {
	case "orcastack":
		return "orcastack"
	case "github":
		return "GitHub"
	case "gitlab":
		return "GitLab"
	case "bitbucket":
		return "Bitbucket"
	case "gitea":
		return "Gitea"
	default:
		return "External Git"
	}
}

func overviewData() (Overview, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	metadataEntries, err := listRepositoryMetadata()
	if err != nil {
		return Overview{}, err
	}

	cloudLayers := []CloudLayer{}
	clusters := []Cluster{}
	automationLanes := []AutomationLane{}
	observability := []ObservabilitySurface{}
	selfManagement := []SelfManagementCapability{}

	repositories := make([]Repository, 0, len(metadataEntries))
	cloneOperations := make([]CloneOperation, 0, len(metadataEntries))
	events := make([]Event, 0, len(metadataEntries))
	providerCounts := map[string]int{}
	importedCount := 0
	for _, metadata := range metadataEntries {
		repository := buildRepository(metadata)
		repositories = append(repositories, repository)
		cloneOperations = append(cloneOperations, buildCloneOperation(metadata, repository))
		providerCounts[metadata.ProviderID]++
		if metadata.ProviderID != "orcastack" {
			importedCount++
		}
		events = append(events, Event{
			ID:           "evt-" + metadata.Name,
			Time:         metadata.UpdatedAt,
			Component:    metadata.Name,
			Kind:         "repository",
			RepositoryID: metadata.Name,
			Action:       metadata.LastAction,
			Result:       "success",
			UPI:          "orcastack:event:" + metadata.Name,
			Summary:      buildRepository(metadata).Summary,
		})
	}

	providerIDs := make([]string, 0, len(providerCounts))
	for providerID := range providerCounts {
		providerIDs = append(providerIDs, providerID)
	}
	sort.Strings(providerIDs)

	providers := make([]Provider, 0, len(providerIDs))
	for _, providerID := range providerIDs {
		status := "connected"
		connected := true
		if providerID == "orcastack" {
			status = "primary"
		}
		providers = append(providers, Provider{
			ID:        providerID,
			Name:      providerName(providerID),
			Status:    status,
			Repos:     providerCounts[providerID],
			Latency:   "local",
			Identity:  "orcastack:provider:" + providerID,
			Connected: connected,
		})
	}

	pipelines := []Pipeline{}
	deployments := []Deployment{}
	containers := []Container{}

	metrics := []Metric{
		{Label: "Projects tracked", Value: fmt.Sprintf("%d", len(repositories)), Hint: "Repositories available for clone and push through the local control plane."},
		{Label: "Hosted locally", Value: fmt.Sprintf("%d", providerCounts["orcastack"]), Hint: "Bare repositories created directly inside the orcastack repo store."},
		{Label: "Imported remotes", Value: fmt.Sprintf("%d", importedCount), Hint: "Repositories mirrored from external Git sources."},
		{Label: "Push targets ready", Value: fmt.Sprintf("%d", len(repositories)), Hint: "Each project exposes an HTTP Git remote you can clone from and push to over the network."},
	}

	activity := []string{}

	overview := Overview{
		Providers:       providers,
		Repositories:    repositories,
		CloneOperations: cloneOperations,
		Reviews:         []Review{},
		Pipelines:       pipelines,
		Deployments:     deployments,
		Containers:      containers,
		Security: DashboardSecurity{
			RepositoryIdentity: "orcastack:directory:repositories",
			UIProcessIdentity:  "orcastack:ui:control-plane",
			Directory:          SecurityState{},
		},
		Events:          events,
		UpdatedAt:       now,
		Metrics:         metrics,
		Activity:        activity,
		CloudLayers:     cloudLayers,
		Clusters:        clusters,
		AutomationLanes: automationLanes,
		Observability:   observability,
		SelfManagement:  selfManagement,
	}

	enrichOverviewFromLiveAPIs(&overview, now)

	return overview, nil
}

func createRepository(req createRepositoryRequest) (repositoryMutationResponse, error) {
	name := sanitizeRepositoryName(req.Name)
	if name == "" {
		return repositoryMutationResponse{}, errors.New("repository name is required")
	}

	defaultBranch := strings.TrimSpace(req.DefaultBranch)
	if defaultBranch == "" {
		defaultBranch = "main"
	}

	if err := ensureRepositoryRoot(); err != nil {
		return repositoryMutationResponse{}, err
	}

	repoDir := repositoryDir(name)
	if _, err := os.Stat(repoDir); err == nil {
		return repositoryMutationResponse{}, fmt.Errorf("repository %s already exists", name)
	}

	if _, err := runGit("init", "--bare", "--initial-branch", defaultBranch, repoDir); err != nil {
		return repositoryMutationResponse{}, err
	}
	if err := enableRepositoryHTTPAccess(repoDir); err != nil {
		return repositoryMutationResponse{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	metadata := repositoryMetadata{
		Name:          name,
		Summary:       strings.TrimSpace(req.Summary),
		DefaultBranch: defaultBranch,
		ProviderID:    "orcastack",
		CreatedAt:     now,
		UpdatedAt:     now,
		LastAction:    "created",
	}
	if err := writeRepositoryMetadata(repoDir, metadata); err != nil {
		return repositoryMutationResponse{}, err
	}

	repository := buildRepository(metadata)
	return repositoryMutationResponse{
		Repository:     repository,
		CloneOperation: buildCloneOperation(metadata, repository),
	}, nil
}

func importRepository(req importRepositoryRequest) (repositoryMutationResponse, error) {
	sourceURL := strings.TrimSpace(req.SourceURL)
	if sourceURL == "" {
		return repositoryMutationResponse{}, errors.New("source repository URL is required")
	}

	name := sanitizeRepositoryName(req.Name)
	if name == "" {
		name = inferRepositoryName(sourceURL)
	}
	if name == "" {
		return repositoryMutationResponse{}, errors.New("repository name is required")
	}

	if err := ensureRepositoryRoot(); err != nil {
		return repositoryMutationResponse{}, err
	}

	repoDir := repositoryDir(name)
	if _, err := os.Stat(repoDir); err == nil {
		return repositoryMutationResponse{}, fmt.Errorf("repository %s already exists", name)
	}

	if _, err := runGit("clone", "--mirror", sourceURL, repoDir); err != nil {
		return repositoryMutationResponse{}, err
	}
	if err := enableRepositoryHTTPAccess(repoDir); err != nil {
		return repositoryMutationResponse{}, err
	}

	defaultBranch := strings.TrimSpace(req.DefaultBranch)
	if defaultBranch == "" {
		if branch, _, _ := inspectRepository(repoDir, "main", time.Now().UTC().Format(time.RFC3339)); branch != "" {
			defaultBranch = branch
		} else {
			defaultBranch = "main"
		}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	metadata := repositoryMetadata{
		Name:          name,
		Summary:       strings.TrimSpace(req.Summary),
		DefaultBranch: defaultBranch,
		ProviderID:    providerForSource(sourceURL),
		SourceURL:     sourceURL,
		CreatedAt:     now,
		UpdatedAt:     now,
		LastAction:    "imported",
	}
	if err := writeRepositoryMetadata(repoDir, metadata); err != nil {
		return repositoryMutationResponse{}, err
	}

	repository := buildRepository(metadata)
	return repositoryMutationResponse{
		Repository:     repository,
		CloneOperation: buildCloneOperation(metadata, repository),
	}, nil
}

func decodeJSON(r *http.Request, target any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func writeError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}

func serveGitHTTP(w http.ResponseWriter, r *http.Request) {
	pathInfo := strings.TrimPrefix(r.URL.Path, "/git")
	pathInfo = path.Clean("/" + strings.TrimPrefix(pathInfo, "/"))
	if pathInfo == "/" || strings.Contains(pathInfo, "..") {
		http.NotFound(w, r)
		return
	}

	var requestBody bytes.Buffer
	if r.Body != nil {
		_, _ = io.Copy(&requestBody, r.Body)
		_ = r.Body.Close()
	}

	cmd := exec.Command("git", "http-backend")
	cmd.Env = append(os.Environ(),
		"GIT_PROJECT_ROOT="+repositoryRoot(),
		"GIT_HTTP_EXPORT_ALL=1",
		"PATH_INFO="+pathInfo,
		"REQUEST_METHOD="+r.Method,
		"QUERY_STRING="+r.URL.RawQuery,
		"CONTENT_TYPE="+r.Header.Get("Content-Type"),
		"CONTENT_LENGTH="+strconv.Itoa(requestBody.Len()),
		"REMOTE_ADDR="+r.RemoteAddr,
	)
	cmd.Stdin = bytes.NewReader(requestBody.Bytes())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil && stdout.Len() == 0 {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		writeError(w, http.StatusInternalServerError, errors.New(message))
		return
	}

	if err := writeCGIResponse(w, stdout.Bytes()); err != nil {
		writeError(w, http.StatusInternalServerError, err)
	}
}

func writeCGIResponse(w http.ResponseWriter, payload []byte) error {
	reader := bufio.NewReader(bytes.NewReader(payload))
	textReader := textproto.NewReader(reader)
	headers, err := textReader.ReadMIMEHeader()
	if err != nil {
		return err
	}

	statusCode := http.StatusOK
	if status := headers.Get("Status"); status != "" {
		parts := strings.Fields(status)
		if len(parts) > 0 {
			if code, convErr := strconv.Atoi(parts[0]); convErr == nil {
				statusCode = code
			}
		}
	}

	for key, values := range headers {
		if strings.EqualFold(key, "Status") {
			continue
		}
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	w.WriteHeader(statusCode)
	_, err = io.Copy(w, reader)
	return err
}
