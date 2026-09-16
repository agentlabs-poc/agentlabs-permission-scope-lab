package lineage_test

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lineage"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/lab"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestResolveHumanFindsOnlyDirectTeamHoldings(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	f.Snapshot.Assignments["fm5b7t4pan0d"] = f.Proposed

	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	want := []domain.Route{{
		Area: area, GrantID: "fk3x9r2m5iv8", Permissions: []string{lab.PayslipRead, lab.PayslipWrite},
		Predicates:    []domain.Predicate{{Key: "dept", Value: "FIN", SourceGrantID: "fk3x9r2m5iv8"}},
		AssignmentIDs: []string{"fm5b7t4p0dq3", "fm5b7t4p5iv8"},
	}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, %v; want %#v", got, err, want)
	}

	f.Snapshot.Memberships = []domain.Membership{{TeamID: "fibggi2juxhc", HumanID: "fi7io4lvjwu8"}}
	got, err = lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil || len(got) != 0 {
		t.Fatalf("nonmember got %#v, %v; want empty", got, err)
	}
}

func TestResolveHumanFiltersPermissionAndSortsByHoldingAssignment(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	f.Snapshot.Memberships = append(f.Snapshot.Memberships, domain.Membership{TeamID: "fibggi2jur5s", HumanID: "fi7io4lvjqio"})

	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil || len(got) != 2 || got[0].AssignmentIDs[len(got[0].AssignmentIDs)-1] != "fm5b7t4p0dq3" || got[1].AssignmentIDs[len(got[1].AssignmentIDs)-1] != "fm5b7t4p5iv8" {
		t.Fatalf("routes not deterministically sorted: %#v, %v", got, err)
	}
	got, err = lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipDelete, time.Time{})
	if err != nil || len(got) != 1 || got[0].GrantID != "fk3x9r2m0dq3" {
		t.Fatalf("permission selection got %#v, %v", got, err)
	}
	f.Snapshot.Memberships = []domain.Membership{{TeamID: "fibggi2juubk", HumanID: "fi7io4lvjqio"}}
	got, err = lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipDelete, time.Time{})
	if err != nil || len(got) != 0 {
		t.Fatalf("mismatched permission got %#v, %v; want empty", got, err)
	}
}

func TestResolveHumanSkipsOnlyInactiveRoutes(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	now := time.Date(2026, 9, 9, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		edit func(*lab.TeamFINC17Case)
	}{
		{"disabled holding", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}},
		{"disabled grant", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2m5iv8"]
			c.Status = "disabled"
			f.Snapshot.Controls["fk3x9r2m5iv8"] = c
		}},
		{"disabled ancestor", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p0dq3"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p0dq3"] = a
		}},
		{"expired grant", func(f *lab.TeamFINC17Case) {
			g := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
			g.Validity = &domain.Validity{ExpiresAt: &now}
			f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g
		}},
		{"missing parent support", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.Assignments, "fm5b7t4p0dq3") }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			tc.edit(&f)
			got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, now)
			if err != nil || len(got) != 0 {
				t.Fatalf("got %#v, %v; want inactive route omitted", got, err)
			}
		})
	}
}

func TestResolveHumanKeepsValidRouteWhenAnotherIsInactive(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	f.Snapshot.Memberships = append(f.Snapshot.Memberships, domain.Membership{TeamID: "fibggi2jur5s", HumanID: "fi7io4lvjqio"})
	a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
	a.Status = "disabled"
	f.Snapshot.Assignments["fm5b7t4p5iv8"] = a

	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil || len(got) != 1 || got[0].GrantID != "fk3x9r2m0dq3" {
		t.Fatalf("got %#v, %v; want surviving fk3x9r2m0dq3 route", got, err)
	}
}

func TestResolveHumanFailsClosedForInvalidEvidence(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	cases := []struct {
		name string
		edit func(*lab.TeamFINC17Case)
		want error
	}{
		{"invalid status", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "retired"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, domain.ErrRejected},
		{"cycle", func(f *lab.TeamFINC17Case) {
			f.Snapshot.Teams["fibggi2jur5s"] = domain.Team{ID: "fibggi2jur5s", Name: "fp8h2w6ykxan", ParentID: "fibggi2juubk"}
		}, domain.ErrRejected},
		{"duplicate support", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.ID = "fm5b7t4p5iv8-copy"
			f.Snapshot.Assignments[a.ID] = a
		}, domain.ErrRejected},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			tc.edit(&f)
			got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
			if !errors.Is(err, tc.want) || len(got) != 0 {
				t.Fatalf("got %#v, %v; want no routes and %v", got, err, tc.want)
			}
		})
	}
}

func TestResolveHumanRejectsInvalidRequestAndDirectUserAssignment(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	otherArea, _ := domain.NewArea("acme", "crm")
	wrongArea := f.Snapshot
	wrongArea.Area = otherArea
	// The walk is about the subject, so a subject with no id is refused and an
	// actor that differs from it is not — an application asking about someone
	// reaches the same routes that person would.
	noSubject := f.Issuer
	noSubject.HumanID = " "
	askedAbout := f.Issuer
	askedAbout.Actor = domain.Actor{Type: "service_account", ID: lab.WorkloadClient}

	for _, tc := range []struct {
		name       string
		s          storage.Snapshot
		identity   domain.Identity
		permission string
		want       error
	}{
		{"nil context", f.Snapshot, f.Issuer, lab.PayslipRead, domain.ErrMalformed},
		{"wrong area", wrongArea, f.Issuer, lab.PayslipRead, domain.ErrRejected},
		{"no subject", f.Snapshot, noSubject, lab.PayslipRead, domain.ErrMalformed},
		{"actor is not the subject", f.Snapshot, askedAbout, lab.PayslipRead, domain.ErrUnsupported},
		{"noncanonical permission", f.Snapshot, f.Issuer, "*", domain.ErrMalformed},
		{"unknown permission", f.Snapshot, f.Issuer, "hrms:missing::read", domain.ErrRejected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := t.Context()
			if tc.name == "nil context" {
				ctx = nil
			}
			got, err := lineage.ResolveHuman(ctx, tc.s, tc.identity, tc.permission, time.Time{})
			if !errors.Is(err, tc.want) || len(got) != 0 {
				t.Fatalf("got %#v, %v; want %v", got, err, tc.want)
			}
		})
	}

	// ResolveHuman still holds the actor to the subject, and this is the test
	// that says why: its caller has no gate, so this rule is the gate. The
	// loosened entry is ResolveAuthority, below.
	if got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, askedAbout, lab.PayslipRead, time.Time{}); !errors.Is(err, domain.ErrUnsupported) || len(got) != 0 {
		t.Fatalf("an actor that is not the subject resolved %#v, %v; want ErrUnsupported", got, err)
	}

	// A direct human assignment is not an eligible route — group-held support is
	// deliberate — and the answer must not say *whose* it is. Holding one to the
	// subject used to abort with ErrUnsupported while holding one to anybody else
	// was skipped, so 501-versus-200 told a caller whether a named human had one.
	// That is the enumeration the handler merges its refusals to prevent, and it
	// also took away every group route the human legitimately held.
	answers := map[string][2]any{}
	for name, recipient := range map[string]string{
		"held by the subject":  "fi7io4lvjqio",
		"held by someone else": "fi7io4lvkfsw",
	} {
		f.Snapshot.Assignments["fm5b7t4pzcp2"] = domain.Assignment{Version: "1", ID: "fm5b7t4pzcp2", GrantID: "fk3x9r2m5iv8", GrantRevision: 1, Recipient: domain.Recipient{Type: "user", ID: recipient}, Status: "enabled"}
		got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
		if err != nil || len(got) != 1 {
			t.Fatalf("%s: a direct assignment disturbed the group route: %#v, %v", name, got, err)
		}
		answers[name] = [2]any{len(got), err}
	}
	if answers["held by the subject"] != answers["held by someone else"] {
		t.Fatalf("the answer says who holds the direct assignment: %v", answers)
	}
}

func TestResolveTeamAssignmentDistinguishesOnlyInactiveEvidence(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	for _, tc := range []struct {
		name     string
		edit     func(*lab.TeamFINC17Case)
		inactive bool
	}{
		{"disabled assignment", func(f *lab.TeamFINC17Case) {
			a := f.Snapshot.Assignments["fm5b7t4p5iv8"]
			a.Status = "disabled"
			f.Snapshot.Assignments["fm5b7t4p5iv8"] = a
		}, true},
		{"disabled control", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2m5iv8"]
			c.Status = "disabled"
			f.Snapshot.Controls["fk3x9r2m5iv8"] = c
		}, true},
		{"missing parent support", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.Assignments, "fm5b7t4p0dq3") }, true},
		{"malformed control status", func(f *lab.TeamFINC17Case) {
			c := f.Snapshot.Controls["fk3x9r2m5iv8"]
			c.Status = "retired"
			f.Snapshot.Controls["fk3x9r2m5iv8"] = c
		}, false},
		{"missing adopted content", func(f *lab.TeamFINC17Case) {
			delete(f.Snapshot.Contents, domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1})
		}, false},
		{"untrusted root", func(f *lab.TeamFINC17Case) { delete(f.Snapshot.TrustedRoots, "fk3x9r2m0dq3") }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := lab.TeamFINC17(area)
			tc.edit(&f)
			_, err := lineage.ResolveTeamAssignment(f.Snapshot, "fm5b7t4p5iv8", time.Time{})
			if !errors.Is(err, domain.ErrRejected) || errors.Is(err, lineage.ErrInactive) != tc.inactive {
				t.Fatalf("got %v; rejected=%v inactive=%v, want inactive=%v", err, errors.Is(err, domain.ErrRejected), errors.Is(err, lineage.ErrInactive), tc.inactive)
			}
		})
	}
}

func TestResolveHumanBoundsDiscovery(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	f.Snapshot.Memberships = make([]domain.Membership, 10_001)
	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if !errors.Is(err, storage.ErrSnapshotLimit) || len(got) != 0 {
		t.Fatalf("got %#v, %v; want snapshot limit and no partial routes", got, err)
	}
}

func TestResolveHumanPreservesAdoptedRevisionValidityAndCancellation(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	expires := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	g := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g.Permissions, g.RoleID, g.RoleRevision = nil, "fi9jvxobqsxs", 1
	g.Validity = &domain.Validity{ExpiresAt: &expires}
	f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g

	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil || len(got) != 1 || !reflect.DeepEqual(got[0].Permissions, []string{lab.PayslipRead}) || len(got[0].Validities) != 1 || got[0].Validities[0].ExpiresAt == nil || !got[0].Validities[0].ExpiresAt.Equal(expires) {
		t.Fatalf("adopted revision changed: %#v, %v", got, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	got, err = lineage.ResolveHuman(ctx, f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if !errors.Is(err, context.Canceled) || len(got) != 0 {
		t.Fatalf("cancellation got %#v, %v", got, err)
	}
}

type cancelOnErrCall struct {
	context.Context
	calls, cancelAt int
}

func (c *cancelOnErrCall) Err() error {
	c.calls++
	if c.calls >= c.cancelAt {
		return context.Canceled
	}
	return nil
}

func TestResolveHumanCancellationWinsOverEmptySuccess(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	t.Run("inactive final candidate", func(t *testing.T) {
		f := lab.TeamFINC17(area)
		f.Snapshot.Memberships = []domain.Membership{{TeamID: "fibggi2juubk", HumanID: "fi7io4lvjqio"}}
		delete(f.Snapshot.Assignments, "fm5b7t4p0dq3")
		ctx := &cancelOnErrCall{Context: context.Background(), cancelAt: 4}
		got, err := lineage.ResolveHuman(ctx, f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
		if !errors.Is(err, context.Canceled) || len(got) != 0 {
			t.Fatalf("got %#v, %v; want cancellation and no routes", got, err)
		}
	})
	t.Run("empty discovery", func(t *testing.T) {
		f := lab.TeamFINC17(area)
		f.Snapshot.Memberships, f.Snapshot.Assignments = nil, nil
		ctx := &cancelOnErrCall{Context: context.Background(), cancelAt: 2}
		got, err := lineage.ResolveHuman(ctx, f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
		if !errors.Is(err, context.Canceled) || len(got) != 0 {
			t.Fatalf("got %#v, %v; want cancellation and no routes", got, err)
		}
	})
}

// The two entries are deliberately asymmetric, and the asymmetry is the whole
// security argument: ResolveHuman is reached through an adapter with no
// administration, so it keeps the acting-human rule. ResolveAuthority is reached
// through a service that runs CheckAuthorityRead first, so it may admit a caller
// who is not the subject.
//
// Making the shared walk permissive once removed the check from the path that
// had nothing else. This pins both halves so that cannot recur quietly.
func TestOnlyTheGatedEntryAdmitsAnActorThatIsNotTheSubject(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	asService := f.Issuer
	asService.Actor = domain.Actor{Type: "service_account", ID: lab.WorkloadClient}

	if _, err := lineage.ResolveHuman(t.Context(), f.Snapshot, asService, lab.PayslipRead, time.Time{}); !errors.Is(err, domain.ErrUnsupported) {
		t.Fatalf("the ungated entry admitted a service actor: %v", err)
	}
	resolved, err := lineage.ResolveAuthority(t.Context(), f.Snapshot, asService, domain.ResolveOptions{}, time.Time{})
	if err != nil {
		t.Fatalf("the gated entry refused a service actor: %v", err)
	}
	if len(resolved.ResolvedGrants) == 0 {
		t.Fatal("the gated entry resolved nothing for a subject who holds authority")
	}
}

// $self is not invalid evidence — it is the handbook's preferred shape. SELF-001:
// "An Employees group may receive one self-scoped payslip-read grant… The scope
// rule is shared; the resolved reach is specific to the human being authorized."
//
// So the walk carries the token through unresolved, and the human it stands for
// is decided at match time by whoever is asking. This case used to be listed
// among the invalid ones and answered ErrUnsupported.
func TestSelfTravelsTheChainUnresolved(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	g := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g.Scope = map[string]string{"user": "$self"}
	f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g

	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil || len(got) != 1 {
		t.Fatalf("a self-scoped grant resolved %#v, %v", got, err)
	}
	if len(got[0].Predicates) != 1 || got[0].Predicates[0].Key != "user" || got[0].Predicates[0].Value != "$self" {
		t.Fatalf("predicates = %#v, want the token carried unresolved", got[0].Predicates)
	}
}

// And it does not take the rest of the chain down with it. The old refusal fired
// at every step of the climb, so a self-scoped grant anywhere killed every grant
// below it — the grandchild here, which says nothing about self at all.
func TestSelfAboveDoesNotKillTheGrantsBelow(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	f := lab.TeamFINC17(area)
	root := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}]
	root.Permissions = []string{lab.PayslipRead, lab.PayslipWrite, lab.PayslipDelete}
	root.Scope = map[string]string{"user": "$self"}
	f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m0dq3", Revision: 1}] = root

	got, err := lineage.ResolveParentTeam(f.Snapshot, f.Child, "fibggi2juxhc", time.Time{})
	if err != nil {
		t.Fatalf("a self-scoped root refused the whole chain: %v", err)
	}
	keys := map[string]string{}
	for _, p := range got.Predicates {
		keys[p.Key] = p.Value
	}
	if keys["user"] != "$self" || keys["dept"] != "FIN" {
		t.Fatalf("predicates = %#v, want the token beside the narrowing below it", got.Predicates)
	}
}

// "Missing support stops the affected authority route, not necessarily all
// authority of that user or group" — authority-lineage.md:169.
//
// Only inactive support was skipped, so authority that no longer *holds* — a
// selected permission retired out of the catalog, a child that no longer narrows
// its parent — aborted the whole walk. One retirement took away every other
// grant the human had, and through localsource that reached the application as
// an outage rather than a denial: a routine lifecycle act paging an operator.
func TestOneIneligibleRouteDoesNotRemoveTheRest(t *testing.T) {
	area, err := domain.NewArea("acme", "hrms")
	if err != nil {
		t.Fatal(err)
	}
	// maya holds two independent routes: Team1's grant, and a second one hung
	// from the root and held by a team she is also in.
	second := func() lab.TeamFINC17Case {
		f := lab.TeamFINC17(area)
		f.Snapshot.Teams["fibggi2jv7qq"] = domain.Team{ID: "fibggi2jv7qq", Name: "fp8h2w6yv7qq", ParentID: "fibggi2jur5s"}
		f.Snapshot.Memberships = append(f.Snapshot.Memberships, domain.Membership{TeamID: "fibggi2jv7qq", HumanID: "fi7io4lvjqio"})
		f.Snapshot.Controls["fk3x9r2mv7qq"] = domain.GrantControl{Version: "1", ID: "fk3x9r2mv7qq", Status: "enabled"}
		f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2mv7qq", Revision: 1}] = domain.GrantContent{
			Version: "1", GrantID: "fk3x9r2mv7qq", Revision: 1, ParentGrantID: "fk3x9r2m0dq3",
			Permissions: []string{lab.PayslipRead}, Scope: map[string]string{"dept": "OPS"},
		}
		f.Snapshot.Assignments["fm5b7t4pv7qq"] = domain.Assignment{
			Version: "1", ID: "fm5b7t4pv7qq", GrantID: "fk3x9r2mv7qq", GrantRevision: 1,
			Recipient: domain.Recipient{Type: "group", ID: "fibggi2jv7qq"}, Status: "enabled",
		}
		return f
	}

	f := second()
	both, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil || len(both) != 2 {
		t.Fatalf("the fixture does not hold two routes, so nothing below means anything: %#v, %v", both, err)
	}

	// Retire the permission only Team1's grant selects. Its route stops; the
	// other one is untouched by the act and must survive it.
	f = second()
	write := f.Snapshot.Catalog.Permissions[lab.PayslipWrite]
	write.Active = false
	f.Snapshot.Catalog.Permissions[lab.PayslipWrite] = write
	got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{})
	if err != nil {
		t.Fatalf("retiring one permission failed the whole answer: %v", err)
	}
	if len(got) != 1 || got[0].GrantID != "fk3x9r2mv7qq" {
		t.Fatalf("the surviving route is not the unaffected one: %#v", got)
	}

	// And an integrity failure still fails closed. A cycle is not one route's
	// problem, and answering "she holds nothing" would be a lie in the
	// reassuring direction.
	f = second()
	f.Snapshot.Teams["fibggi2jur5s"] = domain.Team{ID: "fibggi2jur5s", Name: "fp8h2w6ykxan", ParentID: "fibggi2juubk"}
	if got, err := lineage.ResolveHuman(t.Context(), f.Snapshot, f.Issuer, lab.PayslipRead, time.Time{}); err == nil || len(got) != 0 {
		t.Fatalf("a cycle was answered rather than refused: %#v, %v", got, err)
	}
}
