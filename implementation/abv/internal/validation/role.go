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
	// The id is a base-36 Snowflake, not free text. A generated id removes the
	// ambiguity free text had — Payroll-Admin and payroll-admin were two roles
	// that read as one — and it never needs renaming.
	if !codec.ValidRoleID(role.ID) {
		return domain.ErrMalformed
	}
	// The name is a human label, and deliberately not unique: identity is the id.
	// It must still be storable, because it occupies a key slot.
	if strings.TrimSpace(role.Name) == "" || !utf8.ValidString(role.Name) {
		return domain.ErrMalformed
	}
	// The revision must render into its slot, which is what rejects <= 0 and
	// anything too wide to pad.
	if _, err := codec.RenderRevision(role.Revision); err != nil {
		return err
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
