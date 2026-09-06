# Authorization principles — consolidated approved rules

## Current qualifications and additions — RECON-002

These entries consolidate Q-117 and Q-121–Q-130 without adding new policy.
PC-01 applies to tenant operations; Q-121 distinguishes platform-management
authority outside tenant business scope. PC-14/20/26 remain subject to computed
root coverage, not blanket root revision/adoption for catalog additions. PC-22's
timing is refined by Q-128/Q-129; PC-24/25 include Q-130's complete per-item routes.

| Ref | Requirement and reason | Compliant example | Counterexample | Approved source |
|---|---|---|---|---|
| PC-29 | Authorize capability publication through application platform-management authority, separate from tenant business access. | Authorized HRMS publisher introduces a permission without tenant payslip access. | Tenant payslip-delete alone authorizes catalog publication, or publication grants tenant data access. | [Q-121](application-platform-authority.md) |
| PC-30 | Compute legitimate root permissions from one shared current application catalog; keep tenant lifecycle isolated. | New HRMS delete capability reaches valid roots; ordinary read-only children remain read-only. | Auto-rewrite child grants, broaden scope, or introduce tenant release pins. | [Q-122/123](root-permission-evolution.md) |
| PC-31 | Expose initial authority only after coherent setup; authorized retries retain validated intent. | Resume Maya's incomplete setup, then expose access only after full establishment. | Replace Maya with Nutan silently or expose partial root access. | [Q-117/124](bootstrap-initial-assignment.md) |
| PC-32 | Authorized retirement withdraws the permission despite retained grant references. | Root no longer supplies retired delete; the old grant remains stored without usable delete. | An enabled old assignment keeps supplying retired delete. | [Q-125](permission-lifecycle.md) |
| PC-33 | Preserve permission identifier meaning, including after retirement. | Use a new pay identifier for payment execution. | Redefine approve to also transfer money under existing grants. | [Q-126](permission-lifecycle.md) |
| PC-34 | V1 delegation is directly human-to-proxy, not proxy-to-proxy. | A and B each have authorized human-linked delegations. | A supplies authority to B through a hidden intermediate delegation. | [Q-127](delegation-lifecycle.md) |
| PC-35 | All confirmed reductions govern new checks without stale-authority grace. | New checks cannot use removed Finance membership. | Cached membership survives confirmation because its TTL has time left. | [Q-128](authority-freshness.md) |
| PC-36 | Prior-allowed bounded synchronous application work may finish, preserving Q-074/Q-110 limits. | The same previously allowed read completes without changing its evaluated reach. | Reuse the allow for a retry, a changed boundary, or a conflicting Auth authority write. | [Q-129](concurrent-enforcement.md) |
| PC-37 | Every batch item needs a complete valid route; different items may use different routes. | FIN-delete and ENG-delete separately cover their items before effects. | Borrow FIN-delete permission and ENG-read scope to delete in ENG. | [Q-130](bulk-enforcement.md) |

There are 37 required catalog entries and one approved preference. This expansion
records later approvals, not additional checkpoint credit or implementation proof.
The earlier root-growth qualification is preserved immediately below as history.

<details>
<summary>Previous Q-120A qualification — computation is now selected under Q-122</summary>

**Later qualification — Q-120A:** [root coverage grows automatically](root-permission-evolution.md)
on legitimate application capability upgrades without a separate manual root
update/adoption gate. This qualifies PC-14/20/26's blanket root/catalog-growth
reading, not ordinary child permission selection, tenant/application boundaries,
status, or immutable-record integrity. The root representation/revision mechanism
is open; no live wildcard syntax is adopted. Earlier principle entries are
preserved below with this explicitly linked root-specific qualification.

</details>

This is the numbered catalog required by HC-02-05. It consolidates existing
approvals; it does not adopt new behavior or finish the associated wire schemas.
PC numbers are reading references. The linked Q/decision references remain the
authority for each rule and its history.

**Required** means an approved constraint on a conforming implementation.
**Guidance** means an approved preference, not a reason to deny an otherwise
valid request. The examples below assume other mandatory checks pass. A compliant
example illustrates the stated rule; it is not evidence of deployed enforcement.

## Required principles

| Ref | Requirement and reason | Compliant example | Counterexample | Approved source |
|---|---|---|---|---|
| PC-01 | Keep the trusted tenant as an implicit outer boundary. Local scope cannot remove isolation. | Root scope `{}` remains inside the established tenant. | Use `{}` to read another tenant, or trust a path tenant without binding it to context. | TENANT-001; [scope](scope-model.md) |
| PC-02 | A permission names an operation; scope selects its reach. Preserve both together. | FIN-read and FIN-write are separately authorized capabilities. | Combine tenant-wide read scope with FIN-write permission to write anywhere. | PERMISSION-001 / DECISION-001; [permission](permission-model.md) |
| PC-03 | Use registered permission identifiers with the agreed namespace convention. Prefixes, aliases, and wildcards confer no implicit v1 authority. | Explicitly list both ledger-read and entry-read when both are needed. | Treat ledger-read as authorization for every later child name. | Q-056–Q-059; [permission](permission-model.md) |
| PC-04 | Scope is a required flat string-value object; its keys combine with AND. Alternatives use separate complete grants. | FIN AND C17 restricts the same operation's data by both boundaries. | Use an array for two departments, or interpret keys as OR. | SCOPE-007/008; [scope](scope-model.md) |
| PC-05 | Child permissions stay within parent effective permissions; scope accumulates by AND. Equality is permitted. | Parent FIN read/write, child read plus C17; child `{}` retains FIN. | Replace FIN with ENG, remove inherited scope, or invent child delete. | Q-095 / Q-101; [lineage](authority-lineage.md) |
| PC-06 | Every child team's authority stays within its parent team's authority. Team lineage is not human membership inheritance. | Explicitly assign a supported child grant to Team2 and explicitly add Nutan. | Give Team2 a separate broad grant to bypass Team1, or infer Nutan's Team1 membership. | Q-091 / Q-095; [subgroups](subgroups.md) |
| PC-07 | Keep grant definitions recipient-free and assignments recipient-bearing. Definitions alone give no access. | A valid assignment adopts G1 revision 2 for Team1. | Authorize a user merely because G1 exists in the registry. | Q-090 / Q-107; [grant contracts](grant-revision-format.md) |
| PC-08 | Auth owns authorization memberships; groups contain humans only. Business grouping does not silently supply membership. | Explicit Auth membership supplies a group's valid route. | Admit an agent as a first-class member or infer Auth membership from an application's department. | Q-045 / Q-091; [groups](groups-and-membership.md) |
| PC-09 | Every agent/service account uses trusted human-dependent authority within delegation limits. `$self` remains that human. | Agent A-17 uses Vinay's supported read subset. | Use a caller-supplied Maya ID to switch grants, or keep independent access after required human support disappears. | AUTHORITY-002 / Q-085; [attribution](proxy-attribution.md) |
| PC-10 | Preserve actual actor and authorizing human distinctly. In the approved JWT mapping, human `sub` compatibility cannot bypass proxy checks. | `actor` identifies A-17, `human_id` and `sub` identify Vinay, with verified association. | A legacy consumer uses only Vinay's subject to permit an agent's prohibited delete. | Q-086 / Q-087; [identity](identity-context.md), [JWT](jwt-identity-mapping.md) |
| PC-11 | Each protected endpoint declares exactly one permission and its selected input sources. Declaration is server-owned and authorization-first. | A PUT declares its permission and the required body input binding. | The caller chooses the permission, or a missing body input silently falls back to a path value. | Q-049 / Q-050-B/E/F; [endpoint](endpoint-authorization.md) |
| PC-12 | Use one endpoint-owned authorization gate; no prepared handoff authorizes unfinished work. | Establish sufficient authority/material and enforce the decision before protected output/effects. | Middleware returns provisional allow and the handler treats incomplete checking as permission. | CONTRACT-006; [system](system-overview.md) |
| PC-13 | Application integration establishes or enforces scope relationships on actual output/effects. Matching inputs alone is not proof. | The C17 lookup is constrained by the trusted tenant, FIN, and the requested ID. | Log FIN, then fetch and return C17 by unchecked ID alone. | Q-050-C/D; [endpoint policy](endpoint-policy-format.md) |
| PC-14 | Register applicable permissions/scope contracts before accepting grants. Auth validates abstract contracts; the application owns domain meanings. | Include Auth administration registration before bootstrap grants. | Bootstrap accepts an unsupported permission, or Auth guesses certificate ownership from a key name. | Q-039 / Q-114; [registration](application-registration.md) |
| PC-15 | Declare the application's compatibility-validation mode upfront. Enabled validation applies to all its grants; disabled does not skip other validation. | Validate the existing population before activating the enabled mode under the agreed safeguard. | Treat absent metadata as permission to skip a declared enabled check. | Q-040–Q-042; [registration](application-registration.md) |
| PC-16 | Auth grant administration needs both administrative permission/scope and valid source-boundary support. Neither substitutes for the other. | Maya can assign to Team2 and derives its authority within valid FIN-read support. | Her assignment-create grant alone permits distributing ENG-write. | Q-093 / Q-100; [Auth gate](auth-service-authority-gate.md) |
| PC-17 | Preserve actual continuing support, not an invented permanent issuer dependency. Ownership changes do not import personal authority. | Team-held support continues when Om replaces Maya and all required links remain valid. | Automatically rebind children to Om's broader personal grant. | Q-099; [ownership](ownership-lineage.md) |
| PC-18 | Separate grant-wide status, assignment status, and effective lineage validity. Missing required support makes that route and its dependents ineffective. | Disable G2: G3 loses effective support; its stored enabled flag need not change. | An enabled assignment overrides disabled G2, or a timeout proves orphanhood. | Q-094 / Q-101; [bindings](parent-grant-bindings.md) |
| PC-19 | Guard structural changes using affected bindings, bottom-up; explicit re-enablement checks current reality. Reject grant/team ancestor cycles even when disabled. | Disable/remove relevant dependent assignments before changing the parent, then validate before enabling. | Change a live bound parent silently, or permit a cycle because its nodes are disabled. | Q-101E-3 / Q-111; [bindings](parent-grant-bindings.md), [cycles](lineage-cycles.md) |
| PC-20 | Published authority revisions are immutable and explicitly adopted. New assignments/upgrades must select the latest published revision and validate its eligibility, without fallback; parent support resolves top-down using actual adopted lineage. | Existing A1 stays on revision 2 after revision 3 is published; a new assignment must validate revision 3. | Silently upgrade A1, use an older fallback for a new assignment, or use unrelated latest publication as parent support. | Q-089-B / Q-102–Q-107 / Q-112A; [revisions](grant-revisions.md) |
| PC-21 | Grant time validity belongs to immutable revision content. Start is inclusive, expiry exclusive; enablement cannot extend the window. | Publish and explicitly adopt a new revision to extend local validity, still within upstream support. | Re-enable an expired revision to reset its expiry. | Q-083 / Q-108/109; [validity](grant-validity.md) |
| PC-22 | Preserve checked authority through the corresponding protected Auth write and application boundary use. Confirmed grant deletion cannot be ignored by a later-started check using stale cache. | Stop a write invalidated by a conflicting support change; reject deleted-grant cache use in a new check. | Persist a proposal different from the checked one or treat cache TTL as a post-deletion grace period. | Q-069 / Q-074 / Q-110; [freshness](authority-freshness.md), [write consistency](auth-write-consistency.md) |
| PC-23 | Completed decisions are allow/deny; inability to complete required evaluation is an error. Deny and error both stop protected execution; preserve required reasons/messages. | Required Auth lookup times out with no sufficient evidence: report evaluation error and no protected output. | Turn the timeout into allow, or assert that it proves no grant exists. | Q-051–Q-067; [results](decision-results.md) |
| PC-24 | Evaluate complete alternative routes without cross-grant field mixing or explicit deny grants in v1. A failed candidate alone is not a final denial. | Consider a separate valid route when FIN-read does not cover an ENG request. | Deny at the first inapplicable grant despite another valid route, or splice its fields into another grant. | DECISION-001/002; [grant model](grant-model.md) |
| PC-25 | Keep operation-specific authorization obligations: deny uncovered collection requests, check complete bulk batches before effects, authorize both move boundaries, and reauthorize queued work at execution. | An uncovered item blocks a bulk request's protected execution. | Silently run an authorized subset or reuse an old queue-time allow as execution authority. | Q-068 / Q-071–Q-075; [operations](operation-enforcement.md), [bulk](bulk-enforcement.md), [background](background-authorization.md) |
| PC-26 | Establish root authority through trusted, registration-first setup; explicitly assign it to the administrators group and explicitly admit the selected legitimate human. Completed bootstrap replay changes no authority. | A legitimate selected human obtains authority through membership, without a special username. | Omit a parent in an ordinary operation to manufacture root, or replay setup to restore removed membership. | Q-113–Q-116; [bootstrap](bootstrap-authority.md), [initial assignment](bootstrap-initial-assignment.md) |
| PC-27 | Published contracts carry a supported top-level string version. Format version is not authority revision. | `version: "1"` identifies the contract; `grant_revision: 2` selects content. | Guess missing versions or interpret numeric `version: 1` as the approved string. | Q-050-A / Q-107; [publication](contract-publication.md) |
| PC-28 | Keep this handbook about authorization. Application boundary enforcement is in scope; business-rule design and the external audit system are not. | Establish FIN ownership and constrain the operation; expose approved decision evidence. | Expand Layer 2 into a business workflow engine or prescribe audit retention here. | Q-076 / Q-084; [scope rejection](grant-conditions.md), [audit boundary](authority-change-audit.md) |

## Guidance, not mandatory access restrictions

| Ref | Preferred practice and rationale | Acceptable exception | Source |
|---|---|---|---|
| PC-G01 | Prefer group-based human assignments: reuse and membership administration avoid unnecessary individual grants. | A valid explicitly authorized direct human assignment remains supported and parent-bounded. | GROUP-004 / Q-021 / Q-112; [groups](groups-and-membership.md) |

The two-owner proposal is not added as an approved recommendation: Q-096's
remaining ownership policy is still open. No preference creates implicit owner
privileges, automatic transfer, or independent proxy authority.

## Limits of this consolidation

Missing details remain visible in the [closure checklist](v1-closure.md), notably
support discovery/evidence, complete contracts, delegation mechanics, identity
mapping, and freshness/concurrency. PC-22 does not settle every in-flight race;
PC-25 does not promise database rollback; PC-26 does not approve Q-117 or complete
the bootstrap trust procedure. Generic condition references are unresolved under
Q-084, not silently removed or converted into an expression language.

Closing the catalog's documentary checkpoint is not closing these other criteria
or certifying an implementation. Historical principle wording remains in its
source chapters, with later approved decisions taking precedence.
