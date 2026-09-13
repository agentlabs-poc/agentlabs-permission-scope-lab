package codec

import (
	"agentlabs.local/abv/domain"
	"strings"
)

// A permission identifier is a noun path of one or more segments, then "::",
// then exactly one verb:
//
//	hrms:employee:certificate::read
//
// Storage decomposes it across key slots — the noun path left to right from
// key3, the verb pinned to key10 — so the identifier must be parseable. An
// identifier that cannot be decomposed has no canonical storage representation,
// which is why the shape is enforced here rather than left to convention.
//
// This is the single encoder and parser for the form. Storage never splits the
// string itself, and the contract never assembles one by hand.
const (
	// MaxNounSegments is what the envelope leaves for the noun path: ten slots,
	// less the domain namespace, the record type and the reserved verb slot.
	MaxNounSegments = 7

	verbSeparator = "::"
	nounSeparator = ":"
)

// PermissionKey is an identifier decomposed into the parts storage holds.
type PermissionKey struct {
	Nouns []string
	Verb  string
}

// ParsePermission decomposes a canonical identifier. It rejects anything that
// has no slot representation: no verb separator, more than one, an empty
// segment, a noun path deeper than the envelope allows, or a wildcard.
func ParsePermission(id string) (PermissionKey, error) {
	if invalidString(id) || strings.Contains(id, "*") {
		return PermissionKey{}, domain.ErrMalformed
	}
	nouns, verb, found := strings.Cut(id, verbSeparator)
	if !found || strings.Contains(verb, verbSeparator) {
		return PermissionKey{}, domain.ErrMalformed
	}
	if invalidString(verb) || strings.Contains(verb, nounSeparator) {
		return PermissionKey{}, domain.ErrMalformed
	}
	segments := strings.Split(nouns, nounSeparator)
	if len(segments) == 0 || len(segments) > MaxNounSegments {
		return PermissionKey{}, domain.ErrMalformed
	}
	for _, segment := range segments {
		if invalidString(segment) {
			return PermissionKey{}, domain.ErrMalformed
		}
	}
	return PermissionKey{Nouns: segments, Verb: verb}, nil
}

// Render reassembles the canonical identifier. Parse and Render round-trip
// exactly; a test asserts it for every shape the parser accepts.
func (k PermissionKey) Render() string {
	return strings.Join(k.Nouns, nounSeparator) + verbSeparator + k.Verb
}

// NounPrefix decomposes a listing prefix into whole noun segments.
//
// A prefix must end on a segment boundary, because only whole segments have a
// structural form: they become equality on leading key slots. A prefix ending
// mid-segment would have to fall back to a string match inside one slot, which
// is the cost decomposition exists to avoid — so it is rejected rather than
// silently answered slowly.
//
// An empty prefix selects everything and is valid.
func NounPrefix(prefix string) ([]string, error) {
	if prefix == "" {
		return nil, nil
	}
	if invalidString(prefix) || strings.Contains(prefix, "*") || strings.Contains(prefix, verbSeparator) {
		return nil, domain.ErrMalformed
	}
	trimmed, ok := strings.CutSuffix(prefix, nounSeparator)
	if !ok {
		// Without a trailing separator the last segment may be partial, and a
		// partial segment has no slot representation.
		return nil, domain.ErrMalformed
	}
	segments := strings.Split(trimmed, nounSeparator)
	if len(segments) > MaxNounSegments {
		return nil, domain.ErrMalformed
	}
	for _, segment := range segments {
		if invalidString(segment) {
			return nil, domain.ErrMalformed
		}
	}
	return segments, nil
}

// PermissionSlots is the key-slot representation of an identifier: the noun
// path in key3…key9 padded with empty strings, and the verb in key10.
type PermissionSlots [8]string

// Slots lays the key out for storage. Storage writes these columns verbatim and
// never parses the identifier itself.
func (k PermissionKey) Slots() PermissionSlots {
	var slots PermissionSlots
	copy(slots[:MaxNounSegments], k.Nouns)
	slots[MaxNounSegments] = k.Verb
	return slots
}

// PermissionFromSlots rebuilds the identifier a row holds. It is the inverse of
// Slots, and the round-trip is what keeps storage and the contract from drifting:
// a layout that cannot rebuild its identifier has silently changed its meaning.
func PermissionFromSlots(slots PermissionSlots) (string, error) {
	verb := slots[MaxNounSegments]
	var nouns []string
	for _, noun := range slots[:MaxNounSegments] {
		if noun == "" {
			break
		}
		nouns = append(nouns, noun)
	}
	if len(nouns) == 0 || invalidString(verb) {
		return "", domain.ErrMalformed
	}
	// Padding must be contiguous: a gap means the row was not written by this
	// encoder and its identifier cannot be trusted.
	for _, noun := range slots[len(nouns):MaxNounSegments] {
		if noun != "" {
			return "", domain.ErrMalformed
		}
	}
	return PermissionKey{Nouns: nouns, Verb: verb}.Render(), nil
}
