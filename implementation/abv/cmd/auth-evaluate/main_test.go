package main

import (
	"agentlabs.local/abv/domain"
	"agentlabs.local/abv/internal/lab"
	"agentlabs.local/abv/internal/storage"
	storageSQLite "agentlabs.local/abv/internal/storage/sqlite"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunEvaluatesSQLiteAndRejectsMalformedMaterial(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	fixture.Snapshot.Assignments[fixture.Proposed.ID] = fixture.Proposed
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	base := []string{"--db", dbPath, "--tenant", "acme", "--application", "hrms", "--human", "nutan", "--permission", lab.PayslipRead}

	var stdout, stderr strings.Builder
	args := append(append([]string{}, base...), "--boundary", "dept=FIN", "--boundary", "cert=C17")
	if code := run(args, &stdout, &stderr); code != 0 || stderr.Len() != 0 || stdout.String() != `{"version":"1","decision":"allow","grant_ids":["G0","G1","G2"]}`+"\n" {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	for name, extra := range map[string][]string{
		"duplicate":          {"--boundary", "dept=FIN", "--all", "dept"},
		"malformed boundary": {"--boundary", "dept"},
		"malformed all":      {"--all", "dept=FIN"},
		"missing required":   {"--boundary", "dept=FIN", "--boundary", "cert=C17", "--human", ""},
	} {
		t.Run(name, func(t *testing.T) {
			var out, diag strings.Builder
			if code := run(append(append([]string{}, base...), extra...), &out, &diag); code != 2 || out.Len() != 0 || diag.Len() == 0 {
				t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), diag.String())
			}
		})
	}
}

func TestRunDenyAndMissingDatabaseHaveDistinctExitCodes(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	_ = provider.Close()
	args := []string{"--db", dbPath, "--tenant", "acme", "--application", "hrms", "--human", "nutan", "--permission", lab.PayslipRead, "--boundary", "dept=FIN", "--boundary", "cert=C17"}
	var out, diag strings.Builder
	if code := run(args, &out, &diag); code != 3 || !strings.Contains(out.String(), `"decision":"deny"`) || diag.Len() != 0 {
		t.Fatalf("deny code=%d stdout=%q stderr=%q", code, out.String(), diag.String())
	}
	missing := filepath.Join(t.TempDir(), "missing.db")
	args[1] = missing
	out.Reset()
	if code := run(args, &out, &diag); code != 4 || out.Len() != 0 || diag.Len() == 0 {
		t.Fatalf("error code=%d stdout=%q stderr=%q", code, out.String(), diag.String())
	}
}
