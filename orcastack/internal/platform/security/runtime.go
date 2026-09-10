package security

import (
	"fmt"
	"os"
)

type LDAPConfig struct {
	Address                  string `json:"address"`
	ServiceAccountDN         string `json:"service_account_dn"`
	ServiceAccountPasswordFile string `json:"service_account_password_file"`
	ComponentBaseDN          string `json:"component_base_dn"`
	RepositoryBaseDN         string `json:"repository_base_dn"`
	AuditBaseDN              string `json:"audit_base_dn"`
	AutoRegister             bool   `json:"auto_register"`
}

type RBACConfig struct {
	Realm        string `json:"realm"`
	RoleBaseDN   string `json:"role_base_dn"`
	RequiredRole string `json:"required_role"`
}

type RuntimePolicy struct {
	RepositoryIdentity Identity   `json:"repository_identity"`
	ComponentIdentity  Identity   `json:"component_identity"`
	ProcessIdentity    Identity   `json:"process_identity"`
	LDAP               LDAPConfig `json:"ldap"`
	RBAC               RBACConfig `json:"rbac"`
	Attestation        Attestation `json:"attestation"`
}

type RuntimeOptions struct {
	RepositoryIdentity string
	ComponentType      string
	ComponentIdentity  string
	BuildHash          string
	PrivateKeyPath     string
	PublicKeyPath      string
	LDAP               LDAPConfig
	RBAC               RBACConfig
	ExecutablePath     string
	EnforceSigning     bool
	EnforceDirectory   bool
}

func BuildRuntimePolicy(options RuntimeOptions) (RuntimePolicy, error) {
	repositoryIdentity, err := ParseIdentity(options.RepositoryIdentity)
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("repository identity: %w", err)
	}

	componentIdentity, err := resolveComponentIdentity(options.ComponentType, options.ComponentIdentity)
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("component identity: %w", err)
	}

	processIdentity, err := NewIdentity("process")
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("process identity: %w", err)
	}

	binaryHash, err := HashFile(options.ExecutablePath)
	if err != nil {
		return RuntimePolicy{}, fmt.Errorf("binary hash: %w", err)
	}

	attestation := NewAttestation(
		repositoryIdentity.Raw,
		componentIdentity.Raw,
		processIdentity.Raw,
		binaryHash,
		options.BuildHash,
	)

	if options.EnforceSigning {
		keys, err := LoadKeys(options.PrivateKeyPath, options.PublicKeyPath)
		if err != nil {
			return RuntimePolicy{}, fmt.Errorf("load keys: %w", err)
		}

		signed, err := SignAttestation(keys, attestation)
		if err != nil {
			return RuntimePolicy{}, fmt.Errorf("sign attestation: %w", err)
		}
		if err := VerifyAttestation(keys, signed); err != nil {
			return RuntimePolicy{}, fmt.Errorf("verify attestation: %w", err)
		}
		attestation = signed
	}

	policy := RuntimePolicy{
		RepositoryIdentity: repositoryIdentity,
		ComponentIdentity:  componentIdentity,
		ProcessIdentity:    processIdentity,
		LDAP:               options.LDAP,
		RBAC:               options.RBAC,
		Attestation:        attestation,
	}

	if options.EnforceDirectory {
		if err := policy.ValidateDirectoryRequirements(); err != nil {
			return RuntimePolicy{}, err
		}
		if err := EnforceDirectoryPolicy(policy); err != nil {
			return RuntimePolicy{}, err
		}
	}

	return policy, nil
}

func (policy RuntimePolicy) ValidateDirectoryRequirements() error {
	if policy.LDAP.Address == "" {
		return fmt.Errorf("ldap address is required when directory enforcement is enabled")
	}
	if policy.RBAC.Realm == "" {
		return fmt.Errorf("rbac realm is required when directory enforcement is enabled")
	}
	if policy.RBAC.RoleBaseDN == "" {
		return fmt.Errorf("rbac role base DN is required when directory enforcement is enabled")
	}
	if policy.RBAC.RequiredRole == "" {
		return fmt.Errorf("rbac required role is required when directory enforcement is enabled")
	}
	if policy.LDAP.ComponentBaseDN == "" || policy.LDAP.RepositoryBaseDN == "" || policy.LDAP.AuditBaseDN == "" {
		return fmt.Errorf("ldap base DNs are required when directory enforcement is enabled")
	}
	if policy.LDAP.ServiceAccountDN != "" && policy.LDAP.ServiceAccountPasswordFile == "" {
		return fmt.Errorf("ldap service account password file is required when ldap service account DN is set")
	}
	return nil
}

func resolveComponentIdentity(componentType, raw string) (Identity, error) {
	if raw == "" {
		return NewIdentity(componentType)
	}

	identity, err := ParseIdentity(raw)
	if err != nil {
		return Identity{}, err
	}
	if componentType != "" && identity.ComponentType != componentType {
		return Identity{}, fmt.Errorf("identity %q does not match component type %q", raw, componentType)
	}
	return identity, nil
}

func ExecutablePath() string {
	executablePath, err := os.Executable()
	if err != nil {
		return ""
	}
	return executablePath
}