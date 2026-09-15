package abv_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"testing"
)

type resolver interface {
	ResolveAuthority(context.Context, domain.Area, domain.FixtureContext, domain.Identity, domain.ResolveOptions) (domain.ResolvedAuthority, error)
}

func openResolveLab(t *testing.T) (resolver, domain.Area, domain.Identity) {
	t.Helper()
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "resolve.db")
	if err := (lab.Scenarios{}).Seed(context.Background(), area, "team-fin-c17", path); err != nil {
		t.Fatal(err)
	}
	api, closeConnection, err := lab.Connect(t.Context(), area, path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = closeConnection() })
	resolve, ok := api.(resolver)
	if !ok {
		t.Fatalf("lab application does not expose the authority read: %T", api)
	}
	return resolve, area, lab.TeamFINC17(area).Issuer
}

// The read the enforcement side consumes, and the first non-administrative one
// on the facade. Until it existed, the resolution that powers it was reachable
// only from inside this module — every exported method was administration.
func TestResolveAuthorityAnswersWhatAHumanHolds(t *testing.T) {
	api, area, maya := openResolveLab(t)

	resolved, err := api.ResolveAuthority(t.Context(), area, teamFixture, maya, domain.ResolveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Area != area || resolved.HumanID != maya.HumanID {
		t.Fatalf("answered about the wrong subject: %#v", resolved)
	}
	if len(resolved.ResolvedGrants) == 0 {
		t.Fatal("maya holds Team1's FIN authority and none was resolved")
	}
	grant := resolved.ResolvedGrants[0]

	// Effective, not stored. The scope is folded down the chain, so a client
	// never has to fold one itself.
	if !reflect.DeepEqual(grant.Scope, map[string]string{"dept": "FIN"}) {
		t.Fatalf("scope = %#v, want the folded boundary", grant.Scope)
	}
	if len(grant.Permissions) == 0 {
		t.Fatalf("resolved grant carries no permissions: %#v", grant)
	}

	// The explanation, which was computed and discarded until now: the chain is
	// ordered root-first and every step names its assignment and team.
	if len(grant.Source.Lineage) < 2 {
		t.Fatalf("lineage = %#v, want the chain back to the root", grant.Source.Lineage)
	}
	if !grant.Source.Lineage[0].Root {
		t.Fatalf("lineage is not root-first: %#v", grant.Source.Lineage)
	}
	for _, step := range grant.Source.Lineage {
		if step.GrantID == "" || step.AssignmentID == "" || step.TeamID == "" {
			t.Fatalf("incomplete lineage step: %#v", step)
		}
	}
	if grant.Source.AssignmentID == "" || grant.Source.TeamID == "" || grant.Source.Via != "membership" {
		t.Fatalf("source = %#v", grant.Source)
	}
	// The holding assignment is the last step of its own lineage.
	if last := grant.Source.Lineage[len(grant.Source.Lineage)-1]; last.AssignmentID != grant.Source.AssignmentID {
		t.Fatalf("lineage does not end at the holding assignment: %#v", grant.Source)
	}
}

// A gate deciding one request wants a filter; a menu wants everything. One call
// serves both, and the default is the complete answer.
func TestResolveAuthorityFiltersAndOmitsOnRequest(t *testing.T) {
	api, area, maya := openResolveLab(t)

	all, err := api.ResolveAuthority(t.Context(), area, teamFixture, maya, domain.ResolveOptions{})
	if err != nil {
		t.Fatal(err)
	}
	filtered, err := api.ResolveAuthority(t.Context(), area, teamFixture, maya,
		domain.ResolveOptions{Permissions: []string{lab.PayslipRead}})
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered.ResolvedGrants) == 0 || len(filtered.ResolvedGrants) > len(all.ResolvedGrants) {
		t.Fatalf("filter gave %d of %d", len(filtered.ResolvedGrants), len(all.ResolvedGrants))
	}
	for _, grant := range filtered.ResolvedGrants {
		found := false
		for _, permission := range grant.Permissions {
			if permission == lab.PayslipRead {
				found = true
			}
		}
		if !found {
			t.Fatalf("filtered grant does not carry the permission: %#v", grant)
		}
	}

	// A bearer token must not carry a lineage, which is what OmitSource is for.
	lean, err := api.ResolveAuthority(t.Context(), area, teamFixture, maya, domain.ResolveOptions{OmitSource: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(lean.ResolvedGrants) != len(all.ResolvedGrants) {
		t.Fatalf("omitting the explanation changed the answer: %d vs %d", len(lean.ResolvedGrants), len(all.ResolvedGrants))
	}
	for _, grant := range lean.ResolvedGrants {
		if grant.Source != nil {
			t.Fatalf("source survived OmitSource: %#v", grant.Source)
		}
		// The decision inputs are untouched — omitting the explanation must not
		// change what the grant authorizes.
		if len(grant.Permissions) == 0 || grant.Scope == nil {
			t.Fatalf("OmitSource dropped a decision input: %#v", grant)
		}
	}

	// Asking about a permission the catalog does not register is a caller
	// mistake. Answering "you hold nothing" would hide it.
	if _, err := api.ResolveAuthority(t.Context(), area, teamFixture, maya,
		domain.ResolveOptions{Permissions: []string{"hrms:payroll:payslip::export"}}); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("an unregistered filter gave %v, want ErrRejected", err)
	}
}

// Holding nothing is a completed answer, not a failure. It is the normal result
// for most humans and most permissions, and reading it as an error would make
// every unauthorised request look like an outage.
//
// It is reached here through a permission the human does not hold rather than
// through a second human, because the gate and the subject are still one
// identity — see ResolveAuthority's note.
func TestResolveAuthorityAnswersEmptyRatherThanFailing(t *testing.T) {
	api, area, maya := openResolveLab(t)

	resolved, err := api.ResolveAuthority(t.Context(), area, teamFixture, maya,
		domain.ResolveOptions{Permissions: []string{lab.PayslipDelete}})
	if err != nil {
		t.Fatalf("a permission the human does not hold gave %v, want a complete empty answer", err)
	}
	if len(resolved.ResolvedGrants) != 0 {
		t.Fatalf("maya resolved delete authority: %#v", resolved.ResolvedGrants)
	}
	// The envelope still answers about the right subject and area, so a client
	// can tell "nothing here" from "answered about someone else".
	if resolved.HumanID != maya.HumanID || resolved.Area != area {
		t.Fatalf("empty answer lost its subject: %#v", resolved)
	}
}
