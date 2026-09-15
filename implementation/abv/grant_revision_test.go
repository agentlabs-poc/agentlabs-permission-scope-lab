package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
	"agentlabs.local/abv/lab"
	"context"
	"reflect"
	"testing"
	"time"
)

type publicRevisionAdmin struct{}

func (publicRevisionAdmin) CheckAssignment(context.Context, abv.Evidence, domain.Identity, domain.Assignment, time.Time) error {
	return nil
}
func (publicRevisionAdmin) CheckGrantRevisionPublication(context.Context, abv.Evidence, domain.Identity, string, domain.GrantContent, time.Time) error {
	return nil
}

func TestFacadePublishGrantRevisionPersistsOnlyNewContent(t *testing.T) {
	area, _ := domain.NewArea("tenant-fin", "hrms")
	f := lab.TeamFINC17(area)
	g1 := f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}]
	g1.Permissions, g1.RoleID, g1.RoleRevision = nil, "fi9jvxobqsxs", 1
	f.Snapshot.Contents[domain.GrantKey{ID: "fk3x9r2m5iv8", Revision: 1}] = g1
	f.Snapshot.Controls["fk3x9r2man0d"] = domain.GrantControl{Version: "1", ID: "fk3x9r2man0d", Status: "disabled"}
	before := f.Snapshot
	path := t.TempDir() + "/authority.db"
	p, err := sqlite.CreateFixture(t.Context(), path, []storage.Snapshot{f.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	facade, err := abv.New(p, publicRevisionAdmin{}, clock{now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	candidate := domain.GrantContent{Version: "1", GrantID: "fk3x9r2man0d", Revision: 2, ParentGrantID: "fk3x9r2m5iv8", RoleID: "fi9jvxobqsxs", RoleRevision: 1, Scope: map[string]string{"cert": "C17", "dept": "ENG"}}
	got, err := facade.PublishGrantRevision(t.Context(), area, f.Issuer, "fm5b7t4p5iv8", candidate)
	if err != nil {
		t.Fatal(err)
	}
	if got.GrantID != "fk3x9r2man0d" || got.Revision != 2 {
		t.Fatalf("bad content: %#v", got)
	}
	if err = facade.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := sqlite.Open(t.Context(), path, allowAllRegistry{})
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	err = reopened.Read(t.Context(), area, func(after storage.Snapshot) error {
		if !reflect.DeepEqual(after.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}], before.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 1}]) ||
			!reflect.DeepEqual(after.Contents[domain.GrantKey{ID: "fk3x9r2man0d", Revision: 2}], candidate) ||
			!reflect.DeepEqual(after.Assignments, before.Assignments) || !reflect.DeepEqual(after.Controls, before.Controls) {
			t.Fatalf("publication changed established state: %#v", after)
		}
		if after.Assignments["fm5b7t4p5iv8"].GrantRevision != 1 || after.Controls["fk3x9r2man0d"].Status != "disabled" {
			t.Fatal("publication activated or adopted candidate")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// allowAllRegistry satisfies the port for tests whose subject is storage rather
// than the installation gate.
type allowAllRegistry struct{}

func (allowAllRegistry) ApplicationExists(context.Context, string) (bool, error) { return true, nil }
func (allowAllRegistry) Installed(context.Context, string, string) (bool, error) { return true, nil }
