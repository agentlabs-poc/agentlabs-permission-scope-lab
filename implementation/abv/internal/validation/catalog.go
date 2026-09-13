package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"strings"
	"unicode/utf8"
)

func CheckPermissionRegistration(c domain.Catalog, definition domain.PermissionDefinition, supportedKeys []string) error {
	if err := codec.PermissionList([]string{definition.ID}); err != nil {
		return err
	}
	if _, exists := c.Permissions[definition.ID]; exists {
		return domain.ErrConflict
	}
	if !definition.Active {
		return domain.ErrRejected
	}
	return selectedKeys(c, supportedKeys)
}

// CheckPermissionStatus validates a status change against the catalog. The
// identifier must already be registered: a status change never creates one,
// because Q-126 makes an identifier's meaning permanent. Setting the status a
// definition already holds is valid and writes nothing new.
func CheckPermissionStatus(c domain.Catalog, id string, active bool) error {
	if err := codec.PermissionList([]string{id}); err != nil {
		return err
	}
	existing, ok := c.Permissions[id]
	if !ok || existing.ID != id {
		return domain.ErrNotFound
	}
	return nil
}

func CheckScopeRegistration(c domain.Catalog, definition domain.ScopeDefinition) error {
	if strings.TrimSpace(definition.Key) == "" || !utf8.ValidString(definition.Key) || definition.Key == "*" {
		return domain.ErrMalformed
	}
	if definition.AllowedTokens == nil {
		return domain.ErrMalformed
	}
	if _, exists := c.Scopes[definition.Key]; exists {
		return domain.ErrConflict
	}
	return selectedTokens(definition.AllowedTokens)
}
