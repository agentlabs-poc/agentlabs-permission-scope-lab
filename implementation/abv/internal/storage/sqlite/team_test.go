package sqlite

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"testing"
)

// A tenant's teams are the tenant's, whichever application is being read. That
// is the whole point of dropping the application dimension, and it is the one
// behaviour this fold changes rather than preserves.
func TestATenantsTeamsAreVisibleInEveryApplication(t *testing.T) {
	hrms, _ := domain.NewArea("acme", "hrms")
	crm, _ := domain.NewArea("acme", "crm")
	other, _ := domain.NewArea("globex", "hrms")

	base := func(area domain.Area) storage.Snapshot {
		s := roleSeed(t)
		s.Area = area
		s.Catalog.ApplicationID = area.ApplicationID()
		s.Teams = map[string]domain.Team{
			"fibggi2juubk": {ID: "fibggi2juubk", Name: "fp8h2w6y5iv8"},
		}
		s.Memberships = []domain.Membership{{TeamID: "fibggi2juubk", HumanID: "fi7io4lvjqio"}}
		if area.TenantID() == "globex" {
			s.Teams = map[string]domain.Team{"fibggi2juxhc": {ID: "fibggi2juxhc", Name: "fp8h2w6yp2fs"}}
			s.Memberships = []domain.Membership{{TeamID: "fibggi2juxhc", HumanID: "fi7io4lvjwu8"}}
		}
		return s
	}
	opened, err := CreateFixture(t.Context(), t.TempDir()+"/teams.db", []storage.Snapshot{base(hrms), base(crm), base(other)})
	if err != nil {
		t.Fatal(err)
	}
	p := opened.(*provider)
	defer p.Close()

	// acme's team is visible in both of acme's applications.
	for _, area := range []domain.Area{hrms, crm} {
		if err := p.Read(t.Context(), area, func(s storage.Snapshot) error {
			if _, ok := s.Teams["fibggi2juubk"]; !ok {
				t.Fatalf("acme's team missing in %s: %v", area.ApplicationID(), s.Teams)
			}
			if len(s.Memberships) != 1 || s.Memberships[0].HumanID != "fi7io4lvjqio" {
				t.Fatalf("acme's memberships in %s: %v", area.ApplicationID(), s.Memberships)
			}
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	// Another tenant's team is not.
	if err := p.Read(t.Context(), other, func(s storage.Snapshot) error {
		if _, ok := s.Teams["fibggi2juubk"]; ok {
			t.Fatal("globex can see acme's team")
		}
		if _, ok := s.Teams["fibggi2juxhc"]; !ok {
			t.Fatalf("globex's own team missing: %v", s.Teams)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
