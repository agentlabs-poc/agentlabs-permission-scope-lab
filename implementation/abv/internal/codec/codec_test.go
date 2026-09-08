package codec

import (
	"agentlabs.local/abv/domain"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
)

const goodContent = `{"version":"1","grant_id":"G2","revision":1,"parent_grant_id":"G1","permissions":["hrms:employee:certificate::read"],"scope":{"cert":"C17"}}`
const goodAssignment = `{"version":"1","id":"A2","grant_id":"G2","grant_revision":1,"recipient":{"type":"group","id":"Team2"},"status":"enabled"}`

func TestCanonicalFixturesRoundTrip(t *testing.T) {
	for _, name := range []string{"g1", "g2", "a1", "a2"} {
		raw, err := os.ReadFile("../../testdata/" + name + ".json")
		if err != nil {
			t.Fatal(err)
		}
		var record any
		if name[0] == 'g' {
			record, err = DecodeContent(raw)
		} else {
			record, err = DecodeAssignment(raw)
		}
		if err != nil {
			t.Fatal(name, err)
		}
		got, err := json.Marshal(record)
		if err != nil {
			t.Fatal(err)
		}
		var a, b any
		if err = json.Unmarshal(raw, &a); err != nil {
			t.Fatal(err)
		}
		if err = json.Unmarshal(got, &b); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("%s lost fields: %s", name, got)
		}
	}
}

// Accepting duplicate keys, alternate field spelling or discarded restrictions
// would silently change the proposed authority before validation.
func TestContentRejectsAmbiguousOrUnsupportedJSON(t *testing.T) {
	cases := map[string]string{
		"duplicate root":       strings.Replace(goodContent, `"version":"1"`, `"version":"1","version":"1"`, 1),
		"escaped duplicate":    strings.Replace(goodContent, `"scope":{"cert":"C17"}`, `"scope":{"cert":"C17","\u0063ert":"C18"}`, 1),
		"duplicate scope":      strings.Replace(goodContent, `"cert":"C17"`, `"cert":"C17","cert":"C18"`, 1),
		"missing version":      strings.Replace(goodContent, `"version":"1",`, "", 1),
		"number version":       strings.Replace(goodContent, `"version":"1"`, `"version":1`, 1),
		"future version":       strings.Replace(goodContent, `"version":"1"`, `"version":"2"`, 1),
		"case alias":           strings.Replace(goodContent, `"scope"`, `"Scope"`, 1),
		"null scope":           strings.Replace(goodContent, `{"cert":"C17"}`, "null", 1),
		"missing scope":        strings.Replace(goodContent, `,"scope":{"cert":"C17"}`, "", 1),
		"array scope":          strings.Replace(goodContent, `{"cert":"C17"}`, `["C17"]`, 1),
		"numeric value":        strings.Replace(goodContent, `"C17"`, "17", 1),
		"empty value":          strings.Replace(goodContent, `"C17"`, `""`, 1),
		"wildcard":             strings.Replace(goodContent, `"C17"`, `"*"`, 1),
		"recipient on content": strings.Replace(goodContent, `"version":"1"`, `"version":"1","recipient":{"type":"group","id":"Team2"}`, 1),
		"role half":            strings.Replace(goodContent, `"permissions":["hrms:employee:certificate::read"]`, `"role_id":"reader"`, 1),
		"mixed sources":        strings.Replace(goodContent, `"revision":1`, `"revision":1,"role_id":"reader","role_revision":1`, 1),
		"present empty role":   strings.Replace(goodContent, `"revision":1`, `"revision":1,"role_id":""`, 1),
		"present zero role":    strings.Replace(goodContent, `"revision":1`, `"revision":1,"role_revision":0`, 1),
		"null role":            strings.Replace(goodContent, `"revision":1`, `"revision":1,"role_id":null`, 1),
		"conditions":           strings.Replace(goodContent, `"revision":1`, `"revision":1,"conditions":{"approved":true}`, 1),
		"revision zero":        strings.Replace(goodContent, `"revision":1`, `"revision":0`, 1),
		"revision overflow":    strings.Replace(goodContent, `"revision":1`, `"revision":9223372036854775808`, 1),
		"trailing value":       goodContent + " {}",
		"empty object":         "{}", "null": "null", "array": "[]", "truncated": goodContent[:20],
	}
	for name, raw := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeContent([]byte(raw))
			if err == nil {
				t.Fatalf("accepted invalid content: %s -> %#v", raw, got)
			}
			if !reflect.DeepEqual(got, domain.GrantContent{}) {
				t.Fatal("returned partial usable content")
			}
		})
	}
}

func TestDecoderRejectsUnpairedSurrogatesWithoutRepairingOpaqueIDs(t *testing.T) {
	for name, raw := range map[string]string{
		"high surrogate": strings.Replace(goodContent, `"G2"`, `"G\ud800"`, 1),
		"low surrogate":  strings.Replace(goodContent, `"G2"`, `"G\udc00"`, 1),
		"reversed pair":  strings.Replace(goodContent, `"G2"`, `"G\udc00\ud800"`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeContent([]byte(raw))
			if !errors.Is(err, domain.ErrMalformed) {
				t.Fatalf("silently repaired malformed Unicode: %#v, %v", got, err)
			}
		})
	}

	for name, tc := range map[string][2]string{
		"paired surrogate": {strings.Replace(goodContent, `"G2"`, `"G\ud83d\ude80"`, 1), "G🚀"},
		"unicode literal":  {strings.Replace(goodContent, `"G2"`, `"許可🚀"`, 1), "許可🚀"},
	} {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeContent([]byte(tc[0]))
			if err != nil || got.GrantID != tc[1] {
				t.Fatalf("lost exact valid identifier: %q, %v", got.GrantID, err)
			}
		})
	}
}

func TestContentDistinguishesAbsentParentFromPresentEmptyParent(t *testing.T) {
	absent := strings.Replace(goodContent, `"parent_grant_id":"G1",`, "", 1)
	if got, err := DecodeContent([]byte(absent)); err != nil || got.ParentGrantID != "" {
		t.Fatalf("root-shaped content must remain representable: %#v, %v", got, err)
	}
	presentEmpty := strings.Replace(goodContent, `"parent_grant_id":"G1"`, `"parent_grant_id":""`, 1)
	if got, err := DecodeContent([]byte(presentEmpty)); !errors.Is(err, domain.ErrMalformed) || !reflect.DeepEqual(got, domain.GrantContent{}) {
		t.Fatalf("accepted explicitly empty parent: %#v, %v", got, err)
	}
}

func TestValidateContentRejectsMalformedTypedValues(t *testing.T) {
	valid := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 1, ParentGrantID: "G1", Permissions: []string{"read"}, Scope: map[string]string{}}
	for name, edit := range map[string]func(*domain.GrantContent){
		"invalid UTF-8 grant ID": func(g *domain.GrantContent) { g.GrantID = string([]byte{0xff}) },
		"blank parent":           func(g *domain.GrantContent) { g.ParentGrantID = "  " },
		"invalid permission":     func(g *domain.GrantContent) { g.Permissions = []string{string([]byte{0xff})} },
		"invalid scope key":      func(g *domain.GrantContent) { g.Scope = map[string]string{string([]byte{0xff}): "x"} },
		"invalid scope value":    func(g *domain.GrantContent) { g.Scope = map[string]string{"dept": string([]byte{0xff})} },
	} {
		t.Run(name, func(t *testing.T) {
			g := valid
			edit(&g)
			if err := ValidateContent(g); !errors.Is(err, domain.ErrMalformed) {
				t.Fatalf("accepted malformed typed content: %#v, %v", g, err)
			}
		})
	}
}
func TestAssignmentRejectsMalformedAndIndependentProxyRecords(t *testing.T) {
	for name, raw := range map[string]string{
		"duplicate recipient":   strings.Replace(goodAssignment, `"id":"Team2"`, `"id":"Team2","id":"Team1"`, 1),
		"missing recipient":     strings.Replace(goodAssignment, `"recipient":{"type":"group","id":"Team2"},`, "", 1),
		"proxy":                 strings.Replace(goodAssignment, `"group"`, `"agent"`, 1),
		"unknown status":        strings.Replace(goodAssignment, `"enabled"`, `"orphan"`, 1),
		"no grant revision":     strings.Replace(goodAssignment, `"grant_revision":1,`, "", 1),
		"assignment validity":   strings.Replace(goodAssignment, `"version":"1"`, `"version":"1","validity":{"expires_at":"2026-09-30T00:00:00Z"}`, 1),
		"extra recipient field": strings.Replace(goodAssignment, `"id":"Team2"`, `"id":"Team2","owner":"maya"`, 1),
		"case alias":            strings.Replace(goodAssignment, `"id":"A2"`, `"ID":"A2"`, 1),
		"null status":           strings.Replace(goodAssignment, `"status":"enabled"`, `"status":null`, 1),
	} {
		t.Run(name, func(t *testing.T) {
			got, err := DecodeAssignment([]byte(raw))
			if err == nil {
				t.Fatal("accepted", raw)
			}
			if got != (domain.Assignment{}) {
				t.Fatal("returned partial assignment")
			}
		})
	}
}
func TestRoleAndValidityAndRootShapeArePreserved(t *testing.T) {
	raw := strings.Replace(goodContent, `"permissions":["hrms:employee:certificate::read"]`, `"role_id":"reader","role_revision":2`, 1)
	raw = strings.Replace(raw, `"scope":`, `"validity":{"not_before":"2026-09-01T00:00:00Z","expires_at":"2026-09-30T00:00:00Z"},"scope":`, 1)
	got, err := DecodeContent([]byte(raw))
	if err != nil || got.RoleID != "reader" || got.RoleRevision != 2 || got.Permissions != nil || got.Validity == nil || got.Validity.ExpiresAt == nil {
		t.Fatalf("%#v %v", got, err)
	}
	root := strings.Replace(goodContent, `"parent_grant_id":"G1",`, "", 1)
	got, err = DecodeContent([]byte(root))
	if err != nil || got.ParentGrantID != "" {
		t.Fatal("root shape is representable, not trusted", err)
	}
	empty := strings.Replace(goodContent, `{"cert":"C17"}`, "{}", 1)
	got, err = DecodeContent([]byte(empty))
	if err != nil || got.Scope == nil || len(got.Scope) != 0 {
		t.Fatal("lost explicit empty scope", err)
	}
}
func TestUnsupportedAndMalformedRemainDistinct(t *testing.T) {
	_, err := DecodeContent([]byte(strings.Replace(goodContent, `"version":"1"`, `"version":"2"`, 1)))
	if !errors.Is(err, domain.ErrUnsupported) {
		t.Fatal("future version", err)
	}
	_, err = DecodeContent([]byte("{"))
	if !errors.Is(err, domain.ErrMalformed) {
		t.Fatal("invalid JSON", err)
	}
}
func TestValidityCannotBeSilentlyDropped(t *testing.T) {
	for _, validity := range []string{"null", `{"expires_at":null}`, `{"expires_at":"not-a-date"}`, `{"extra":"ignored"}`, `{"not_before":"2026-10-01T00:00:00Z","expires_at":"2026-09-01T00:00:00Z"}`} {
		raw := strings.Replace(goodContent, `"scope":`, `"validity":`+validity+`,"scope":`, 1)
		if _, err := DecodeContent([]byte(raw)); err == nil {
			t.Fatal("accepted", validity)
		}
	}
}
func FuzzContentNeverReturnsPartialOnFailure(f *testing.F) {
	f.Add([]byte(goodContent))
	f.Add([]byte("null"))
	f.Add([]byte("{}"))
	f.Fuzz(func(t *testing.T, raw []byte) {
		got, err := DecodeContent(raw)
		if err != nil && !reflect.DeepEqual(got, domain.GrantContent{}) {
			t.Fatal("partial record returned")
		}
		if err == nil {
			encoded, e := json.Marshal(got)
			if e != nil {
				t.Fatal(e)
			}
			again, e := DecodeContent(encoded)
			if e != nil || !reflect.DeepEqual(got, again) {
				t.Fatal("unstable supported record", e)
			}
		}
	})
}
