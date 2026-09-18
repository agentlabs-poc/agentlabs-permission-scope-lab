package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"errors"
	"testing"
	"time"
)

// A corrupt administrative record answers "could not check", not "not authorized"
// — Q-051 / DECISION-003, which is the rule the client applies to an answer it
// cannot read, and the property holds on this side too.
//
// It holds *upstream* rather than in Authorize, and that is the finding worth
// writing down. The client validates the route it was handed because it arrives
// over the wire from a source it does not trust; here content validation refuses
// the same shapes before a route is built at all, and the refusal propagates as
// ErrMalformed. Adding the client's route check to Authorize was tried and was
// dead code — three mutation checks against it all survived. So this test pins the
// behaviour rather than a second implementation of it: if the upstream refusal ever
// softens into a route-scoped skip, an administrator starts being told they lack
// authority they do have, and this fails.
func TestACorruptAdministrativeRecordCannotBeCheckedRatherThanDenied(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	maya := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjqio"}, HumanID: "fi7io4lvjqio"}
	priya := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)

	// The baseline: the same call on the untouched chain authorizes.
	sound, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	material := map[string]string{"team": "fibggi2juxhc"}
	if err := lineage.Authorize(t.Context(), sound, maya, lab.GroupWrite, material, now); err != nil {
		t.Fatalf("the sound chain did not authorize: %v", err)
	}

	// A wildcard in a scope value: a shape no writer will store and the one a
	// corrupted store could hold.
	bent, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	key := domain.GrantKey{ID: lab.TeamAdminGrant, Revision: 1}
	content := bent.Contents[key]
	content.Scope = map[string]string{"team": "*"}
	bent.Contents[key] = content
	// priya, because her only route is the one being corrupted. maya holds an
	// unscoped route beside it, and Q-051 is explicit that a route another route can
	// cover is not an outage — so this case has to be the human with nothing else.
	err = lineage.Authorize(t.Context(), bent, priya, lab.GroupWrite, material, now)
	if errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unreadable route answered as a denial: %v", err)
	}
	if !errors.Is(err, domain.ErrMalformed) {
		t.Fatalf("gave %v, want ErrMalformed", err)
	}
	// And it takes away only the authority that travelled through it. maya does not
	// hold the corrupted grant — she is not in the team it is assigned to — so her
	// unscoped route still authorizes over the same store. "Missing support stops the
	// affected authority route, not necessarily all authority."
	if err := lineage.Authorize(t.Context(), bent, maya, lab.GroupWrite, material, now); err != nil {
		t.Fatalf("one unreadable route took away another human's sound one: %v", err)
	}

	// A reserved token that is not $self is refused too, but *route-scoped*: the
	// route is dropped the way one with missing support is dropped, and the answer is
	// a denial rather than an error. The two corruptions therefore answer differently,
	// which is worth pinning rather than discovering. Whether a route-scoped skip
	// should itself be an evaluation error under Q-051 is a question about every
	// route, not about administrative ones — the business path answers it the same way
	// today — so it is recorded here and not decided here.
	reserved, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	content = reserved.Contents[key]
	content.Scope = map[string]string{"team": "$tenant"}
	reserved.Contents[key] = content
	if err := lineage.Authorize(t.Context(), reserved, priya, lab.GroupWrite, material, now); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a route dropped upstream gave %v, want ErrRejected from the older rule", err)
	}
}

// And the mirroring's other half still holds: a sound route that simply does not
// match is a refusal, not an error. Without this the check above could be met by
// making everything an error.
func TestARouteThatDoesNotMatchIsStillARefusal(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	chain, err := lab.AuthAdministration(area)
	if err != nil {
		t.Fatal(err)
	}
	priya := domain.Identity{Version: "1", Actor: domain.Actor{Type: "user", ID: "fi7io4lvjwu8"}, HumanID: "fi7io4lvjwu8"}
	now := time.Date(2026, 9, 15, 0, 0, 0, 0, time.UTC)
	err = lineage.Authorize(t.Context(), chain, priya, lab.GroupWrite, map[string]string{"team": "fibggi2juubk"}, now)
	if !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("a sound route that does not match gave %v, want ErrRejected", err)
	}
	var _ storage.Snapshot = chain
}
