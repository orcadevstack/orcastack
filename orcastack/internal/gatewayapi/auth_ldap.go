package gatewayapi

import (
	"errors"
	"fmt"
	"os"
	"strings"

	platformconfig "github.com/orcastack/orcastackapi/internal/platform/config"
	ldap "github.com/go-ldap/ldap/v3"
)

var errInvalidCredentials = errors.New("invalid username or password")

type ldapAuthConfig struct {
	Address                    string
	ServiceAccountDN           string
	ServiceAccountPasswordFile string
	UserBaseDN                 string
	UserFilter                 string
	UsernameAttribute          string
	FullNameAttribute          string
	EmailAttribute             string
	GroupAttribute             string
	AdminGroupDN               string
	ReleaseGroupDN             string
	DefaultRole                string
	DefaultRealm               string
}

func loadLDAPAuthConfig() ldapAuthConfig {
	return ldapAuthConfig{
		Address:                    platformconfig.String("ORCASTACK_AUTH_LDAP_ADDRESS", platformconfig.String("ORCASTACK_LDAP_ADDRESS", "")),
		ServiceAccountDN:           platformconfig.String("ORCASTACK_AUTH_LDAP_SERVICE_ACCOUNT_DN", platformconfig.String("ORCASTACK_LDAP_SERVICE_ACCOUNT_DN", "")),
		ServiceAccountPasswordFile: platformconfig.String("ORCASTACK_AUTH_LDAP_SERVICE_ACCOUNT_PASSWORD_FILE", platformconfig.String("ORCASTACK_LDAP_SERVICE_ACCOUNT_PASSWORD_FILE", "")),
		UserBaseDN:                 platformconfig.String("ORCASTACK_AUTH_LDAP_USER_BASE_DN", ""),
		UserFilter:                 platformconfig.String("ORCASTACK_AUTH_LDAP_USER_FILTER", "(|(uid=%s)(mail=%s)(sAMAccountName=%s))"),
		UsernameAttribute:          platformconfig.String("ORCASTACK_AUTH_LDAP_USERNAME_ATTRIBUTE", "uid"),
		FullNameAttribute:          platformconfig.String("ORCASTACK_AUTH_LDAP_FULL_NAME_ATTRIBUTE", "cn"),
		EmailAttribute:             platformconfig.String("ORCASTACK_AUTH_LDAP_EMAIL_ATTRIBUTE", "mail"),
		GroupAttribute:             platformconfig.String("ORCASTACK_AUTH_LDAP_GROUP_ATTRIBUTE", "memberOf"),
		AdminGroupDN:               platformconfig.String("ORCASTACK_AUTH_LDAP_ADMIN_GROUP_DN", ""),
		ReleaseGroupDN:             platformconfig.String("ORCASTACK_AUTH_LDAP_RELEASE_GROUP_DN", ""),
		DefaultRole:                platformconfig.String("ORCASTACK_AUTH_DEFAULT_ROLE", "platform-operator"),
		DefaultRealm:               platformconfig.String("ORCASTACK_AUTH_DEFAULT_REALM", platformconfig.String("ORCASTACK_RBAC_REALM", "platform")),
	}
}

func authenticateLDAPUser(username, password string) (AuthUser, error) {
	trimmedUsername := strings.TrimSpace(username)
	trimmedPassword := strings.TrimSpace(password)
	if trimmedUsername == "" || trimmedPassword == "" {
		return AuthUser{}, errInvalidCredentials
	}

	cfg := loadLDAPAuthConfig()
	if err := cfg.validate(); err != nil {
		return AuthUser{}, err
	}

	entry, userDN, err := cfg.lookupUser(trimmedUsername)
	if err != nil {
		return AuthUser{}, err
	}

	if err := cfg.verifyPassword(userDN, trimmedPassword); err != nil {
		return AuthUser{}, err
	}

	return cfg.authUserFromEntry(trimmedUsername, userDN, entry), nil
}

func (cfg ldapAuthConfig) validate() error {
	if cfg.Address == "" {
		return fmt.Errorf("ldap authentication is not configured: set ORCASTACK_AUTH_LDAP_ADDRESS or ORCASTACK_LDAP_ADDRESS")
	}
	if cfg.UserBaseDN == "" {
		return fmt.Errorf("ldap authentication is not configured: set ORCASTACK_AUTH_LDAP_USER_BASE_DN")
	}
	if cfg.UserFilter == "" {
		return fmt.Errorf("ldap authentication is not configured: set ORCASTACK_AUTH_LDAP_USER_FILTER")
	}
	if cfg.ServiceAccountDN != "" && cfg.ServiceAccountPasswordFile == "" {
		return fmt.Errorf("ldap service account password file is required when a service account DN is configured")
	}
	return nil
}

func (cfg ldapAuthConfig) lookupUser(username string) (*ldap.Entry, string, error) {
	connection, err := ldap.DialURL(cfg.Address)
	if err != nil {
		return nil, "", fmt.Errorf("dial ldap %s: %w", cfg.Address, err)
	}
	defer connection.Close()

	if cfg.ServiceAccountDN != "" {
		password, err := readSecretFile(cfg.ServiceAccountPasswordFile)
		if err != nil {
			return nil, "", err
		}
		if err := connection.Bind(cfg.ServiceAccountDN, password); err != nil {
			return nil, "", fmt.Errorf("ldap service account bind failed: %w", err)
		}
	}

	filter, err := cfg.userSearchFilter(username)
	if err != nil {
		return nil, "", err
	}

	result, err := connection.Search(ldap.NewSearchRequest(
		cfg.UserBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		2,
		0,
		false,
		filter,
		cfg.searchAttributes(),
		nil,
	))
	if err != nil {
		return nil, "", fmt.Errorf("ldap user search failed: %w", err)
	}
	if len(result.Entries) != 1 {
		return nil, "", errInvalidCredentials
	}

	entry := result.Entries[0]
	if strings.TrimSpace(entry.DN) == "" {
		return nil, "", errInvalidCredentials
	}

	return entry, entry.DN, nil
}

func (cfg ldapAuthConfig) verifyPassword(userDN, password string) error {
	connection, err := ldap.DialURL(cfg.Address)
	if err != nil {
		return fmt.Errorf("dial ldap %s: %w", cfg.Address, err)
	}
	defer connection.Close()

	if err := connection.Bind(userDN, password); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return errInvalidCredentials
		}
		return fmt.Errorf("ldap user bind failed: %w", err)
	}

	return nil
}

func (cfg ldapAuthConfig) userSearchFilter(username string) (string, error) {
	escaped := ldap.EscapeFilter(username)
	placeholderCount := strings.Count(cfg.UserFilter, "%s")
	if placeholderCount == 0 {
		return "", fmt.Errorf("ldap user filter must contain at least one %%s placeholder")
	}

	args := make([]any, 0, placeholderCount)
	for index := 0; index < placeholderCount; index++ {
		args = append(args, escaped)
	}
	return fmt.Sprintf(cfg.UserFilter, args...), nil
}

func (cfg ldapAuthConfig) searchAttributes() []string {
	attributes := []string{cfg.UsernameAttribute, cfg.FullNameAttribute, cfg.EmailAttribute, cfg.GroupAttribute}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(attributes))
	for _, attribute := range attributes {
		trimmed := strings.TrimSpace(attribute)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func (cfg ldapAuthConfig) authUserFromEntry(fallbackUsername, userDN string, entry *ldap.Entry) AuthUser {
	username := cfg.entryValue(entry, cfg.UsernameAttribute, fallbackUsername)
	fullName := cfg.entryValue(entry, cfg.FullNameAttribute, username)
	email := cfg.entryValue(entry, cfg.EmailAttribute, "")
	groups := cfg.entryValues(entry, cfg.GroupAttribute)
	role, permissions := cfg.roleAndPermissions(groups)

	return AuthUser{
		Username:    username,
		FullName:    fullName,
		Email:       email,
		Role:        role,
		Identity:    "ldap:user:" + username,
		RBACRealm:   cfg.DefaultRealm,
		Permissions: permissions,
	}
}

func (cfg ldapAuthConfig) entryValue(entry *ldap.Entry, attribute, fallback string) string {
	if strings.TrimSpace(attribute) == "" {
		return fallback
	}
	value := strings.TrimSpace(entry.GetAttributeValue(attribute))
	if value == "" {
		return fallback
	}
	return value
}

func (cfg ldapAuthConfig) entryValues(entry *ldap.Entry, attribute string) []string {
	if strings.TrimSpace(attribute) == "" {
		return nil
	}
	return entry.GetAttributeValues(attribute)
}

func (cfg ldapAuthConfig) roleAndPermissions(groups []string) (string, []string) {
	if groupMatch(groups, cfg.AdminGroupDN) {
		return "platform-admin", []string{"repositories:read", "pipelines:write", "deployments:write", "control-panel:admin", "community:read"}
	}
	if groupMatch(groups, cfg.ReleaseGroupDN) {
		return "release-operator", []string{"repositories:read", "pipelines:write", "deployments:write", "control-panel:read", "community:read"}
	}
	return cfg.DefaultRole, []string{"repositories:read", "pipelines:write", "deployments:write", "control-panel:read", "community:read"}
}

func groupMatch(groups []string, expected string) bool {
	needle := normalizeGroup(expected)
	if needle == "" {
		return false
	}

	for _, group := range groups {
		if normalizeGroup(group) == needle {
			return true
		}
	}
	return false
}

func normalizeGroup(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func readSecretFile(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("secret file path is required")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read secret file %s: %w", path, err)
	}
	return strings.TrimSpace(string(content)), nil
}
