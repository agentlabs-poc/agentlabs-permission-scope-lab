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

func TestRunWalksHTTPDemoWithoutListening(t *testing.T) {
	area, _ := domain.NewArea("acme", "hrms")
	fixture := lab.TeamFINC17(area)
	dbPath := filepath.Join(t.TempDir(), "authority.db")
	provider, err := storageSQLite.CreateFixture(t.Context(), dbPath, []storage.Snapshot{fixture.Snapshot})
	if err != nil {
		t.Fatal(err)
	}
	if err := provider.Close(); err != nil {
		t.Fatal(err)
	}
	var out, diag strings.Builder
	code := run([]string{"--db", dbPath, "--tenant", "acme", "--application", "hrms", "--human", "maya"}, &out, &diag)
	if code != 0 || diag.Len() != 0 {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), diag.String())
	}
	for _, want := range []string{
		"LAB ONLY: trusted flag context; no network listener or JWT authentication",
		"GET /api/v1/acme/FIN/C17 200",
		"GET /api/v1/acme/FIN/C18 404",
		"PUT /api/v1/acme/certificates/C17 200",
		`"title":"FIN walkthrough"`,
		"GET /api/v1/acme/departments/FIN/certificates 200",
		"GET /api/v1/acme/certificates 403",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("stdout %q missing %q", out.String(), want)
		}
	}
	if strings.Contains(out.String(), "ENG confidential") {
		t.Fatalf("walkthrough disclosed ENG title: %q", out.String())
	}
}

func TestRunRequiresTrustedContextFlags(t *testing.T) {
	var out, diag strings.Builder
	if code := run(nil, &out, &diag); code != 2 || out.Len() != 0 || !strings.Contains(diag.String(), "db, tenant, application and human are required") {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, out.String(), diag.String())
	}
}
