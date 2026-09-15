package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/lineage"
	"errors"
	"testing"
	"time"
)

func TestHasSourceRequiresActingHumanMembershipInActualHolder(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name string
		edit func(*lab.TeamFINC17Case)
		want error
	}{
		{"Maya is an explicit fp8h2w6y5iv8 member", func(*lab.TeamFINC17Case) {}, nil},
		{"Nutan cannot substitute for Maya", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Memberships = []domain.Membership{{TeamID: "fibggi2juxhc", HumanID: "fi7io4lvjwu8"}}
		}, domain.ErrRejected},
		{"ancestor membership cannot substitute", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Memberships = []domain.Membership{{TeamID: "fibggi2jur5s", HumanID: "fi7io4lvjqio"}}
		}, domain.ErrRejected},
		{"differently scoped holder cannot substitute", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fp8h2w6y4hu7"] = domain.Team{Name: "team",ID: "fp8h2w6y4hu7", ParentID: "fibggi2jur5s"}
			f.Snapshot.Memberships = []domain.Membership{{TeamID: "fp8h2w6y4hu7", HumanID: "fi7io4lvjqio"}}
		}, domain.ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			parent, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(&fixture)
			if err = lineage.HasSource(fixture.Snapshot, fixture.Issuer, parent, time.Time{}); !errors.Is(err, tc.want) {
				t.Fatalf("got %v; want %v", err, tc.want)
			}
		})
	}
}

func TestHasSourceRevalidatesRouteAndIdentity(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name string
		edit func(*lab.TeamFINC17Case, *domain.Route, *domain.Identity)
		want error
	}{
		{"caller-crafted route is not a ticket", func(_ *lab.TeamFINC17Case, route *domain.Route, _ *domain.Identity) {
			route.Permissions = append(route.Permissions, lab.PayslipDelete)
		}, domain.ErrRejected},
		{"disabled evidence is rechecked", func(f *lab.TeamFINC17Case, _ *domain.Route, _ *domain.Identity) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"unknown identity version", func(_ *lab.TeamFINC17Case, _ *domain.Route, identity *domain.Identity) {
			identity.Version = "2"
		}, domain.ErrUnsupported},
		{"service actor is not a human", func(_ *lab.TeamFINC17Case, _ *domain.Route, identity *domain.Identity) {
			identity.Actor.Type = "service"
		}, domain.ErrUnsupported},
		{"proxy issuance is unsettled", func(_ *lab.TeamFINC17Case, _ *domain.Route, identity *domain.Identity) {
			identity.Actor.ID = "proxy"
		}, domain.ErrUnsupported},
		{"missing human ID is malformed", func(_ *lab.TeamFINC17Case, _ *domain.Route, identity *domain.Identity) {
			identity.HumanID = ""
		}, domain.ErrMalformed},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			parent, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			identity := fixture.Issuer
			tc.edit(&fixture, &parent, &identity)
			if err = lineage.HasSource(fixture.Snapshot, identity, parent, time.Time{}); !errors.Is(err, tc.want) {
				t.Fatalf("got %v; want %v", err, tc.want)
			}
		})
	}
}

func TestHasSourceRejectsDuplicateFinalSourceBinding(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name       string
		status     string
		selectCopy bool
	}{
		{"duplicate added after route resolution", "disabled", false},
		{"caller-crafted route selects duplicate", "enabled", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := lab.TeamFINC17(area)
			parent, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
			if err != nil {
				t.Fatal(err)
			}
			copy := fixture.Snapshot.Assignments["fm5b7t4p5iv8"]
			copy.ID, copy.Status = "fm5b7t4p5iv8-copy", tc.status
			fixture.Snapshot.Assignments[copy.ID] = copy
			if tc.selectCopy {
				parent.AssignmentIDs[len(parent.AssignmentIDs)-1] = copy.ID
			}
			if err = lineage.HasSource(fixture.Snapshot, fixture.Issuer, parent, time.Time{}); !errors.Is(err, domain.ErrRejected) {
				t.Fatalf("duplicate final source binding passed: %v", err)
			}
		})
	}
}

func TestHasSourceRechecksExactExpiry(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	expiry := time.Date(2026, 9, 8, 1, 0, 0, 0, time.UTC)
	g1 := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g1.Validity = &domain.Validity{ExpiresAt: &expiry}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g1
	parent, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", expiry.Add(-time.Nanosecond))
	if err != nil {
		t.Fatal(err)
	}
	if err = lineage.HasSource(fixture.Snapshot, fixture.Issuer, parent, expiry); !errors.Is(err, domain.ErrRejected) {
		t.Fatalf("exact-expiry source evidence remained eligible: %v", err)
	}
}

func TestHasSourceLeavesDirectHumanAndSelfBindingExplicitlyUnsupported(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	parent, err := lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	direct := domain.Assignment{Version: "1", ID: "fm5b7t4pzcp2", GrantID: "fk3x9r2m5iv8", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: "fi7io4lvjqio"}, Status: "enabled"}
	directContent := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	directContent.Revision = 2
	directContent.Permissions = []string{lab.PayslipRead}
	directContent.Scope = map[string]string{"dept": "ENG"}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 2}] = directContent
	direct.GrantRevision = 2
	fixture.Snapshot.Assignments[direct.ID] = direct
	parent.AssignmentIDs[len(parent.AssignmentIDs)-1] = direct.ID
	parent.Permissions = []string{lab.PayslipRead}
	parent.Predicates = []domain.Predicate{{Key: "dept", Value: "ENG", SourceGrantID: "fk3x9r2m5iv8"}}
	if err := lineage.HasSource(fixture.Snapshot, fixture.Issuer, parent, time.Time{}); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("direct-human differing support was guessed: %v", err)
	}

	fixture = lab.TeamFINC17(area)
	parent, err = lineage.ResolveParentTeam(fixture.Snapshot, fixture.Child, "fibggi2juxhc", time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	g1 := fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g1.Scope = map[string]string{"user": "$self"}
	fixture.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g1
	if err := lineage.HasSource(fixture.Snapshot, fixture.Issuer, parent, time.Time{}); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("recipient-relative source was treated as literal equality: %v", err)
	}
}
