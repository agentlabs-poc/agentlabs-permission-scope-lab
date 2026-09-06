# Team lineage, authority lineage, and scope narrowing

**Current refinement — Q-101:** [parent-grant bindings](parent-grant-bindings.md)
supersedes the specific-supporting-assignment identity requirement below. A
child declares `parent_grant_id`; Auth validates actual assignments in its
recipient's parent-team context. The four-part binding and bottom-up structural
changes remain support-aware without a `parent_assignment_id`. Disabled bindings
also permit parent changes; explicit enablement validates current reality.
Earlier conflicting wording below is preserved as history, not a second current
model. Q-094's orphan definition and permission/scope ceilings remain agreed.

This chapter explains Q-090/Q-091/Q-093 together and records Q-095's canonical
meaning of “sub-,” at the user's request to make parent/child relationships clear.
It does not finalize new wire fields or a separate stored scope entity. Tenant
remains the implied mandatory boundary.

## Canonical meaning of sub- — Q-095

Status: **AGREED.** The user clarified that a subteam is always a subset of its
parent in scope and permissions, just as a subgrant is a subset of its parent
grant. The user then specified that the entities should continue to be called
teams and grants, with a canonical definition of the “sub-” relationship.

- **Team/group:** the canonical entity. A team is a **subteam/subgroup** relative
  to its parent when its total effective authority must remain within that
  parent's authority. Every assigned route must respect this ceiling; an
  independent assignment cannot widen the child team beyond it.
- **Grant:** the canonical reusable authority definition. A grant is a
  **subgrant** relative to a selected parent grant when its permissions are a
  subset of the parent's effective permissions and its effective scope is the
  parent scope AND additional child constraints. The dependency is live, not
  an independent copy; usable authority retains the supporting assignments.
- **Sub-:** qualifies a parent-dependent, non-expanding relationship. It is not
  another entity kind, permission prefix, role, or new wire discriminator.

Subset permits equality: choosing all permitted parent permissions or adding
scope `{}` does not require artificial narrowing. Neither kind of relationship
automatically supplies access or establishes human membership in the parent team.
No `type: subteam` or `type: subgrant` field is introduced.

Rationale: one entity vocabulary keeps the model small; explicit parent links
explain inherited ceilings and continuing dependencies. The team-level ceiling
closes the loophole of deriving one narrow grant while separately attaching a
broader grant to the same child team. Concrete parent-link fields remain open.

## Keep the relationships separate

| Relationship | Meaning | What it does not establish |
|---|---|---|
| Team lineage | Team2 is a child of Team1 and all Team2 authority stays within Team1's authority. | Membership in Team1, automatic access to all parent grants, or a department boundary. |
| Grant/assignment authority lineage | A child authority route depends on one selected parent grant through a specific supporting assignment. | Independent authority merely because both definitions exist. |
| Scope lineage | Restrictions accumulate along that selected authority lineage. | A replacement scope, a union of parent boundaries, or a new standalone scope record. |
| Human membership | A human receives a group's valid assigned authority through membership. | Permission to administer the group or membership of ancestor groups. |
| Ownership/administration — Q-099 | Explicit authority to administer a team, separate from its team-held supporting lineage. | Automatic business access, grant-assignment authority, or import of an owner's personal grants. |

Team lineage tells us the group relationship. Explicit authority lineage tells
us which source supplies permission and boundary constraints. Neither can be
inferred solely from team names, department labels, or the other relationship.
The child-team ceiling is now explicit. Its concrete hierarchy/assignment
representation and validation mechanics remain open; the ceiling itself is not.

## Example: Team1 → Team2

```text
Team relationship:
  Team1
    └── Team2

Selected authority relationship:
  Parent grant: read/write, dept = FIN
    └── Supporting assignment to Team1
          └── Dependent child authority: read; additional cert = C17
                └── Assignment to Team2
                      └── Nutan's Team2 membership → Nutan's applicable authority
```

For this example, the application supports both boundaries for the same
operation. The child has read within FIN AND C17, not write and not ENG.
Maya's separate administrative grant permits creating the relevant assignment;
it is not the business permission source shown in this chain.

Under [Q-099](ownership-lineage.md), Maya must be currently authorized when she
acts, but the selected team-held route continues through Team1's supporting
assignment, not her membership merely because she created it. Any explicitly
personal dependency in the actual selected lineage remains required. The
[ownership SVG](assets/ownership-lineage.svg) shows that distinction.

<details>
<summary>Earlier example qualification — blanket membership retention superseded by Q-099</summary>

If Maya uses her Team1 membership to supply the supporting parent authority
under Q-093, preserve that membership as part of the selected source route too.
The simplified diagram does not remove that dependency or imply that merely
being the team's creator supplies it.

</details>

## Resolution along a selected route

1. Establish the requesting human, trusted tenant, and valid direct/group
   assignment routes. Membership does not traverse ancestor groups implicitly.
2. Load each route's grant definition and its explicit parent support, retaining
   assignment identities and required membership/delegation dependencies.
3. Require the supporting lineage to remain valid. Missing required support
   cannot supply authority; do not fall back to a stored definition alone.
4. Preserve permission subsets and AND the inherited and added scope constraints.
   At every level, the child cannot widen the selected parent grant authority or
   the parent team's authority. No separate assignment may bypass that ceiling.
5. Evaluate the endpoint's one required permission and trusted material against
   a complete route. The endpoint enforces the resulting boundaries on actual
   data and effects. Resolution alone is not an allow.

```text
Child team effective authority ⊆ Parent team effective authority

Child grant permissions ⊆ Parent grant effective permissions

Child grant effective scope = Parent effective scope AND Child additional scope
```

Applied repeatedly, narrowing accumulates along a supported chain. Additional
`{}` preserves the parent's boundary. FIN AND ENG does not become either
department; conflicting constraints cannot authorize an operation. AND must
preserve meaning, not use last-write-wins JSON merging.

This explains recursive evaluation; it is not approval of unlimited hierarchy
depth or completed cycle-detection, consistency, or traversal contracts.

## Orphan grant — Q-094, agreed definition

Status: **AGREED.** The user approved the definition and emphasized explicit,
not automatic, ownership transfer. Previous status: proposed definition.

An **orphan grant** is a grant whose required parent support no longer exists
in its declared authority lineage. It remains a grant, not a separate entity type.
With reusable definitions, orphaning is assessed for the affected assignment/
lineage: a broken route does not invalidate every other valid assignment of
the same definition.

The orphaned route and descendants requiring it cannot supply authority. Their
stored records may remain; missing support does not authorize automatic deletion
or automatic attachment to a different parent.

- An unassigned reusable definition is not automatically orphaned.
- A legitimate root grant with no required parent is not orphaned.
- A missing required parent grant or supporting assignment breaks that route.
- Disabled/expired support makes a dependent route unusable, but does not by
  itself mean the parent record is missing.
- Failure to fetch support, such as an Auth timeout, is not proof that it does
  not exist. It must not permit access or be silently treated as an empty lineage.

Rationale: distinguish stored definitions from usable authority, and missing
parent support from merely having no recipients. The safety consequence follows
the existing dependency rule. Detailed recovery/rebinding and cleanup contracts
remain open; this does not introduce a new wire status or a full lifecycle.

## Changes and stored records

| Change | Consequence |
|---|---|
| Nutan leaves Team2 | Nutan loses this membership-derived route; stored team grants need not be deleted. |
| A required parent assignment is disabled/deleted | Authority through its dependent descendants is unavailable; definitions may remain stored. |
| A required parent definition is missing | The selected lineage cannot supply authority. |
| Maya loses membership required by the selected source route | That membership can no longer support this dependent route. |
| Team1's owners change; actual team-held support and other memberships remain valid and unchanged | Downstream authority through that support is unchanged. Owner rotation does not import personal grants or rebind parents. |
| A grant definition has no assignments | It supplies no access, but is not automatically an invalid definition requiring deletion. |

Q-094's agreed definition above means missing required support, not just an
unassigned reusable definition. Missing support stops the affected authority
route, not necessarily all authority of that user or group. The definition is
agreed; detailed orphan lifecycle remains open. Basic dependency consequences
also follow from Q-091/Q-093.

Separate valid assignments remain alternatives within mandatory ceilings. There
is no global “narrowest grant wins” rule across a human's unrelated routes. Nutan
may have separately authorized access outside Team2 through a direct assignment
or another team; that is not authority of Team2. But no assignment to Team2 may
give Team2 authority beyond its parent team's ceiling. A different assignment
of the same definition does not automatically heal a broken selected lineage.

## Team-based ownership — Q-096, proposed refinement

**Current status: partly resolved by Q-099.** The separation of team-held
continuing support from the acting administrator is now **AGREED**, as explained
in [ownership and lineage](ownership-lineage.md). Required personal dependencies
are not silently removed. Ownership records, transfer permissions, and the
two-owner recommendation remain open; no two-person approval rule is adopted.

<details>
<summary>Original Q-096 proposal — preserved; dependency-agreement gap resolved by Q-099</summary>

The user proposed explicit ownership transfer, recommending two human owners
and team-held source authority rather than separate personal authority sources.
The rationale is continuity without letting a replacement or additional owner's
broader personal grants silently expand authority for existing recipients.

The proposed shape keeps explicit owner/administrative assignments separate from
the selected authority lineage. Both owners administer within the same team-held
source boundaries. Two owners means administrative redundancy, not an implied
requirement for two approvals on every operation. Adding an owner or changing
ownership does not automatically merge personal grants, switch authority parents,
or broaden descendant scope/permissions. Ownership transfer must be explicit.

One dependency distinction still needs agreement: for team-held authority, the
continuing parent would be the team's supporting assignment, not the original
owner's personal membership route. Each acting administrator must still be
currently authorized. This would refine Q-093's retained assigner-membership
dependency; it is not silently applied to existing rules by this proposal.
If the actual team-held parent support disappears, its dependent route still
stops. Ownership records, permissions, transfer/rebinding contracts, and the
status of the two-owner recommendation remain pending.

</details>

## Next contract work

Team-parent representation and authority-parent representation must be explicit
and distinct. Their exact fields, applicable source-selection rules, parent
changes/rebinding, membership evidence, revisions, and cycle/freshness validation
remain open. Names alone do not settle these contracts. No runtime migration or
security-completeness claim follows from this explanatory map.
