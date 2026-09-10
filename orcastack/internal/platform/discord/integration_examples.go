package discord

// ExampleIntegrations demonstrates how to integrate Discord notifications into various parts of ORCASTACK

// Example 1: Integration in Pipeline Service
// Place this in your pipeline execution handler (e.g., in orcastack-ci-service)
/*
func (ps *PipelineService) ExecutePipeline(ctx context.Context, pipelineID string) error {
	// ... existing pipeline logic ...
	
	// Initialize Discord manager
	discordManager := discord.GetManager()
	
	// Notify pipeline start
	_ = discordManager.NotifyPipelineEvent(ctx, 
		"start", 
		pipelineID, 
		branch, 
		commit,
		"https://orcastack.example.com/pipelines/"+pipelineID,
	)
	
	// ... execute pipeline ...
	
	if err != nil {
		// Notify failure
		_ = discordManager.NotifyPipelineEvent(ctx,
			"failure",
			pipelineID,
			branch,
			commit,
			"https://orcastack.example.com/pipelines/"+pipelineID,
		)
		return err
	}
	
	// Notify success
	_ = discordManager.NotifyPipelineEvent(ctx,
		"success",
		pipelineID,
		branch,
		commit,
		"https://orcastack.example.com/pipelines/"+pipelineID,
	)
	
	return nil
}
*/

// Example 2: Integration in Deployment Service
// Place this in your deployment handler (e.g., in orcastack-cd-service)
/*
func (ds *DeploymentService) Deploy(ctx context.Context, service, env, version string) error {
	discordManager := discord.GetManager()
	
	// Notify deployment start
	_ = discordManager.NotifyDeploymentEvent(ctx,
		"start",
		env,
		service,
		version,
		"https://orcastack.example.com/deployments",
	)
	
	// ... perform deployment ...
	
	if err != nil {
		_ = discordManager.NotifyDeploymentEvent(ctx,
			"failure",
			env,
			service,
			version,
			"https://orcastack.example.com/deployments",
		)
		return err
	}
	
	_ = discordManager.NotifyDeploymentEvent(ctx,
		"success",
		env,
		service,
		version,
		"https://orcastack.example.com/deployments",
	)
	return nil
}
*/

// Example 3: Integration in Git Service
// Place this in your repository event handlers (e.g., in orcastack-git-service)
/*
func (gs *GitService) OnMerge(ctx context.Context, repoName, sourceBranch, targetBranch string) error {
	discordManager := discord.GetManager()
	
	_ = discordManager.NotifyRepositoryEvent(ctx,
		"merged",
		repoName,
		"Branch "+sourceBranch+" merged into "+targetBranch,
		"https://orcastack.example.com/repos/"+repoName,
	)
	
	return nil
}

func (gs *GitService) OnBranchCreated(ctx context.Context, repoName, branchName string) error {
	discordManager := discord.GetManager()
	
	_ = discordManager.NotifyRepositoryEvent(ctx,
		"branch_created",
		repoName,
		"New branch created: "+branchName,
		"https://orcastack.example.com/repos/"+repoName,
	)
	
	return nil
}
*/

// Example 4: Integration in Security/Auth Layer
// Place this in your auth handlers
/*
func (ah *AuthHandler) OnAuthFailure(ctx context.Context, username string) error {
	discordManager := discord.GetManager()
	
	_ = discordManager.NotifySecurityEvent(ctx,
		"auth_failure",
		"Failed authentication attempt",
		username,
		"Multiple failed login attempts detected",
		"https://orcastack.example.com/security/audit",
	)
	
	return nil
}

func (ah *AuthHandler) OnRBACViolation(ctx context.Context, username, resource string) error {
	discordManager := discord.GetManager()
	
	_ = discordManager.NotifySecurityEvent(ctx,
		"violation",
		"RBAC violation",
		username,
		"Attempted access to "+resource+" without permission",
		"https://orcastack.example.com/security/violations",
	)
	
	return nil
}
*/

// Example 5: Integration in Health Check Service
// Place this in your health monitoring (e.g., in orcastack-analytics-service)
/*
func (hc *HealthChecker) CheckSystemHealth(ctx context.Context) error {
	discordManager := discord.GetManager()
	
	cpuUsage := getCPUUsage()
	if cpuUsage > 80 {
		_ = discordManager.NotifyHealthEvent(ctx,
			"warning",
			"CPU",
			"usage_percent",
			fmt.Sprintf("%.2f", cpuUsage),
			"https://orcastack.example.com/health",
		)
	}
	
	memUsage := getMemoryUsage()
	if memUsage > 90 {
		_ = discordManager.NotifyHealthEvent(ctx,
			"critical",
			"Memory",
			"usage_percent",
			fmt.Sprintf("%.2f", memUsage),
			"https://orcastack.example.com/health",
		)
	}
	
	return nil
}
*/

// Example 6: Integration in Gateway/Signup
// Place this in your signup/access request handlers (e.g., in gatewayapi)
/*
func (gw *Gateway) OnUserSignup(ctx context.Context, username, email string) error {
	discordManager := discord.GetManager()
	
	_ = discordManager.NotifyUserEvent(ctx,
		"signup",
		username,
		email,
		"New user registered",
		"https://orcastack.example.com/users",
	)
	
	return nil
}

func (gw *Gateway) OnAccessRequest(ctx context.Context, username string, requestedResources []string) error {
	discordManager := discord.GetManager()
	
	resources := strings.Join(requestedResources, ", ")
	_ = discordManager.NotifyUserEvent(ctx,
		"access_request",
		username,
		"",
		"Access request for: "+resources,
		"https://orcastack.example.com/access-requests",
	)
	
	return nil
}
*/

// Example 7: Custom Event Notification
// For any custom event not covered by the standard types
/*
func SomeCustomHandler(ctx context.Context) error {
	discordManager := discord.GetManager()
	
	// Create custom event
	event := &discord.NotificationEvent{
		EventType:    discord.EventType("custom.event"),
		Title:        "Custom Event",
		Description:  "Something important happened",
		Severity:     discord.SeverityWarning,
		DashboardURL: "https://orcastack.example.com/custom",
		Timestamp:    time.Now(),
		Details: map[string]string{
			"Key1": "Value1",
			"Key2": "Value2",
		},
	}
	
	return discordManager.NotifyCustomEvent(ctx, event)
}
*/

// InitializeDiscordNotifications should be called in main.go or service initialization
// This sets up the global Discord notification manager
func InitializeDiscordNotifications() {
	Initialize()
	manager := GetManager()
	if manager.IsEnabled() {
		// Log success
	}
}
