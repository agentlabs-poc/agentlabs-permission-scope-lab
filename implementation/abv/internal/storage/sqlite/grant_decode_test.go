package sqlite

import (
	"agentlabs.local/abv/domain"
	"errors"
	"testing"
)

// A stored row that holds no scope is a malformed record, and saying so is the
// only way anyone finds out. This layer used to substitute {} on the way out,
// which is the same default the handbook forbids on the way in — "must not
// silently turn malformed or omitted scope into {}" — and it would have read
// every such row back as the widest child its parent allows.
//
// It is tested here rather than through the store because no operation can
// write such a row any more: the write path refuses it, which is exactly what
// makes the read path's substitution invisible and worth pinning.
func TestDecodeRevisionDoesNotInventAnEmptyScope(t *testing.T) {
	const id, revision = "fk3x9r2m5iv8", int64(1)
	for name, raw := range map[string]string{
		"scope omitted":     `{"parent_grant_id":"fk3x9r2m0dq3","permissions":["hrms:payroll:payslip::read"]}`,
		"scope null":        `{"parent_grant_id":"fk3x9r2m0dq3","permissions":["hrms:payroll:payslip::read"],"scope":null}`,
		"nothing but scope": `{}`,
	} {
		t.Run(name, func(t *testing.T) {
			content, err := decodeRevision(id, revision, []byte(raw))
			if !errors.Is(err, domain.ErrMalformed) {
				t.Fatalf("decoded %#v err=%v, want ErrMalformed", content, err)
			}
			if content.Scope != nil {
				t.Fatalf("a refused row still produced a scope: %#v", content.Scope)
			}
		})
	}

	// The control: an explicit empty scope is a real value and still decodes, so
	// this is a refusal to invent rather than a new restriction.
	content, err := decodeRevision(id, revision, []byte(`{"parent_grant_id":"fk3x9r2m0dq3","permissions":["hrms:payroll:payslip::read"],"scope":{}}`))
	if err != nil {
		t.Fatalf("an explicit empty scope was refused: %v", err)
	}
	if content.Scope == nil || len(content.Scope) != 0 {
		t.Fatalf("scope = %#v, want an empty map", content.Scope)
	}
}
