# ABV first-slice acceptance evidence

## Current delivery

CP1 and CP2 are delivered. CP3 has reviewed parent-team lineage and protected
assignment creation, published through `1786b22`. The working CLI is verified
and independently approved through `5815abf`. Full first-slice source-case and
benchmark acceptance remain Task 7; this is not full ABV or production Auth
completion.

- [CP2 provider evidence](sqlite-provider.md)
- [CP3 lineage evidence](lineage-resolution.md)
- [CP3 protected assignment evidence](assignment-creation.md)
- [Verified CLI walkthrough](local-testing.md)

## Historical CP1 acceptance evidence

**Current state: CP1 implemented, verified and independently reviewed.**
This page preserves CP1's evidence and limits at delivery, when no SQLite or
production ABV tests existed. SQLite persistence is now independently verified
in [CP2 evidence](sqlite-provider.md); actual lineage and protected assignment
are covered by the CP3 links above. Production Auth remains later work.
Statements below about missing persistence or resolution describe CP1.

## Scope

Task 1: mandatory outer context, core record representations, internal errors,
reusable application interface and no-write CLI shell.

Task 2: strict supported grant/assignment JSON, registered definitions and exact
role selection, permission subsets and accumulated scope/validity restrictions.

## Test-first evidence

- `go test ./domain ./cli -count=1` first failed because `Area`, core records,
  errors and `Run` were absent. After implementation both packages passed.
- `go test ./internal/codec ./internal/validation -count=1` first failed because
  decoders, `CheckContent` and `Narrow` were absent. Implemented those interfaces
  and checked their behavior with positive and negative fixtures.
- An escaped-key fixture initially double-escaped its key, making it a distinct
  literal rather than a duplicate. Corrected the test input after observing the
  exact decoded value; the real escaped duplicate is rejected.
- Initial full `go test ./... -count=1`, race suite and vet passed before the
  subagent hardening pass. The post-review evidence is recorded below.

## Hardening and verification — 8 September 2026

The sol-medium coding subagent added failing tests exposing unpaired Unicode
surrogate repair, present-empty parent ambiguity, invalid UTF-8 in typed content,
corrupt selected definition metadata and malformed CLI grammar. It then fixed
those behaviors. Valid Unicode remains exact; no new authority fields were added.
The human actor fixture was corrected to canonical `user`, not `human`.

The subagent reported a bounded three-second decoder fuzz run with two workers:
8,142 executions, no failure. This is bounded fuzz evidence, not exhaustive proof.

The coordinator independently ran these commands on the finished code:

```text
go test ./... -count=1          PASS: 22 named tests plus the fuzz seed corpus
go test -race ./... -count=1    PASS
go vet ./...                   PASS, no diagnostics
go build ./...                 PASS (packages, not a CLI executable yet)
git diff --check               PASS
```

Existing website regression suite: all 10 tests passed; `npm run build` passed.
The checkpoint has no external Go dependencies, SQLite driver, server or raw
write method. CLI currently admits grant/assignment inspection syntax; wider
inspection kinds from Task 6 will be implemented with their adapters at CP3.

## Independent review — correction before integration

The reviewer examined checkpoint `5751b62` and found an important role-path gap:
`CheckContent` could validate a child selecting a read-only role, while the old
`Narrow` interface accepted a separately supplied write permission from a broader
parent. A comment requiring correct expansion was not a sufficient API invariant.
The code has not been deployed or integrated with any persistence path.

The correction removes that expansion parameter. Both helpers derive direct or
exact-role-revision permissions through a shared implementation, and narrowing
must retain the complete selected set or reject it. It cannot substitute a
different parent permission or silently trim a role. Regression tests and a
scoped independent re-review were completed before CP1 completion.

The reviewer also requested an explicit cancellation-category choice. Standard
`context.Canceled` and `context.DeadlineExceeded` are retained for classification
with `errors.Is`, rather than inventing a canonical public error code.

**Re-review result at `6a192f4`:** both findings addressed; no new CP1 regression;
spec compliant, code quality approved and ready to merge. The coordinator also
reran full tests (25 named tests plus fuzz seeds), race detection, vet and package
build on the corrected code, all passing. The initial 22-test result above is
retained as evidence of the earlier checkpoint, not the final test count.

## Limits and source-case coverage

Source-case coverage is deliberately limited:

| Source case | CP1 evidence | Remaining proof |
|---|---|---|
| C01-T05: permission expansion | `TestNarrowRejectsPermissionAndOuterBoundaryEscape` rejects a permission outside the parent. | Actual source discovery and no-write transaction test in CP3. |
| C01-T06: child `{}` | `TestNarrowKeepsAllRestrictionsAndCopiesInputs` retains FIN. | End-to-end supported assignment in CP3. |
| C01-T07: FIN AND ENG | The same narrowing test preserves both predicates. | No new contradictory-scope publication policy is chosen. |
| C01-T21: cross-tenant support | Unit checks reject mismatched tenant or application on a supplied route. | Provider lookups, installation checks and actual lineage isolation in CP2/CP3. |
| C01-T22: validity | Time limits are decoded and deep-copied into narrowed routes. | Current-time eligibility at resolution and mutation. |
| C01-T28: cross-recipient self | Registration may recognize `$self`; this is not binding proof. | Still unresolved; no supported issuance claim. |

All other source cases remain later-checkpoint work; C01-T27 is application
actual-data enforcement, not an ABV success criterion. See the
[source matrix](../../../docs/authority-boundary-validation.md).

They do not establish actual parent-team support, source possession, authenticated
administration, provider isolation, latest revision selection, current time
eligibility, atomic save or application actual-data enforcement. Those belong to
later checkpoints. No whole C01 scenario matrix is claimed passing at CP1.
