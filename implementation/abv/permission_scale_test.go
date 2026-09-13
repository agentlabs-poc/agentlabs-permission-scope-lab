//go:build !race

// Timing assertions are meaningless under the race detector, which is roughly
// fifteen times slower. This file measures; -race proves correctness elsewhere.

package abv_test

import (
	"agentlabs.local/abv/domain"
	"fmt"
	"testing"
	"time"
)

// TestListAtScale measures what an offset actually costs at a realistic catalog
// size, and where the time really goes.
func TestListAtScale(t *testing.T) {
	const n = 600
	permissions := make([]string, 0, n)
	domains := []string{"employee", "payroll", "leave", "recruitment", "benefits", "timesheet"}
	resources := []string{"certificate", "profile", "record", "request", "policy", "document", "entry", "summary", "report", "attachment"}
	verbs := []string{"read", "write", "approve", "delete", "export", "list", "submit", "reject", "archive", "restore"}
	for i := 0; i < n; i++ {
		permissions = append(permissions, fmt.Sprintf("hrms:%s:%s::%s",
			domains[i%len(domains)],
			resources[(i/len(domains))%len(resources)],
			verbs[(i/(len(domains)*len(resources)))%len(verbs)]))
	}
	f, app, id := seeded(t, true, permissions, nil)

	measure := func(label string, filter domain.PermissionFilter) domain.PermissionPage {
		t.Helper()
		const runs = 200
		var page domain.PermissionPage
		start := time.Now()
		for i := 0; i < runs; i++ {
			var err error
			if page, err = f.ListPermissions(t.Context(), app, id, filter); err != nil {
				t.Fatal(err)
			}
		}
		per := time.Since(start) / runs
		t.Logf("%-34s %6d returned  total=%d  %v", label, len(page.Permissions), page.Total, per.Round(time.Microsecond))
		return page
	}

	measure("offset 0, limit 50", domain.PermissionFilter{Limit: 50})
	measure("offset 300, limit 50", domain.PermissionFilter{Offset: 300, Limit: 50})
	measure("offset 550, limit 50", domain.PermissionFilter{Offset: 550, Limit: 50})
	measure("offset 599, limit 50", domain.PermissionFilter{Offset: 599, Limit: 50})
	measure("prefix hrms:payroll:", domain.PermissionFilter{Prefix: "hrms:payroll:", Limit: 50})
	measure("active only", domain.PermissionFilter{ActiveOnly: true, Limit: 50})
	measure("whole catalog, limit 500", domain.PermissionFilter{Limit: 500})

	start := time.Now()
	const gets = 200
	for i := 0; i < gets; i++ {
		if _, err := f.GetPermission(t.Context(), app, id, permissions[i%len(permissions)]); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%-34s %v", "get one by identifier", (time.Since(start) / gets).Round(time.Microsecond))
}
