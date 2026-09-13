# Demonstrations

Captured CLI output, one file per concern. Every value is real — run against a
real SQLite store through the CLI, never by reading or writing the tables
directly. Only the layout is reconstructed, for legibility.

| File | Shows |
|---|---|
| `demo-1-create-and-read.svg` | Registration is add-only, the identifier grammar is enforced, a typed read rejects unregistered identifiers and wildcards |
| `demo-2-retire-and-restore.svg` | Retirement is one reversible idempotent operation; a retired permission stays visible to administration and cannot be created by a status change |
| `demo-3-list-and-page.svg` | 603 permissions in one application — total, generation, whole-segment prefix, offset paging into the middle and past the end |
| `demo-4-bounds-and-cost.svg` | Page size capped, partial-segment prefixes rejected, an empty page rather than a fallback, and the measured cost of each operation |

Regenerate after any contract change. A demonstration that describes behaviour
the code no longer has is worse than none — these were rebuilt once already,
when offset paging replaced the original cursor.

The measurements in the fourth come from `permission_scale_test.go` and
`internal/storage/sqlite/catalog_scale_test.go`, which run at 600 permissions in
one application and assert the catalog read stays within budget. They are
excluded from `-race`, where instrumentation makes a timing assertion
meaningless.
