package app

import (
	"github.com/orcastack/orcastackapi/internal/platform/config"
	"github.com/orcastack/orcastackapi/internal/platform/security"
)

const (
	DefaultGatewayIdentity   = security.DefaultGatewayIdentity
	DefaultGitIdentity       = security.DefaultGitIdentity
	DefaultReviewIdentity    = security.DefaultReviewIdentity
	DefaultCIIdentity        = security.DefaultCIIdentity
	DefaultCDIdentity        = security.DefaultCDIdentity
	DefaultAnalyticsIdentity = security.DefaultAnalyticsIdentity
	DefaultRunnerIdentity    = security.DefaultRunnerIdentity
	DefaultHWAutoIdentity    = security.DefaultHWAutoIdentity
	DefaultSWAutoIdentity    = security.DefaultSWAutoIdentity
	DefaultDeviceOrchIdentity = security.DefaultDeviceOrchIdentity
)

func WithServiceSecurity(cfg Config, componentIdentityKey, componentIdentityDefault string) Config {
	cfg.ComponentType = "service"
	cfg.ComponentIdentity = config.String(componentIdentityKey, componentIdentityDefault)
	cfg.RepositoryIdentity = config.String("ORCASTACK_REPOSITORY_IDENTITY", security.DefaultRepositoryIdentity)
	cfg.BuildHash = config.String("ORCASTACK_BUILD_HASH", "sha256:unavailable")
	cfg.PrivateKeyPath = config.String("ORCASTACK_SIGNING_PRIVATE_KEY_PATH", "")
	cfg.PublicKeyPath = config.String("ORCASTACK_SIGNING_PUBLIC_KEY_PATH", "")
	cfg.EnforceSigning = config.Bool("ORCASTACK_ENFORCE_SIGNING", false)
	cfg.EnforceDirectory = config.Bool("ORCASTACK_ENFORCE_DIRECTORY", false)
	cfg.LDAP = security.LDAPConfig{
		Address:                    config.String("ORCASTACK_LDAP_ADDRESS", ""),
		ServiceAccountDN:           config.String("ORCASTACK_LDAP_SERVICE_ACCOUNT_DN", ""),
		ServiceAccountPasswordFile: config.String("ORCASTACK_LDAP_SERVICE_ACCOUNT_PASSWORD_FILE", ""),
		ComponentBaseDN:            config.String("ORCASTACK_LDAP_COMPONENT_BASE_DN", ""),
		RepositoryBaseDN:           config.String("ORCASTACK_LDAP_REPOSITORY_BASE_DN", ""),
		AuditBaseDN:                config.String("ORCASTACK_LDAP_AUDIT_BASE_DN", ""),
		AutoRegister:               config.Bool("ORCASTACK_LDAP_AUTO_REGISTER", false),
	}
	cfg.RBAC = security.RBACConfig{
		Realm:        config.String("ORCASTACK_RBAC_REALM", ""),
		RoleBaseDN:   config.String("ORCASTACK_RBAC_ROLE_BASE_DN", ""),
		RequiredRole: config.String("ORCASTACK_RBAC_REQUIRED_ROLE", "platform-service"),
	}
	return cfg
}