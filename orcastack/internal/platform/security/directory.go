package security

import (
	"fmt"
	"os"
	"strings"

	ldap "github.com/go-ldap/ldap/v3"
)

func EnforceDirectoryPolicy(policy RuntimePolicy) error {
	connection, err := ldap.DialURL(policy.LDAP.Address)
	if err != nil {
		return fmt.Errorf("dial ldap %s: %w", policy.LDAP.Address, err)
	}
	defer connection.Close()

	if policy.LDAP.ServiceAccountDN != "" {
		password, err := readSecretFile(policy.LDAP.ServiceAccountPasswordFile)
		if err != nil {
			return err
		}
		if err := connection.Bind(policy.LDAP.ServiceAccountDN, password); err != nil {
			return fmt.Errorf("ldap bind failed: %w", err)
		}
	}

	repositoryDN, err := ensureDirectoryIdentity(connection, policy.LDAP.RepositoryBaseDN, policy.RepositoryIdentity, policy.LDAP.AutoRegister)
	if err != nil {
		return fmt.Errorf("repository identity enforcement failed: %w", err)
	}

	componentDN, err := ensureDirectoryIdentity(connection, policy.LDAP.ComponentBaseDN, policy.ComponentIdentity, policy.LDAP.AutoRegister)
	if err != nil {
		return fmt.Errorf("component identity enforcement failed: %w", err)
	}

	if _, err := ensureDirectoryIdentity(connection, policy.LDAP.AuditBaseDN, policy.ProcessIdentity, policy.LDAP.AutoRegister); err != nil {
		return fmt.Errorf("process audit enforcement failed: %w", err)
	}

	if err := enforceRBACMembership(connection, policy.RBAC, componentDN, repositoryDN); err != nil {
		return err
	}

	return nil
}

func ensureDirectoryIdentity(connection *ldap.Conn, baseDN string, identity Identity, autoRegister bool) (string, error) {
	request := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		identityFilter(identity.Raw),
		[]string{"dn", "cn", "description"},
		nil,
	)

	result, err := connection.Search(request)
	if err != nil {
		return "", fmt.Errorf("search %s in %s: %w", identity.Raw, baseDN, err)
	}
	if len(result.Entries) > 0 {
		return result.Entries[0].DN, nil
	}
	if !autoRegister {
		return "", fmt.Errorf("identity %s is not registered under %s", identity.Raw, baseDN)
	}

	dn := fmt.Sprintf("cn=%s,%s", identity.UUID, baseDN)
	addRequest := ldap.NewAddRequest(dn, nil)
	addRequest.Attribute("objectClass", []string{"top", "extensibleObject"})
	addRequest.Attribute("cn", []string{identity.UUID})
	addRequest.Attribute("description", []string{identity.Raw})
	addRequest.Attribute("uid", []string{identity.Raw})
	if err := connection.Add(addRequest); err != nil {
		return "", fmt.Errorf("register %s in ldap: %w", identity.Raw, err)
	}

	return dn, nil
}

func enforceRBACMembership(connection *ldap.Conn, config RBACConfig, componentDN, repositoryDN string) error {
	if config.RoleBaseDN == "" {
		return fmt.Errorf("rbac role base DN is required when directory enforcement is enabled")
	}
	if config.RequiredRole == "" {
		return fmt.Errorf("rbac required role is required when directory enforcement is enabled")
	}

	request := ldap.NewSearchRequest(
		config.RoleBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		rbacFilter(config, componentDN, repositoryDN),
		[]string{"dn"},
		nil,
	)

	result, err := connection.Search(request)
	if err != nil {
		return fmt.Errorf("search rbac role %s: %w", config.RequiredRole, err)
	}
	if len(result.Entries) == 0 {
		return fmt.Errorf("component is not authorized for required role %s in realm %s", config.RequiredRole, config.Realm)
	}

	return nil
}

func identityFilter(raw string) string {
	escaped := ldap.EscapeFilter(raw)
	return fmt.Sprintf("(|(description=%s)(uid=%s)(cn=%s))", escaped, escaped, escaped)
}

func rbacFilter(config RBACConfig, componentDN, repositoryDN string) string {
	role := ldap.EscapeFilter(config.RequiredRole)
	realm := ldap.EscapeFilter(config.Realm)
	component := ldap.EscapeFilter(componentDN)
	repository := ldap.EscapeFilter(repositoryDN)
	return fmt.Sprintf("(&(|(cn=%s)(ou=%s)(roleName=%s))(|(businessCategory=%s)(description=%s)(o=%s))(|(member=%s)(uniqueMember=%s)(roleOccupant=%s)(owner=%s)))", role, role, role, realm, realm, realm, component, component, component, repository)
}

func readSecretFile(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("secret file path is required")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read secret file %s: %w", path, err)
	}
	return strings.TrimSpace(string(content)), nil
}