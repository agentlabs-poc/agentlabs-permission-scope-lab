# CP3 — protected assignment creation

**Task 5 implemented at `651bb2a`, corrected and independently approved at
`d7fea85`.**
This is the local prototype's transaction boundary, not production Auth
integration. The [design](../../plan/abv-design.md) and Task 5 in the
[execution plan](../../plan/abv-implementation-plan.md) govern this work.

## Two different checks, one write

Maya having G1's Finance read/write authority does not itself let her administer
Team2. Conversely, assignment-administration permission for Team2 does not give
her G1's business authority. Both checks must succeed for the exact proposed A2.

The [lineage resolver](lineage-resolution.md) establishes Team2's actual parent
holding and Maya's source access. The coordinator adds the administrative check,
proposal/latest-revision checks and persistence ordering.

![Component responsibilities and transactional provider boundary](../../plan/assets/abv-provider-architecture.svg)

```text
Exact A2 proposal + trusted identity + tenant/application
  → provider acquires write protection and reads consistent evidence
  → administrative check on an isolated copy of that evidence
  → ABV checks the original evidence and the complete proposal
  → recheck every applicable time window immediately before issuing writes
  → return only the exact validated assignment write set
  → provider commits
  → return receipt
```

The administrative adapter receives a separate copy because the provider
snapshot contains mutable maps, slices and time pointers. Editing a control or
membership in that copy must not manufacture the evidence that ABV later uses.
This is an internal defensive boundary, not a new canonical record. The copying
cost should be measured with the bounded-snapshot benchmarks before production.

## What must still be true at the write

- The explicit tenant/application is valid and matches the loaded installation.
- The exact proposed grant content exists and is latest for new assignment
  creation; no older-revision fallback or silent replacement is permitted.
- Required parent teams hold the actual adopted grants. A newer publication
  does not silently upgrade a parent's assignment.
- Every selected permission fits the complete parent route, and every scope
  restriction is retained with AND. An invalid parent is not repaired by a
  narrower grandchild.
- Required controls, source membership and time windows are valid. A disabled
  existing assignment still counts when checking duplicate grant/recipient pairs.
- The administrative check authorizes this human's operation for this recipient,
  independently of the business-source check.

The second clock read is the prototype's defined pre-write eligibility point.
Database locks do not freeze time. This does not claim an unreviewed production
commit-time expiry guarantee.

A diagnostic `check` cannot establish administrative permission or issue a
ticket for a later write. Creation must read and validate again under the write
transaction. A failed mandatory check produces no write set; a failed provider
commit produces no success receipt. Conflicts are not automatically replayed
against different evidence.

## Lab administration versus production administration

The lab scenario supplies a fixed, trusted **test premise** for Maya to perform
`auth:assignment::create` for Team2, with explicit administration-group membership.
The adapter must check its configured area, human, permission premise, recipient
and current test prerequisites. It is not an always-allow evaluator.

The Auth administrative permission is not registered into the HRMS business
catalog to make the fixture convenient. The adapter does not claim to have
resolved a real authenticated Auth namespace. Production identity, administrative
grant resolution and their transactional integration remain CP5 work.

Maya's Team1 source membership remains a separate check. Nutan's membership,
team ownership, arbitrary actor IDs or a remembered successful diagnostic cannot
substitute for either gate. Unresolved direct-human/proxy/self issuance cases
retain the explicit first-slice limits described in the lineage document.

## Acceptance record

Candidate `651bb2a` passes full Go tests, race checks, vet and build. Tests cover
baseline A2 save/reopen; independent admin/source failure; exact recipient,
context and revision; overbroad/unsupported proposals; snapshot-edit isolation;
disabled duplicate bindings; a stale diagnostic followed by withdrawal; two-handle
ordering/conflicts; cancellation without replay; expiry during validation; no
receipt on provider failure; and facade inspection/diagnosis.

The independent reviewer confirmed both gates, deep-copy completeness, actual
parent adoption, accumulated validity and receipt ordering. Review found a test
race: two buffered success signals could both be ready and `select` could report
incorrect ordering despite a legitimate commit followed by withdrawal. The fix
adds a receipt acknowledgement before withdrawal. Focused re-review confirms
the issue is addressed with no new breakage; the Task 5 gate is closed. It was
not an authorization bypass. Final acceptance will additionally isolate individual
snapshot-mutation cases and add parent-only expiry-crossing evidence.

Final source at `d7fea85` passes `go test ./... -count=1`,
`go test -race ./... -count=1`, `go vet ./...` and `go build ./...`.

The facade exposes a data-only `Evidence` alias so external administrative
adapters can name their input type. It exposes no provider or raw write method.
An external-package test checks the adapter seam; `OpenSQLite` provides the
SQLite-backed construction path. Internal relationship/definition inspection is
table-shaped, not a new canonical representation.

Working CLI commands and compiled end-to-end workflows remain Task 6; the
first-slice scenario matrix and benchmarks remain Task 7. This page does not
mark CP3 or the whole ABV implementation complete.
