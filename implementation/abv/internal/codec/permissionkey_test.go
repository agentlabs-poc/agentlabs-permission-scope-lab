package codec

import (
	"agentlabs.local/abv/domain"
	"errors"
	"strings"
	"testing"
)

func TestParsePermissionDecomposesAndRoundTrips(t *testing.T) {
	for _, id := range []string{
		"hrms:employee:certificate::read",
		"codehost:repository:branch:protection:rule::write",
		"a:b::c",
		"single::verb",
		"hrms::read",                        // a one-segment noun path is legitimate
		"one:two:three:four:five:six::verb", // exactly the envelope's limit
	} {
		key, err := ParsePermission(id)
		if err != nil {
			t.Fatalf("%q: %v", id, err)
		}
		// Round trip is the contract between the layers: a slot layout that
		// cannot rebuild its identifier has silently changed its meaning.
		if got := key.Render(); got != id {
			t.Fatalf("round trip %q -> %q", id, got)
		}
		if len(key.Nouns) > MaxNounSegments {
			t.Fatalf("%q: %d nouns exceeds the envelope", id, len(key.Nouns))
		}
	}
}

func TestParsePermissionPlacesSegmentsCorrectly(t *testing.T) {
	key, err := ParsePermission("hrms:employee:certificate::read")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"hrms", "employee", "certificate"}
	if len(key.Nouns) != len(want) {
		t.Fatalf("nouns=%v, want %v", key.Nouns, want)
	}
	for i := range want {
		if key.Nouns[i] != want[i] {
			t.Fatalf("noun %d = %q, want %q", i, key.Nouns[i], want[i])
		}
	}
	if key.Verb != "read" {
		t.Fatalf("verb=%q, want read", key.Verb)
	}
}

func TestParsePermissionRejectsWhatHasNoSlotRepresentation(t *testing.T) {
	for name, id := range map[string]string{
		"no verb separator":   "hrms:employee:certificate:read",
		"two verb separators": "hrms::employee::read",
		"empty verb":          "hrms:employee::",
		"leading empty noun":  ":hrms::read",
		"trailing empty noun": "hrms::read:",
		"empty middle noun":   "hrms::employee::read",
		"verb holds a colon":  "hrms:employee::read:write",
		"wildcard":            "hrms:*::read",
		"blank":               "   ",
		"too deep":            "a:b:c:d:e:f:g:h::verb",
	} {
		if _, err := ParsePermission(id); !errors.Is(err, domain.ErrMalformed) {
			t.Errorf("%s (%q): err=%v, want ErrMalformed", name, id, err)
		}
	}
}

func TestNounPrefixRequiresWholeSegments(t *testing.T) {
	segments, err := NounPrefix("hrms:employee:")
	if err != nil || len(segments) != 2 || segments[0] != "hrms" || segments[1] != "employee" {
		t.Fatalf("segments=%v err=%v", segments, err)
	}
	if segments, err := NounPrefix(""); err != nil || segments != nil {
		t.Fatalf("empty prefix selects everything: %v %v", segments, err)
	}

	// A partial segment has no structural form. Rejecting it keeps a caller
	// from silently receiving a query that cannot use the index.
	for name, prefix := range map[string]string{
		"partial segment":    "hrms:emp",
		"no trailing colon":  "hrms:employee",
		"wildcard":           "hrms:*:",
		"reaches into verbs": "hrms:employee::",
		"empty segment":      "hrms::",
		"too deep":           "a:b:c:d:e:f:g:h:",
	} {
		if _, err := NounPrefix(prefix); !errors.Is(err, domain.ErrMalformed) {
			t.Errorf("%s (%q): err=%v, want ErrMalformed", name, prefix, err)
		}
	}
}

// TestNounPrefixMatchesParsedIdentifiers ties the two halves of the codec
// together: a prefix built from an identifier's own leading segments must
// select that identifier.
func TestNounPrefixMatchesParsedIdentifiers(t *testing.T) {
	const id = "hrms:employee:certificate::read"
	key, err := ParsePermission(id)
	if err != nil {
		t.Fatal(err)
	}
	for depth := 1; depth <= len(key.Nouns); depth++ {
		prefix := strings.Join(key.Nouns[:depth], ":") + ":"
		segments, err := NounPrefix(prefix)
		if err != nil {
			t.Fatalf("depth %d prefix %q: %v", depth, prefix, err)
		}
		for i, segment := range segments {
			if key.Nouns[i] != segment {
				t.Fatalf("depth %d: prefix segment %d = %q, identifier has %q", depth, i, segment, key.Nouns[i])
			}
		}
	}
}
