package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"strings"
	"unicode/utf8"
)

func CheckRolePublication(area domain.Area, catalog domain.Catalog, role domain.RoleContent) error {
	if err := area.Validate(); err != nil {
		return err
	}
	if catalog.ApplicationID != area.ApplicationID() {
		return domain.ErrRejected
	}
	if strings.TrimSpace(role.ID) == "" || strings.Contains(role.ID, "*") || !utf8.ValidString(role.ID) || role.Revision <= 0 {
		return domain.ErrMalformed
	}
	if err := codec.PermissionList(role.Permissions); err != nil {
		return err
	}
	for _, permission := range role.Permissions {
		definition, ok := catalog.Permissions[permission]
		if !ok || definition.ID != permission || !definition.Active {
			return domain.ErrRejected
		}
	}
	return nil
}
