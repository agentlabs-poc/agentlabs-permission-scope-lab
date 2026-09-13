package validation

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/codec"
	"strings"
	"unicode/utf8"
)

func CheckPermissionRegistration(c domain.Catalog, definition domain.PermissionDefinition) error {
	return CheckPermissionRegistrationAt(domain.ApplicationBoundary, c, definition)
}

// CheckPermissionRegistrationAt validates a registration at a named boundary.
//
// The leading-noun rule is the only thing the boundary changes. At the
// application boundary a permission's first noun must be the application, which
// is what lets key3 answer "which application owns this row" by construction. At
// the platform boundary key3 holds a namespace the platform defines and no
// application can claim, so there is nothing to compare it against — and
// comparing would require Auth-AL to hold the platform's reserved list, which is
// the auth service's vocabulary rather than ours.
func CheckPermissionRegistrationAt(boundary domain.Boundary, c domain.Catalog, definition domain.PermissionDefinition) error {
	if err := codec.PermissionList([]string{definition.ID}); err != nil {
		return err
	}
	// The identifier must decompose into the noun path and verb that storage
	// holds in its key slots. One that cannot be parsed has no canonical
	// storage representation, so the shape is enforced here rather than left to
	// convention. This is also the only parser: nothing splits the string
	// itself.
	key, err := codec.ParsePermission(definition.ID)
	if err != nil {
		return err
	}
	if !boundary.Valid() {
		return domain.ErrMalformed
	}
	// At the application boundary the first noun segment IS the application. An
	// identifier starting with anything else is canonically incorrect: key3
	// carries the application, and the identifier would disagree with the row
	// storing it. That rule is also what removes the duplication — the segment
	// is stored once, in key3, and the noun path holds only what follows.
	//
	// At the platform boundary there is no application to compare against.
	if boundary == domain.ApplicationBoundary && key.Nouns[0] != c.ApplicationID {
		return domain.ErrRejected
	}
	if _, exists := c.Permissions[definition.ID]; exists {
		return domain.ErrConflict
	}
	if !definition.Active {
		return domain.ErrRejected
	}
	return nil
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
	// A wildcard anywhere, not merely a key that is exactly "*": a scope key is
	// matched for equality during evaluation, so a key containing "*" could only
	// ever mislead a reader into thinking it matches more than itself.
	if strings.TrimSpace(definition.Key) == "" || !utf8.ValidString(definition.Key) || strings.Contains(definition.Key, "*") {
		return domain.ErrMalformed
	}
	if _, exists := c.Scopes[definition.Key]; exists {
		return domain.ErrConflict
	}
	return nil
}
