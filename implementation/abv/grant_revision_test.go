package abv_test

import (
	"agentlabs.local/abv"
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	"agentlabs.local/abv/internal/storage/sqlite"
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
	g1 := f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}]
	g1.Permissions, g1.RoleID, g1.RoleRevision = nil, "payslip-reader", 1
	f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}] = g1
	f.Snapshot.Controls["G2"] = domain.GrantControl{Version: "1", ID: "G2", Status: "disabled"}
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
	candidate := domain.GrantContent{Version: "1", GrantID: "G2", Revision: 2, ParentGrantID: "G1", RoleID: "payslip-reader", RoleRevision: 1, Scope: map[string]string{"cert": "C17", "dept": "ENG"}}
	got, err := facade.PublishGrantRevision(t.Context(), area, f.Issuer, "A1", candidate)
	if err != nil {
		t.Fatal(err)
	}
	if got.GrantID != "G2" || got.Revision != 2 {
		t.Fatalf("bad content: %#v", got)
	}
	if err = facade.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := sqlite.Open(t.Context(), path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	err = reopened.Read(t.Context(), area, func(after storage.Snapshot) error {
		if !reflect.DeepEqual(after.Contents[domain.GrantKey{ID: "G2", Revision: 1}], before.Contents[domain.GrantKey{ID: "G2", Revision: 1}]) ||
			!reflect.DeepEqual(after.Contents[domain.GrantKey{ID: "G2", Revision: 2}], candidate) ||
			!reflect.DeepEqual(after.Assignments, before.Assignments) || !reflect.DeepEqual(after.Controls, before.Controls) {
			t.Fatalf("publication changed established state: %#v", after)
		}
		if after.Assignments["A1"].GrantRevision != 1 || after.Controls["G2"].Status != "disabled" {
			t.Fatal("publication activated or adopted candidate")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
