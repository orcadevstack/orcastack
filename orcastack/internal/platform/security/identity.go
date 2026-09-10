package security

import (
	"crypto/rand"
	"fmt"
	"regexp"
)

var identityPattern = regexp.MustCompile(`^orca:([a-z][a-z0-9-]*):([0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})$`)

type Identity struct {
	Raw           string
	ComponentType string
	UUID          string
}

func NewIdentity(componentType string) (Identity, error) {
	if componentType == "" {
		return Identity{}, fmt.Errorf("component type is required")
	}

	uuid, err := newUUIDv4()
	if err != nil {
		return Identity{}, err
	}

	raw := fmt.Sprintf("orca:%s:%s", componentType, uuid)
	return ParseIdentity(raw)
}

func ParseIdentity(raw string) (Identity, error) {
	matches := identityPattern.FindStringSubmatch(raw)
	if matches == nil {
		return Identity{}, fmt.Errorf("identity %q does not match orca:<component-type>:<uuidv4>", raw)
	}

	return Identity{
		Raw:           raw,
		ComponentType: matches[1],
		UUID:          matches[2],
	}, nil
}

func MustParseIdentity(raw string) Identity {
	identity, err := ParseIdentity(raw)
	if err != nil {
		panic(err)
	}
	return identity
}

func newUUIDv4() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf(
		"%08x-%04x-%04x-%04x-%012x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	), nil
}