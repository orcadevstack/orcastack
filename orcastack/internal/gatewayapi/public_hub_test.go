package gatewayapi

import "testing"

func TestPublicViewerRoleAndCapabilities(t *testing.T) {
	if got := publicViewerRole(sessionRecord{}, false); got != "viewer" {
		t.Fatalf("anonymous role: got %s", got)
	}
	developer := sessionRecord{User: AuthUser{Role: "developer", Permissions: []string{"deployments:write"}}}
	if got := publicViewerRole(developer, true); got != "developer" {
		t.Fatalf("developer role: got %s", got)
	}
	if !canAccessDeploymentDashboard(developer, true) {
		t.Fatal("deployment writer should access deployment dashboard")
	}
	viewer := sessionRecord{User: AuthUser{Role: "viewer", Permissions: []string{"repositories:read"}}}
	if canAccessDeploymentDashboard(viewer, true) {
		t.Fatal("viewer should not access deployment dashboard")
	}
	administrator := sessionRecord{User: AuthUser{Role: "platform-admin"}}
	if !canPublishCommunity(administrator) {
		t.Fatal("platform administrator should publish community posts")
	}
}
