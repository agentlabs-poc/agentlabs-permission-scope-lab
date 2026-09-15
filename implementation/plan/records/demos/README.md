# Demonstrations

Captured CLI output, one file per concern. Every value is real — run against a
real SQLite store through the CLI, never by reading or writing the tables
directly. Only the layout is reconstructed, for legibility.

| File | Shows |
|---|---|
| `demo-16-ownership-record.svg` · `.md` | Ownership as authority to administer a team and nothing else: plural owners per team and per human, add-only and a remove that says whether it did anything, and a human on each side of the line Q-099 draws — an owner who is not a member, and a member who owns nothing |
| `demo-1-create-and-read.svg` | Registration is add-only, the identifier grammar is enforced, a typed read rejects unregistered identifiers and wildcards |
| `demo-2-retire-and-restore.svg` | Retirement is one reversible idempotent operation; a retired permission stays visible to administration and cannot be created by a status change |
| `demo-3-list-and-page.svg` | 603 permissions in one application — total, generation, whole-segment prefix, offset paging into the middle and past the end |
| `demo-4-bounds-and-cost.svg` | Page size capped, partial-segment prefixes rejected, an empty page rather than a fallback, and the measured cost of each operation |
| `demo-5-scope-register-and-read.svg` | Scope registration is add-only and rejects a wildcard anywhere in the key; a typed read returns the key and nothing else |
| `demo-6-scope-list.svg` | Scope listing ordered by key, offset paging, an empty page past the end, a capped page size |
| `demo-7-storage-table.svg` | The record store itself: permissions and scopes in one table, told apart by `key2`, the application in `key3` for both, and an empty scope payload |
| `demo-8-role-publish-and-get.svg` | The two publication paths side by side — the platform administrator shipping a role with `--application`, a tenant administrator composing one without it, a new revision of each, and the boundary enforced rather than merely recorded |
| `demo-9-role-list-revisions.svg` | One listing carrying both kinds each labelled, `--managed` narrowing to either, and `--latest` taking the highest revision of each across both |
| `demo-15-assignment-record.svg` · `.md` | The assignment keyed by the binding: both listing directions and the refusal to answer neither, publication not moving an adoption, an upgrade taking the latest and never the intermediate, delete refusing while a route rests on it, and the rows showing grant and recipient in the key path with the id in the value |
| `demo-14-grant-record.md` | The same run as the SVG beside it, rendered from the same capture — the commands, their output and their exit codes as text, with a table of what the run proves and what it cannot |
| `demo-14-grant-record.svg` | The grant as two record types in one table: the root told apart from its children, a child stating its permissions and narrowing where the root does neither, create issuing an id and writing revision 1 in the same act, four refusals that are four different answers, and delete refusing while anything depends on it |
| `demo-13-team-writes.svg` | The three team operations: create issuing an id, delete refusing while a child, a member or an assignment depends on the team, a cycle refused at the write, and membership writes that say whether they did anything |
| `demo-12-team-membership.svg` | Teams and membership as tenant-scoped L1 records: both directions of the roster question, ids rather than names, the parent as an id, and both tables folded away |
| `demo-11-envelope-drift-correction.svg` | The envelope before and after: `boundary`, `application_id` and `revision` removed, every record type sharing one shape, and `key3` holding the application in all of them — which is what let `application_id` go |
| `demo-10-role-storage-table.svg` | The role rows: an application role carrying no tenant beside tenant roles that do, the revision zero-padded in `key5`, the name in `key6`, and the drifted `revision` column left at `0` |

Every demonstration is **both** an SVG and a markdown file, generated from one
capture by one parse — so the image and the text cannot disagree. That was not
true until now: the four oldest had images and no text, because demo markdown
only started at the grant.

Demonstrations are captured, not described, and they earn their place: the
wildcard hole in scope registration was found by running demo 5, not by reading
the code.

Regenerate after any contract change. A demonstration that describes behaviour
the code no longer has is worse than none — these were rebuilt once already,
when offset paging replaced the original cursor.

The measurements in the fourth come from `permission_scale_test.go` and
`internal/storage/sqlite/catalog_scale_test.go`, which run at 600 permissions in
one application and assert the catalog read stays within budget. They are
excluded from `-race`, where instrumentation makes a timing assertion
meaningless.
