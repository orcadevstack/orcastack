package gatewayapi

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type rancherClusterListResponse struct {
	Data []rancherCluster `json:"data"`
}

type rancherCluster struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	State      string         `json:"state"`
	Transition string         `json:"transitioning"`
	Driver     string         `json:"driver"`
	APIEndpoint string        `json:"apiEndpoint"`
	Version    rancherVersion `json:"version"`
}

type rancherVersion struct {
	GitVersion string `json:"gitVersion"`
}

type rancherNodeListResponse struct {
	Data []rancherNode `json:"data"`
}

type rancherNode struct {
	ClusterID    string            `json:"clusterId"`
	ControlPlane bool              `json:"controlPlane"`
	Worker       bool              `json:"worker"`
	Labels       map[string]string `json:"labels"`
}

type rancherProjectListResponse struct {
	Data []rancherProject `json:"data"`
}

type rancherProject struct {
	ClusterID string `json:"clusterId"`
	Name      string `json:"name"`
}

type openStackAuthResponse struct {
	Token openStackToken `json:"token"`
}

type openStackToken struct {
	Project openStackProject          `json:"project"`
	Catalog []openStackCatalogService `json:"catalog"`
}

type openStackProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type openStackCatalogService struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Endpoints []struct {
		Interface string `json:"interface"`
		Region    string `json:"region"`
		URL       string `json:"url"`
	} `json:"endpoints"`
}

type openStackProjectListResponse struct {
	Projects []openStackProject `json:"projects"`
}

func enrichOverviewFromLiveAPIs(overview *Overview, now string) {
	if snapshot, err := loadRancherSnapshot(); err == nil {
		applyRancherSnapshot(overview, snapshot, now)
	} else if err != errCloudIntegrationNotConfigured {
		markCloudLayer(overview, "Cluster management", "degraded", "Rancher API call failed: "+err.Error())
		overview.Events = append(overview.Events, Event{
			ID:        "evt-rancher-live-error",
			Time:      now,
			Component: "rancher-api",
			Kind:      "process",
			Action:    "query",
			Result:    "degraded",
			UPI:       "orcastack:event:rancher-live",
			Summary:   "Rancher live overview refresh failed: " + err.Error(),
		})
	}

	if snapshot, err := loadOpenStackSnapshot(); err == nil {
		applyOpenStackSnapshot(overview, snapshot, now)
	} else if err != errCloudIntegrationNotConfigured {
		markCloudLayer(overview, "Private IaaS", "degraded", "OpenStack API call failed: "+err.Error())
		overview.Events = append(overview.Events, Event{
			ID:        "evt-openstack-live-error",
			Time:      now,
			Component: "openstack-api",
			Kind:      "process",
			Action:    "query",
			Result:    "degraded",
			UPI:       "orcastack:event:openstack-live",
			Summary:   "OpenStack live overview refresh failed: " + err.Error(),
		})
	}

	setMetricValue(overview, "Clusters under Rancher", fmt.Sprintf("%d", len(overview.Clusters)))
}

var errCloudIntegrationNotConfigured = fmt.Errorf("cloud integration not configured")

type rancherSnapshot struct {
	serverURL string
	clusters  []Cluster
}

func loadRancherSnapshot() (rancherSnapshot, error) {
	serverURL := firstNonEmptyEnv("ORCASTACK_RANCHER_URL", "RANCHER_URL")
	token := firstNonEmptyEnv("ORCASTACK_RANCHER_BEARER_TOKEN", "RANCHER_BEARER_TOKEN")
	if serverURL == "" || token == "" {
		return rancherSnapshot{}, errCloudIntegrationNotConfigured
	}

	client := cloudHTTPClient()
	var clusterResponse rancherClusterListResponse
	if err := doCloudJSONRequest(client, http.MethodGet, strings.TrimRight(serverURL, "/")+"/v3/clusters", token, "", &clusterResponse); err != nil {
		return rancherSnapshot{}, err
	}

	var nodeResponse rancherNodeListResponse
	if err := doCloudJSONRequest(client, http.MethodGet, strings.TrimRight(serverURL, "/")+"/v3/nodes?limit=-1", token, "", &nodeResponse); err != nil {
		return rancherSnapshot{}, err
	}

	var projectResponse rancherProjectListResponse
	if err := doCloudJSONRequest(client, http.MethodGet, strings.TrimRight(serverURL, "/")+"/v3/projects?limit=-1", token, "", &projectResponse); err != nil {
		return rancherSnapshot{}, err
	}

	projectByCluster := map[string]string{}
	for _, project := range projectResponse.Data {
		if _, exists := projectByCluster[project.ClusterID]; !exists && strings.TrimSpace(project.Name) != "" {
			projectByCluster[project.ClusterID] = project.Name
		}
	}

	nodeCounts := map[string]struct{ controlPlanes, workers, gpuWorkers int }{}
	for _, node := range nodeResponse.Data {
		counts := nodeCounts[node.ClusterID]
		if node.ControlPlane {
			counts.controlPlanes++
		}
		if node.Worker {
			counts.workers++
		}
		if value, ok := node.Labels["nvidia.com/gpu.present"]; ok && strings.EqualFold(value, "true") {
			counts.gpuWorkers++
		}
		nodeCounts[node.ClusterID] = counts
	}

	clusters := make([]Cluster, 0, len(clusterResponse.Data))
	for _, cluster := range clusterResponse.Data {
		counts := nodeCounts[cluster.ID]
		clusters = append(clusters, Cluster{
			ID:                 cluster.ID,
			Name:               cluster.Name,
			Provider:           rancherProviderLabel(cluster.Driver),
			Status:             rancherStateToStatus(cluster.State, cluster.Transition),
			Version:            cluster.Version.GitVersion,
			ControlPlanes:      counts.controlPlanes,
			Workers:            counts.workers,
			GPUWorkers:         counts.gpuWorkers,
			RancherProject:     projectByCluster[cluster.ID],
			RegistrationStatus: rancherRegistrationStatus(cluster.State, cluster.Transition),
			UpgradePolicy:      "orcastack-governed",
			APIEndpoint:        cluster.APIEndpoint,
		})
	}

	return rancherSnapshot{serverURL: serverURL, clusters: clusters}, nil
}

type openStackSnapshot struct {
	authURL        string
	projectName    string
	serviceTypes   []string
	serviceCatalog []openStackCatalogService
	projectCount   int
}

func loadOpenStackSnapshot() (openStackSnapshot, error) {
	authURL := firstNonEmptyEnv("ORCASTACK_OPENSTACK_AUTH_URL", "OS_AUTH_URL")
	if authURL == "" {
		return openStackSnapshot{}, errCloudIntegrationNotConfigured
	}

	authBody, err := buildOpenStackAuthBody()
	if err != nil {
		return openStackSnapshot{}, err
	}

	client := cloudHTTPClient()
	requestURL := strings.TrimRight(authURL, "/") + "/auth/tokens"
	request, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(authBody))
	if err != nil {
		return openStackSnapshot{}, err
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return openStackSnapshot{}, err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return openStackSnapshot{}, fmt.Errorf("unexpected status %d from Keystone", response.StatusCode)
	}

	token := strings.TrimSpace(response.Header.Get("X-Subject-Token"))
	if token == "" {
		return openStackSnapshot{}, fmt.Errorf("Keystone token missing from response")
	}

	var authResponse openStackAuthResponse
	if err := json.NewDecoder(response.Body).Decode(&authResponse); err != nil {
		return openStackSnapshot{}, err
	}

	serviceTypes := make([]string, 0, len(authResponse.Token.Catalog))
	for _, service := range authResponse.Token.Catalog {
		serviceTypes = append(serviceTypes, service.Type)
	}

	projectCount := 0
	identityURL := openStackEndpoint(authResponse.Token.Catalog, "identity")
	if identityURL != "" {
		var projectResponse openStackProjectListResponse
		if err := doCloudJSONRequest(client, http.MethodGet, strings.TrimRight(identityURL, "/")+"/auth/projects", token, "", &projectResponse); err == nil {
			projectCount = len(projectResponse.Projects)
		}
	}

	return openStackSnapshot{
		authURL:        authURL,
		projectName:    authResponse.Token.Project.Name,
		serviceTypes:   serviceTypes,
		serviceCatalog: authResponse.Token.Catalog,
		projectCount:   projectCount,
	}, nil
}

func buildOpenStackAuthBody() (string, error) {
	appCredentialID := firstNonEmptyEnv("ORCASTACK_OPENSTACK_APP_CRED_ID", "OS_APPLICATION_CREDENTIAL_ID")
	appCredentialSecret := firstNonEmptyEnv("ORCASTACK_OPENSTACK_APP_CRED_SECRET", "OS_APPLICATION_CREDENTIAL_SECRET")
	if appCredentialID != "" && appCredentialSecret != "" {
		return fmt.Sprintf(`{"auth":{"identity":{"methods":["application_credential"],"application_credential":{"id":%q,"secret":%q}}}}`, appCredentialID, appCredentialSecret), nil
	}

	username := firstNonEmptyEnv("ORCASTACK_OPENSTACK_USERNAME", "OS_USERNAME")
	password := firstNonEmptyEnv("ORCASTACK_OPENSTACK_PASSWORD", "OS_PASSWORD")
	userDomain := firstNonEmptyEnv("ORCASTACK_OPENSTACK_USER_DOMAIN", "OS_USER_DOMAIN_NAME", "Default")
	projectName := firstNonEmptyEnv("ORCASTACK_OPENSTACK_PROJECT", "OS_PROJECT_NAME")
	projectDomain := firstNonEmptyEnv("ORCASTACK_OPENSTACK_PROJECT_DOMAIN", "OS_PROJECT_DOMAIN_NAME", "Default")
	if username != "" && password != "" && projectName != "" {
		return fmt.Sprintf(`{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":%q,"password":%q,"domain":{"name":%q}}}},"scope":{"project":{"name":%q,"domain":{"name":%q}}}}}`, username, password, userDomain, projectName, projectDomain), nil
	}

	return "", fmt.Errorf("OpenStack credentials are not configured")
}

func applyRancherSnapshot(overview *Overview, snapshot rancherSnapshot, now string) {
	markCloudLayer(overview, "Cluster management", "online", fmt.Sprintf("Rancher live API connected at %s with %d discovered clusters.", snapshot.serverURL, len(snapshot.clusters)))
	for index := range overview.CloudLayers {
		if overview.CloudLayers[index].Name == "Cluster management" {
			overview.CloudLayers[index].Endpoint = snapshot.serverURL
			break
		}
	}

	if len(snapshot.clusters) > 0 {
		overview.Clusters = mergeClusters(overview.Clusters, snapshot.clusters)
	}

	overview.Events = append(overview.Events, Event{
		ID:        "evt-rancher-live-refresh",
		Time:      now,
		Component: "rancher-api",
		Kind:      "process",
		Action:    "query",
		Result:    "success",
		UPI:       "orcastack:event:rancher-live",
		Summary:   fmt.Sprintf("Rancher live overview refresh succeeded for %d clusters.", len(snapshot.clusters)),
	})
}

func applyOpenStackSnapshot(overview *Overview, snapshot openStackSnapshot, now string) {
	serviceSummary := strings.Join(snapshot.serviceTypes, ", ")
	if serviceSummary == "" {
		serviceSummary = "catalog unavailable"
	}

	markCloudLayer(overview, "Private IaaS", "online", fmt.Sprintf("Keystone authenticated for project %s with services: %s.", fallbackString(snapshot.projectName, "unknown"), serviceSummary))
	for index := range overview.CloudLayers {
		if overview.CloudLayers[index].Name == "Private IaaS" {
			overview.CloudLayers[index].Endpoint = snapshot.authURL
			break
		}
	}

	setMetricValue(overview, "Cloud layers managed", fmt.Sprintf("%d", len(overview.CloudLayers)))
	overview.Events = append(overview.Events, Event{
		ID:        "evt-openstack-live-refresh",
		Time:      now,
		Component: "openstack-api",
		Kind:      "process",
		Action:    "query",
		Result:    "success",
		UPI:       "orcastack:event:openstack-live",
		Summary:   fmt.Sprintf("OpenStack live overview refresh succeeded for project %s with %d visible projects.", fallbackString(snapshot.projectName, "unknown"), snapshot.projectCount),
	})
}

func mergeClusters(seed []Cluster, live []Cluster) []Cluster {
	seedByKey := map[string]Cluster{}
	for _, cluster := range seed {
		seedByKey[cluster.ID] = cluster
		seedByKey[strings.ToLower(cluster.Name)] = cluster
	}

	merged := make([]Cluster, 0, len(live))
	for _, cluster := range live {
		candidate, ok := seedByKey[cluster.ID]
		if !ok {
			candidate, ok = seedByKey[strings.ToLower(cluster.Name)]
		}
		if ok {
			if cluster.ControlPlanes == 0 {
				cluster.ControlPlanes = candidate.ControlPlanes
			}
			if cluster.Workers == 0 {
				cluster.Workers = candidate.Workers
			}
			if cluster.GPUWorkers == 0 {
				cluster.GPUWorkers = candidate.GPUWorkers
			}
			if cluster.RancherProject == "" {
				cluster.RancherProject = candidate.RancherProject
			}
			if cluster.UpgradePolicy == "" {
				cluster.UpgradePolicy = candidate.UpgradePolicy
			}
			if cluster.APIEndpoint == "" {
				cluster.APIEndpoint = candidate.APIEndpoint
			}
		}
		merged = append(merged, cluster)
	}

	return merged
}

func markCloudLayer(overview *Overview, layerName, status, summary string) {
	for index := range overview.CloudLayers {
		if overview.CloudLayers[index].Name == layerName {
			overview.CloudLayers[index].Status = status
			overview.CloudLayers[index].Summary = summary
			return
		}
	}
}

func setMetricValue(overview *Overview, label, value string) {
	for index := range overview.Metrics {
		if overview.Metrics[index].Label == label {
			overview.Metrics[index].Value = value
			return
		}
	}
}

func rancherProviderLabel(driver string) string {
	driver = strings.TrimSpace(driver)
	if driver == "" {
		return "Rancher"
	}

	return "Rancher / " + strings.ToUpper(driver)
}

func rancherStateToStatus(state, transition string) string {
	if strings.TrimSpace(transition) != "" && !strings.EqualFold(transition, "no") {
		return "syncing"
	}

	switch strings.ToLower(strings.TrimSpace(state)) {
	case "active":
		return "online"
	case "pending", "provisioning", "updating-active":
		return "syncing"
	case "error", "failed", "unavailable":
		return "degraded"
	default:
		if state == "" {
			return "unknown"
		}
		return strings.ToLower(state)
	}
}

func rancherRegistrationStatus(state, transition string) string {
	if strings.TrimSpace(transition) != "" && !strings.EqualFold(transition, "no") {
		return "reconciling"
	}

	if strings.EqualFold(state, "active") {
		return "registered"
	}

	if strings.TrimSpace(state) == "" {
		return "unknown"
	}

	return strings.ToLower(state)
}

func openStackEndpoint(catalog []openStackCatalogService, serviceType string) string {
	for _, service := range catalog {
		if service.Type != serviceType {
			continue
		}
		for _, endpoint := range service.Endpoints {
			if endpoint.Interface == "public" || endpoint.Interface == "internal" {
				return endpoint.URL
			}
		}
	}

	return ""
}

func doCloudJSONRequest(client *http.Client, method, url, bearerToken string, body string, target any) error {
	var requestBody *strings.Reader
	if body != "" {
		requestBody = strings.NewReader(body)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if bearerToken != "" {
		request.Header.Set("Authorization", "Bearer "+bearerToken)
	}

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("unexpected status %d from %s", response.StatusCode, url)
	}

	if target == nil {
		return nil
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func cloudHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if strings.EqualFold(firstNonEmptyEnv("ORCASTACK_CLOUD_INSECURE_TLS", "ORCASTACK_INSECURE_TLS"), "true") {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}

	return ""
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}

	return value
}