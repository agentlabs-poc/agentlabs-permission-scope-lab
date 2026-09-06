# Subteams / subgroups — Q-091

**Current refinement — Q-101:** [four-part bindings](parent-grant-bindings.md)
join grant lineage, team parentage, and assignments. `parent_grant_id` plus
actual parent-team holdings replaces the older specific-parent-assignment-ID
dependency wording below. Support cannot be borrowed from an unrelated team.
Affected assignment changes proceed bottom-up. Removed OR disabled bindings
permit parent changes; explicit assignment enablement validates current support
and boundaries. Earlier conflicting wording is preserved as superseded history;
permission/scope ceilings and direct human membership remain unchanged.

Status: **AGREED.** Subteam and subgroup mean the same thing, following team =
group. The user approved making the scratch model canonical after preserving
the preceding handbook as tag `0.0.1`.

## Canonical meaning of “sub-” — Q-095

The canonical entity remains a **team/group**. “Subteam” or “subgroup” describes
a team in a parent-child relationship; it is not a separate entity type.
Its total effective authority must always be a subset of its parent team's
effective authority, preserving each permission's associated scope and constraints.

This is a ceiling, not an automatic access grant. Authority must still be explicitly
assigned through valid supporting routes. No separate assignment to the child team
may widen it beyond its parent team's authority. Subset includes equality; further
narrowing is permitted, but broader child authority is not.

The same “sub-” qualifier applies to a **grant** derived from a parent grant;
it does not create another grant entity type. See the
[canonical relationship definitions](authority-lineage.md#canonical-meaning-of-sub--q-095).
The user explicitly clarified both subset invariants and requested that the
entities continue to be called teams and grants.

## Canonical authority relationship

A subgroup can receive explicitly selected, dependent authority from its parent
group. It does not automatically receive all parent grants or inherit membership.

For each derived authority route:

- Select one supporting parent grant through its specific parent assignment.
- Child permissions must be a subset of the parent's effective permissions.
- Effective child scope is the effective parent scope AND the child's additional
  constraints. Parent constraints cannot be removed or replaced.
- The parent assignment remains a live dependency; the child is not an
  independent copy. Disabling or deleting that supporting assignment removes
  authority through this route, even if its grant definition still exists.
- Other valid routes remain separate, but every route assigned to a child team
  must remain within the parent team's authority. An unrelated broader grant
  cannot be used to bypass the team-level ceiling.

Tenant remains the implied outer boundary. Group hierarchy is not department
hierarchy: `dept = FIN` remains an application-defined scope restriction, not
something inferred from a team name or membership.

## Example: Team1, Team2, and Nutan

Team1 has the FIN read/write assignment in [Q-090](grant-assignments.md).
An authorized administrator establishes a Team2 derived route selecting read.

| | Parent authority held by Team1 | Derived authority held by Team2 |
|---|---|---|
| Permissions | Payslip read, write | Payslip read |
| Scope | `dept = FIN` | Parent scope AND `{}` = `dept = FIN` |
| Supporting route | Team1's FIN assignment | Depends on that specific assignment |

Nutan's direct membership of Team2 supplies its valid derived authority to her.
It does not make her a member of Team1 or give her Team1's write permission.
No independent grant is copied onto Nutan. The tentative membership wire format
from the scratchpad is not finalized by this decision.

Team1's separate ENG authority cannot donate its scope to the FIN route. To
derive ENG access too, Team2 needs a separate route with its own supporting
parent assignment. No multi-parent field merging is introduced.

## Scope narrowing means AND, not replacement

Child scope `{}` adds no constraints; it does not discard the inherited boundary.
If a supported extra constraint is added, both the parent and child constraints
must govern the same operation's actual data. For example, `dept = FIN` plus a
supported `cert = C17` constraint means FIN AND C17, not an unchecked C17 read.
Application definitions and endpoint enforcement remain required.

`dept = ENG` cannot override a parent's `dept = FIN`. Conflicting restrictions
cannot authorize an operation. Exact creation-time rejection rules remain open.
There are still no scope arrays or OR operators: each descriptor remains a flat
string-value object, and alternatives remain separate complete authority routes.
AND composition is semantic, not a last-write-wins merge of JSON objects.

Selecting fewer permissions and adding restrictions constructs narrower
authority. It does not settle arbitrary administrative scope containment,
recipient eligibility, or permission to create the dependent assignment.

## Administration and membership remain separate

Q-093 now requires both assignment-operation authority and supporting authority
available to the assigner through a valid direct/group route. See
[assignment authority](assignment-authority.md). For a child team, the parent-team
ceiling also applies; the assigner's broader authority cannot override it.

Q-092 now approves [team administration](team-administration.md): create covers
teams and subteams; write includes human membership; delete removes teams.
The initial grant's empty scope is tenant-wide. Creating a subgroup does not
automatically assign derived authority; that requires separate assignment
authorization. The narrow Team1 membership grant discussed earlier below remains
an illustration, not the newly approved initial administrative grant.

Being a group member does not authorize adding other members. Maya's Team1
membership-management grant does not automatically authorize creating Team2,
managing Team2, or assigning its authority. Those operations require explicit
administrative authorization within the applicable boundaries.

The proposed owner record, automatic owner-grant issuance, owner permission set,
and self-addition rules remain open. “Owner” is not an evaluator bypass.
Agents and service accounts are still human-dependent proxies, not group members.

## Relationship to Q-077 and Q-090

Q-077 / GROUP-005's blanket exclusion of authorization subgroups is superseded
for this explicit dependent-authority model. Its restriction against transitive
membership remains: only explicit human-to-group memberships establish membership.
This is not general nested-group membership or automatic ancestor-grant access.

[Q-090](grant-assignments.md) removes recipients from grant definitions. Therefore
the combined dependency must preserve the supporting **assignment**, not merely
the reusable grant ID. The exact dependency field and team-to-team hierarchy
record remain open. No `parent_assignment_id` schema is silently adopted here.

<details>
<summary>Earlier scratch-canonical derived grant — superseded layout, preserved</summary>

Before grant/assignment separation, the following format was made canonical only
in the scratch folder. `parent_grant_id` selected a recipient-bearing parent
grant. This is not the combined post-Q-090 schema.

```json
{
  "version": "1",
  "id": "G-TEAM2-PAYSLIP-FIN",
  "recipient": {
    "type": "group",
    "id": "team-2"
  },
  "parent_grant_id": "G-TEAM1-PAYSLIP-FIN",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {},
  "status": "enabled"
}
```

Its one-source dependency, permission subset, and scope-AND intent survive.
Its recipient-bearing structure and parent-grant-only reference do not specify
the new assignment-aware contract.

</details>

## Still open

Combined derived-assignment and group hierarchy contracts; parent changes and
revision adoption; reactivation/freshness; cycles and deeper chains; full lifecycle
and administrative validation. No silent adoption of broader parent authority is
approved. This settles the model, not a runtime implementation or full schema.
