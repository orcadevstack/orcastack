package gatewayapi

import "testing"

func TestSessionHasPermission(t *testing.T) {
	session := sessionRecord{User: AuthUser{Permissions: []string{"repositories:read", "control-panel:read"}}}
	if !sessionHasPermission(session, "repositories:read") {
		t.Fatal("expected repository permission")
	}
	if sessionHasPermission(session, "control-panel:admin") {
		t.Fatal("operator session must not receive administrator permission")
	}

	administrator := sessionRecord{User: AuthUser{Permissions: []string{"control-panel:admin"}}}
	if !sessionHasPermission(administrator, "control-panel:read") {
		t.Fatal("administrator permission must inherit control-panel read access")
	}
}

func TestValidateHeaderAudit(t *testing.T) {
	valid := headerAuditRequest{Action: "navigate", Section: "projects", TargetPath: "/app/repositories"}
	if err := validateHeaderAudit(valid); err != nil {
		t.Fatalf("expected valid audit event: %v", err)
	}

	invalid := []headerAuditRequest{
		{Action: "delete", Section: "projects"},
		{Action: "open", Section: "billing"},
		{Action: "navigate", Section: "settings", TargetPath: "https://example.com"},
	}
	for _, request := range invalid {
		if err := validateHeaderAudit(request); err == nil {
			t.Fatalf("expected invalid audit event: %#v", request)
		}
	}
}

func TestProfileSectionSupportsAuditedActions(t *testing.T) {
	requests := []headerAuditRequest{
		{Action: "open", Section: "profile"},
		{Action: "navigate", Section: "profile", TargetPath: "/app/settings"},
	}
	for _, request := range requests {
		if err := validateHeaderAudit(request); err != nil {
			t.Fatalf("expected valid profile audit event: %v", err)
		}
	}
}

func TestDuplicateProfileSectionsAreNotHeaderSections(t *testing.T) {
	for _, section := range []string{"accounts", "settings"} {
		request := headerAuditRequest{Action: "open", Section: section}
		if err := validateHeaderAudit(request); err == nil {
			t.Fatalf("expected removed header section %s to be rejected", section)
		}
	}
}
