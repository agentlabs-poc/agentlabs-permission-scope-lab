# Demonstrations

Captured output, one file per concern. Every value is real — run against a real
SQLite store through the CLI, or through a running service over HTTP, never by
reading or writing the tables directly. Only the layout is reconstructed, for
legibility.

| File | Shows |
|---|---|
| `demo-21-two-processes.svg` · `.md` | **The first capture with two processes.** Ordinary HTTP requests reach an application's four endpoints, the gate on the client decides, and the handler body runs or does not — 200 inside the boundary, 403 outside it, and an all-values listing refused rather than narrowed (Q-071). One woman reads a certificate and cannot write it while the other writes it through the same endpoint and the same policy, which is the selected-versus-inherited asymmetry arriving at the HTTP layer. Every question that crossed the wire is printed in full: each names a human and nothing else — no permission, endpoint, method, resource or verdict. What the application links is `authmiddleware` and `authclient` and not `abv`; and with Auth stopped the answer is 503, never 403 |
| `demo-20-a-decision.svg` · `.md` | **The first capture of a decision.** A human asks, and the answer comes back with the chain of grants that authorized it — root first. The same request denied one value outside its boundary; a deeper route needing every predicate its chain accumulated; write allowed for one human and denied for another through the same permission, which is the selected-versus-inherited asymmetry arriving at a decision; an all-values selection denied rather than narrowed (Q-071); and one self-scoped grant to a group reaching a different person for each member who asks — the handbook's preferred shape, working for the first time |
| `demo-19-service-credential.svg` · `.md` | The caller need not be the subject: an application's own credential resolving a human it is not, the same credential returning that human's own answer unchanged, and two refusals — a credential this deployment never issued, and a human the fixture gate does not admit. The impersonation rule is named and not shown, because `--human` supplies both the actor and the subject and no command line can construct one |
| `demo-18-resolve.svg` · `.md` | The read an enforcing client consumes: one human's entitlements as the canonical document, two routes at different lineage depths with the deeper one's scope folded to `dept=FIN` AND `cert=C17`, the same call narrowed to one permission, the form without the explanation that a bearer token would carry, and three permissions the catalog does not supply — never registered, registered then retired, registered and simply not held — answered identically, because a filter is a narrowing and not an assertion |
| `demo-17-root-establishment.svg` · `.md` | Where a lineage begins: a tenant with teams and not one grant, an establishment refused before any permission is registered, the holder rules, the three records written in one transaction, and the same child proposal resolving under the Auth root and refused under the application root — the catalog listing beside it holding every permission both were checked against |
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
the code, and demo 20 is why we found that `$self` had never been implemented past the chain walk.

Regenerate after any contract change. A demonstration that describes behaviour
the code no longer has is worse than none — these were rebuilt once already,
when offset paging replaced the original cursor.

The measurements in the fourth come from `permission_scale_test.go` and
`internal/storage/sqlite/catalog_scale_test.go`, which run at 600 permissions in
one application and assert the catalog read stays within budget. They are
excluded from `-race`, where instrumentation makes a timing assertion
meaningless.
