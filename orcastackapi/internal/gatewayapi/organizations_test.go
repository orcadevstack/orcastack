package gatewayapi

import "testing"

func TestOrganizationRoleAllows(t *testing.T) {
	tests := []struct {
		role, required string
		allowed        bool
	}{
		{"owner", "owner", true},
		{"owner", "maintainer", true},
		{"maintainer", "maintainer", true},
		{"maintainer", "owner", false},
		{"developer", "maintainer", false},
	}
	for _, test := range tests {
		if got := organizationRoleAllows(test.role, test.required); got != test.allowed {
			t.Fatalf("role %s required %s: got %v", test.role, test.required, got)
		}
	}
}

func TestOrganizationInputValidation(t *testing.T) {
	if got := normalizeOrganizationSlug(" Platform Engineering "); got != "platform-engineering" {
		t.Fatalf("unexpected slug: %s", got)
	}
	for _, role := range []string{"owner", "maintainer", "developer"} {
		if err := validateOrganizationRole(role); err != nil {
			t.Fatalf("valid role %s: %v", role, err)
		}
	}
	if err := validateOrganizationRole("administrator"); err == nil {
		t.Fatal("expected invalid organization role")
	}
	for _, strategy := range []string{"feature", "release", "hotfix"} {
		if err := validateBranchStrategy(strategy); err != nil {
			t.Fatalf("valid strategy %s: %v", strategy, err)
		}
	}
	if err := validateBranchStrategy("trunk"); err == nil {
		t.Fatal("expected invalid branch strategy")
	}
}
