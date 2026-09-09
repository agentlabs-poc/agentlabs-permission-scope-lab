# Auth middleware local checkpoint

From `implementation/abv`, create the existing lab database (the command refuses
an existing path):

```sh
go run ./cmd/abv scenario seed team-fin-c17 --db /tmp/authority.db --tenant acme --app hrms
```

Evaluate Maya's existing A1 authority for FIN:

```sh
go run ./cmd/auth-evaluate --db /tmp/authority.db --tenant acme --application hrms --human maya --permission hrms:payroll:payslip::read --boundary dept=FIN
```

The same command with `--boundary dept=ENG` prints canonical deny JSON. The
program's exit codes are: allow 0, malformed flags 2, deny 3, read/evaluation/output
errors 4. `go run` reports nonzero program codes as `exit status N`; build the
command to observe those exit codes directly:

```sh
go build -o /tmp/auth-evaluate ./cmd/auth-evaluate
```
Use repeatable `--boundary key=value` for exact selections and `--all key` for
all-values selections.

These flags are trusted **test context**, not authentication or HTTP permission
selection. The adapter supports literal SQLite scope values only, reads a bounded
full snapshot, and has no JWT verification or production freshness mechanism.
The HTTP wrapper and handler integration remain pending; this checkpoint is only
the reusable core evaluator plus a local in-process SQLite test command.

To exercise Nutan's narrower Team2 grant through the existing protected lab flow:

```sh
go run ./cmd/abv assign --db /tmp/authority.db --tenant acme --app hrms --fixture-context maya-team1 --file testdata/a2.json
/tmp/auth-evaluate --db /tmp/authority.db --tenant acme --application hrms --human nutan --permission hrms:payroll:payslip::read --boundary dept=FIN --boundary cert=C17
```

Expected allow: `{"version":"1","decision":"allow","grant_ids":["G0","G1","G2"]}`.
Changing FIN to ENG or C17 to C18 denies. `--all dept` also denies this route.
Grant/assignment changes are observed on the next read; no allow cache is used.
These decisions do not prove certificate ownership: endpoint/handler enforcement
is the next integration step. Runtime `$self` is tested using resolved fixtures;
the current SQLite lineage adapter explicitly rejects it rather than guessing.
