package codec

import (
	"agentlabs.local/abv/domain"
	"strconv"
	"strings"
)

// A role record's identity is its id and its revision, and both live in key
// slots: the id in key4, the revision in key5. This file is the single place
// either is rendered for storage or parsed back, the same rule the permission
// identifier follows — storage never formats a key itself.
const (
	// RoleIDAlphabet is base 36: the digits and the lowercase letters, which is
	// what a Snowflake renders to. An id is generated rather than chosen, so it
	// needs no separators and no case distinction.
	RoleIDAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

	// RoleIDMaxLen is 13, the base-36 width of the largest int64 — the ceiling a
	// Snowflake can reach.
	RoleIDMaxLen = 13

	// RevisionWidth pads a revision to a fixed width because a key slot is TEXT
	// and "10" sorts before "2". Ten digits exceeds any plausible revision count
	// and keeps ORDER BY key5 identical to numeric order.
	RevisionWidth = 10
)

// ValidRoleID reports whether an id is a base-36 token that could be a
// Snowflake. It does not check that the value decodes to any particular number:
// the generator lives in the auth service, and Auth-AL validates shape, not
// provenance.
func ValidRoleID(id string) bool {
	if id == "" || len(id) > RoleIDMaxLen {
		return false
	}
	for _, r := range id {
		if !strings.ContainsRune(RoleIDAlphabet, r) {
			return false
		}
	}
	return true
}

// RenderRevision lays a revision out for its key slot, zero-padded.
func RenderRevision(revision int64) (string, error) {
	if revision <= 0 {
		return "", domain.ErrMalformed
	}
	rendered := strconv.FormatInt(revision, 10)
	if len(rendered) > RevisionWidth {
		return "", domain.ErrMalformed
	}
	return strings.Repeat("0", RevisionWidth-len(rendered)) + rendered, nil
}

// ParseRevision reads a revision back out of its slot. Padding is a storage
// rendering, so a caller never sees it: the contract carries an int64.
func ParseRevision(slot string) (int64, error) {
	if len(slot) != RevisionWidth {
		return 0, domain.ErrMalformed
	}
	revision, err := strconv.ParseInt(slot, 10, 64)
	if err != nil || revision <= 0 {
		return 0, domain.ErrMalformed
	}
	return revision, nil
}
