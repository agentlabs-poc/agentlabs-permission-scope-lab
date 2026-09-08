# CP3 — actual parent/team lineage

**Task 4 complete and independently approved at `4550860`.**
This describes the supported team route, not a production authorization result.
Source: [C01 working contract](../../../docs/authority-boundary-validation.md),
the [implementation design](../../plan/abv-design.md) and Task 4 of the
[execution plan](../../plan/abv-implementation-plan.md).

## What the resolver establishes

The proposed assignment is A2: grant G2, revision 1, to Team2. Its selected
content names G1 as its parent; Team2's recorded parent is Team1. Therefore the
resolver must find Team1's actual current G1 assignment, A1, and follow the
revision A1 adopted. G1 being stored, or being assigned to TeamX, is not that proof.

![Actual team/grant support and the separate acting-human source check](assets/team-lineage-resolution.svg)

The controlled test fixture makes the root premise explicit: RootTeam holds G0
through A0, G0 has an established-root entry, and Team1 is RootTeam's child.
This is a disposable lab fixture, not a canonical root format or production
bootstrap. A missing parent field is not evidence of root trust.

The resolver walks toward that root using the actual adopted contents, then
builds the effective route back down. For this example, G1 permits payslip
read/write in FIN. The coordinator can subsequently check that G2 selects read
and adds C17, producing FIN AND C17. A broad unrelated holder or newer unadopted
revision must not enlarge this route.

## Exact inputs and responsibility separation

`ResolveParentTeam` receives the partitioned snapshot, exact selected child
content, recipient team ID and evaluation time. The selected content is explicit
so the resolver does not silently choose a revision. It returns the supported
**parent route**, not an assignment receipt. Latest-only child creation is a
coordinator check; parent traversal uses actual adoptions.

`HasSource` separately establishes the acting human's access to the required
source route from explicit current membership and assignments. Maya's membership
in Team1 can supply G1; Nutan's membership in Team2 cannot substitute for it.
Membership in an ancestor team is not implicitly membership in Team1. Neither
membership nor possession establishes assignment-administration permission.

The supported team route remains based on team/grant support after issuance.
It does not acquire a permanent dependency on Maya being owner or member.
Task 5 still has to implement the administrative check, boundary checks and
atomic write together. These pure helpers cannot authorize storage by themselves.

## Invariants and rationale

| Invariant | Why it matters |
|---|---|
| Mandatory tenant/application and matching catalog | Matching IDs elsewhere cannot supply missing support. |
| Current assignment, exact adopted revision | Avoids importing another holder's broader or newer authority. |
| All ancestor grant and assignment controls | An enabled child cannot override a disabled parent. |
| All applicable validity windows | Enabled flags do not extend time-bound authority. |
| AND predicates with provenance, not map overwrite | Child ENG cannot replace inherited FIN. |
| Complete selected permissions must fit | A narrow grandchild cannot repair an overbroad parent. |
| Relevant cycle detection and traversal bound | Invalid graphs cannot authorize or consume unbounded work. |
| No union of unused historical edges | Obsolete revisions must not create fictitious support or cycles. |
| No issuer or ownership shortcut | Administrative ownership and business authority remain distinct. |

The 256-step traversal bound is an internal safety limit, not a canonical maximum team
depth. Exhaustion must return an evaluation error, never a truncated route.
Time is injected: start is inclusive and expiry exclusive. All returned mutable
route values must be copied so caller edits cannot change authoritative evidence.

## Unsupported first-slice cases

The initial source-access implementation supports the actual required team
holding through explicit human membership. Direct-human support selection that
needs the unresolved discovery rule, proxy issuance without delegation evidence,
and cross-recipient `$self` binding remain explicit unsupported cases. This does
not prohibit those models canonically or treat copying `$self` as valid proof.

Permission retirement/evolution operations and full lifecycle writes belong to
later checkpoints. This initial resolver requires fully active selected source
definitions; partial effective-permission propagation after retirement is not
implemented and must not be inferred from a successful active-catalog example.
Application actual-data enforcement still belongs to the
application endpoint; this component does not query HRMS to locate C17.

## Acceptance record

| Evidence | Current result |
|---|---|
| Root → Team1 gives read/write, FIN and `[A0, A1]` | Pass |
| Unrelated holder cannot substitute; exact adopted revision retained | Pass |
| Disabled/missing support, relevant cycles and selected-content mismatch | Pass |
| Trusted root cannot skip a non-root team's ceiling | Pass |
| Inclusive start, exclusive expiry, inherited validity and copy isolation | Pass |
| Exact adopted root role, no implicit latest-role expansion | Pass |
| Source membership, identity and complete reconstructed route checked | Pass, including duplicate final binding introduced after resolution |
| Direct/proxy/self unsupported cases stay explicit | Pass |
| SQLite fixture roundtrip and both tenant/application isolation dimensions | Pass |
| Full tests, race, vet, package build and staged whitespace checks | Pass |
| Independent review | Pass: both findings addressed; spec and quality approved at `4550860` |

The tests live in [the lineage package](../internal/lineage/resolve_test.go),
[source checks](../internal/lineage/source_test.go), and
[provider-backed fixture/isolation checks](../internal/lineage/provider_test.go).
The root route uses `validation.SelectedPermissions`, a narrow shared helper
over the existing exact direct/adopted-role logic. This prevents a second loose
permission-expansion interface from appearing at the root.

Initial review found that `HasSource` reconstructed the named final assignment
without rechecking uniqueness of that grant/team binding. A duplicate introduced
after route resolution could therefore pass this helper, even though parent-team
resolution rejects the same snapshot. The source check must re-establish the
unique binding, not treat a remembered assignment ID as sufficient evidence.
This is a consistency correction to the existing assignment rule, not a new
canonical source-selection policy.

Commit `4550860` corrects this by requiring a unique exact source binding before
route reconstruction. Its tests cover disabled duplicates introduced after
resolution and caller-crafted routes selecting enabled duplicates. A separate
regression establishes genuine source expiry: resolve immediately before expiry,
then reject `HasSource` at exact expiry. Independent re-review approved both
corrections. Full tests, race checks, vet, build and staged whitespace checks
pass on the corrected source. The SVG was rendered and visually checked.

CP3 remains incomplete until Tasks 5–7 integrate both gates, protected CLI
commands, transaction ordering and the restart/negative-case demonstrations.
