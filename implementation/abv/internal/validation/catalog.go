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
