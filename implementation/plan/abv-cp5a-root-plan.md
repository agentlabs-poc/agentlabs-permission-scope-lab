# CP5-A Computed Root Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development. No reviewer dispatches per user; keep test-first verification.

**Goal:** Compute valid trusted roots' permission ceiling from their application catalog.

**Architecture:** Change the existing root-route construction, not grant storage.
The existing resolver still establishes trusted root/team status and eligibility.
Only roots compute catalog coverage; ordinary `SelectedPermissions` is unchanged.

**Tech Stack:** Go 1.25-compatible, existing tests; no dependencies.

**Spec:** [CP5-A](abv-cp5a-design.md); [Q-122/Q-123](../../docs/root-permission-evolution.md).

## Global Constraints

- No new wildcard/root wire encoding, root creation, revision or assignment writes.
- Preserve tenant/application context, scope, validity, enabled controls and actual lineage.
- Parent omission alone never establishes trusted-root status.
- Ordinary direct/role permission selections and explicit adoptions are unchanged.
- This slice handles catalog additions, not retirement operations or definition edits.
- No production Auth integration, PostgreSQL, deletion, reparenting, CLI or roles here.
- User requests no review passes. One 10-minute coding attempt and one 6-minute
  focused correction maximum; stop and reassess if the cap is reached.

### Task 1: Compute root route coverage

**Files:** modify `implementation/abv/internal/lineage/resolve.go`; create
`implementation/abv/internal/lineage/root_catalog_test.go`. Read `source.go`,
`resolve_test.go`, `validation/content.go`, `lab/cases.go` and `dependents.go`
to understand source reconstruction and the separate dependency inventory.

**Interfaces:** keep all public/exported signatures unchanged. `rootRoute`
continues producing `domain.Route`; `HasSource` reconstructs it through the same
resolver. Do not alter ordinary `validation.SelectedPermissions` or `Narrow`.

- [ ] **1. RED:** use `lab.TeamFINC17` and `ResolveParentTeam` with its G1 content
  and Team1 recipient, so the returned parent is G0, not narrowed G1. Add one
  active permission only to the catalog and require its presence in the root
  route without modifying stored G0. Independently expect sorted read/write/
  delete/export values. Example core call:

```go
root, err := lineage.ResolveParentTeam(f.Snapshot,
    f.Snapshot.Contents[domain.GrantKey{ID: "G1", Revision: 1}], "Team1", now)
if err != nil { t.Fatal(err) }
if !slices.Contains(root.Permissions, "hrms:payroll:payslip::export") {
    t.Fatal("trusted root did not reflect registered addition")
}
```

  Assert ordinary Team2 parent route remains G1's exact read/write. Include two
  tenant fixtures for the same app, separate-app catalog without export, malformed
  catalog key/definition mismatch, untrusted parentless G0, disabled grant,
  disabled assignment, expired validity, missing assignment, preserved nonempty
  root scope/validity, deterministic order, inactive extra permission excluded,
  and source revalidation through `HasSource` for a root holder. Deep-compare
  unchanged snapshot including validity values. Compatibility enabled must not
  silently cover new permissions incompatible with root scope: fail closed.
- [ ] **2. Run RED:** `go test ./internal/lineage -run RootCatalog -count=1`.
- [ ] **3. Implement:** after the existing trusted-root guard, derive active
  permission IDs, validate key/definition identity and existing permission grammar,
  sort, and use the derived list for root route permissions. Validate computed
  selection against the root's scope/compatibility using a temporary content copy
  with direct permissions and cleared role reference; never mutate stored content.
  All original shape, scope, validity and enablement checks remain. Reject an empty
  effective catalog through existing validation rather than produce authority.
  Do not reinterpret untrusted or ordinary grants as computed roots.

```go
// Use a value copy, not an edit of s.Contents:
computed := content
computed.Permissions = permissions // sorted active catalog IDs
computed.RoleID, computed.RoleRevision = "", 0
if err := validation.CheckContent(s.Area, s.Catalog, computed, s.Roles); err != nil {
    return domain.Route{}, err
}
```

- [ ] **4. GREEN:** focused command, `go test ./...`,
  `go test -race ./internal/lineage`, `go vet ./...`, `git diff --check`.
- [ ] **5. Commit exact files:** `feat(abv): compute trusted root catalog coverage`.
  Report RED/GREEN, SHA and limitations. No push, review agent or unrelated edits.

## Completion boundary

This delivers additive root computation only. Existing behavior can conservatively
reject a root whose stored permission reference is retired; complete retirement
semantics and APIs remain a separately scoped CP5 task. Do not claim them finished.
CLI end-to-end demonstration is the next unit.
