package codec

import (
	"agentlabs.local/abv/domain"
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func DecodeContent(raw []byte) (domain.GrantContent, error) {
	obj, err := object(raw)
	if err != nil {
		return domain.GrantContent{}, err
	}
	if err = version(obj); err != nil {
		return domain.GrantContent{}, err
	}
	if err = keys(obj, "version", "grant_id", "revision", "parent_grant_id", "permissions", "role_id", "role_revision", "scope", "validity"); err != nil {
		return domain.GrantContent{}, err
	}
	if parent, exists := obj["parent_grant_id"]; exists {
		value, ok := parent.(string)
		if !ok || blank(value) {
			return domain.GrantContent{}, domain.ErrMalformed
		}
	}
	_, direct := obj["permissions"]
	_, role := obj["role_id"]
	_, revision := obj["role_revision"]
	if (direct && (role || revision)) || (!direct && (!role || !revision)) {
		return domain.GrantContent{}, domain.ErrMalformed
	}
	if rawValidity, exists := obj["validity"]; exists {
		validity, ok := rawValidity.(map[string]any)
		if !ok {
			return domain.GrantContent{}, domain.ErrMalformed
		}
		if err = keys(validity, "not_before", "expires_at"); err != nil {
			return domain.GrantContent{}, err
		}
	}
	var result domain.GrantContent
	if err = json.Unmarshal(raw, &result); err != nil {
		return domain.GrantContent{}, domain.ErrMalformed
	}
	if err = ValidateContent(result); err != nil {
		return domain.GrantContent{}, err
	}
	return result, nil
}

// ValidateContent also checks typed inputs, which do not necessarily pass through
// the JSON decoder. It does not establish parent trust or active time eligibility.
func ValidateContent(g domain.GrantContent) error {
	if g.Version == "" || !utf8.ValidString(g.Version) {
		return domain.ErrMalformed
	}
	if g.Version != "1" {
		return domain.ErrUnsupported
	}
	if invalidString(g.GrantID) || g.Revision <= 0 || g.Scope == nil {
		return domain.ErrMalformed
	}
	if g.ParentGrantID != "" && invalidString(g.ParentGrantID) {
		return domain.ErrMalformed
	}
	// Three permission sources, and exactly one applies.
	//
	// Q-118: a direct list, or the complete role pair, never a mixture. Root
	// content is the third — it names neither, because its coverage is computed
	// from the registered catalog at resolution rather than stored (Q-122). A
	// stored list on a root is never read, so storing one can only go stale and
	// mislead whoever reads the row.
	//
	// This cannot check that the grant *is* a trusted root: Q-119 is explicit
	// that omitting a field proves nothing, and the evidence lives on the head
	// rather than in the content. What it can check is that content shaped like
	// a root is shaped like one completely — no parent, and no local narrowing,
	// because a root is the ceiling for its whole area and a pre-narrowed root
	// is a ceiling lower than the region it bounds.
	switch {
	case g.Permissions != nil:
		if g.RoleID != "" || g.RoleRevision != 0 {
			return domain.ErrMalformed
		}
		if err := PermissionList(g.Permissions); err != nil {
			return err
		}
	case g.RoleID != "" || g.RoleRevision != 0:
		if invalidString(g.RoleID) || g.RoleRevision <= 0 {
			return domain.ErrMalformed
		}
	default:
		if g.ParentGrantID != "" || len(g.Scope) != 0 {
			return domain.ErrMalformed
		}
	}
	for key, value := range g.Scope {
		if invalidString(key) || invalidString(value) || key == "*" || value == "*" {
			return domain.ErrMalformed
		}
	}
	if g.Validity != nil {
		v := g.Validity
		if v.NotBefore == nil && v.ExpiresAt == nil {
			return domain.ErrUnsupported
		}
		if v.NotBefore != nil && v.ExpiresAt != nil && !v.NotBefore.Before(*v.ExpiresAt) {
			return domain.ErrMalformed
		}
	}
	return nil
}
func PermissionList(values []string) error {
	if len(values) == 0 {
		return domain.ErrMalformed
	}
	seen := map[string]bool{}
	for _, p := range values {
		if invalidString(p) || strings.Contains(p, "*") || seen[p] {
			return domain.ErrMalformed
		}
		seen[p] = true
	}
	return nil
}
func blank(s string) bool         { return strings.TrimSpace(s) == "" }
func invalidString(s string) bool { return blank(s) || !utf8.ValidString(s) }
