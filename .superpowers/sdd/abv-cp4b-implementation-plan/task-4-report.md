# CP4-B Task 4 report

Status: implementation complete; Task 4 and final independent review pending.

## Scope and rationale

Base: `e2f7c3d` (Tasks 1–3 reviewed). Added the optional
`AssignmentStatusAPI`, strict `assignment enable|disable` CLI dispatch, and a
marker/Area/fixture-bound lab adapter that embeds the existing grant-status
adapter. The CLI performs no pre-read and reuses the existing Assignment JSON
and typed-nil capability guard. The fixture, schema, canonical records, snapshot
limit and 256-step lineage bound are unchanged. Real Auth-service integration
is out of scope; the lab premise is prototype-only and not authentication.

## TDD evidence

RED:

```text
go test ./cli -run 'TestAssignmentStatus|TestInvalidCommand'
FAIL: assignment command returned exit 2 "unknown command".

go test ./cli ./internal/lab -run 'TestAssignmentStatus|TestInvalidCommand'
FAIL: undefined: NewAssignmentStatusAdministration.
```

GREEN/focused:

```text
go test ./cli ./internal/lab -run 'TestAssignmentStatus|TestInvalidCommand'
PASS

go test ./cli ./internal/lab ./cmd/abv
PASS
```

One initial full-suite run exposed a test setup error: the direct lab adapter
test exercised proposed A2 before inserting it into the snapshot. The setup was
corrected (production behavior was unchanged), then all final gates below were
rerun from the final tree.

## Final verification

Run from `implementation/abv/` unless stated otherwise:

```text
go test ./... -count=1                              PASS
go test -race ./... -count=1                        PASS
go vet ./...                                        PASS
go build -o /tmp/abv-cp4b ./cmd/abv                PASS
go mod verify                                       PASS (all modules verified)
rg import-boundary check over production packages   PASS (no forbidden match)
```

Run from repository root:

```text
npm run build                                       PASS
node --test tests/*.test.mjs                        PASS (10/10)
git diff --check                                    PASS
```

The compiled-process test preserves the original seed and records, then proves:
create A2; reject disabling A1 first; disable A2 then A1; reject enabling A2
first; enable A1 without cascading A2; enable A2; reopen exact A1/A2; inspect
unchanged G1/G2 content/control evidence; and restore successful diagnosis.

## Self-review and limits

Self-review found no task-scope defect after the final gates. Parsing rejects
missing context, unknown verbs, and revision/recipient flags before connection;
absent, pointer typed-nil and non-pointer typed-nil capabilities return status 5;
output failure returns status 4; connections close once; existing commands remain
covered. Administrative refusal and marker/context refusal leave assignments
unchanged. Unsupported dependency discovery remains an error and status writes
never cascade.

This remains a bounded local SQLite prototype. It does not provide production
authentication, a root operation, schema migration, PostgreSQL support or new
canonical JSON. Task 4 and full-slice independent review are pending; this report
does not pre-claim either verdict.
