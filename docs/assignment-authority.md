# Assignment authority and supporting lineage — Q-093

Status: **AGREED AT RULE LEVEL; combined contract pending.** The user clarified
that the parent grant must be assigned to Maya or a group she belongs to, then
instructed: “lets record Q-093 and proceed with otherthings.” The distinction
between team lineage and scope/authority lineage is explained in
[lineage and resolution](authority-lineage.md).

## Two separate checks

[Q-100](auth-service-authority-gate.md) now shows these checks inside Auth Service:
the administrative evaluator checks operation/recipient authority; the
authority-boundary validator checks the proposed authority's supporting source
and ceilings. Both are mandatory within one endpoint-owned gate before writing.
The [SVG](assets/auth-service-authority-gate.svg) illustrates the flow; exact
validator and dependency contracts remain open.

1. Maya must have administrative authority for the assignment operation and its
   intended recipient.
2. The authority she assigns must derive from a valid supporting parent route
   available to her: directly assigned, or assigned to a group of which she is
   a valid member. Child permissions are a subset; effective child scope is
   parent scope AND additional constraints.

Permission to assign does not itself supply the authority being distributed.
Possessing the source authority does not itself permit assignment. Both checks
are required, along with tenant, registration, validity, and other mandatory limits.

For assignment to a child team, Q-095 also requires all resulting team authority
to remain within its parent team's authority. Even if Maya possesses broader
authority elsewhere, she cannot use it to bypass that child-team ceiling.
This is in addition to each child grant's own parent-dependent subset rule.

## Administrative grant example

This is the operation/recipient-boundary illustration from Q-093, not a complete
derived-assignment contract. The `group` key identifies the group receiving the
assignment, not an application department.

```json
{
  "version": "1",
  "id": "G-ASSIGN-TO-TEAM1",
  "permissions": ["auth:assignment::create"],
  "scope": {
    "group": "team1"
  }
}
```

The definition must itself be assigned to Maya through a valid route. This
administrative grant permits the operation for Team1; it does not permit arbitrary
business permissions or reach. The supporting parent belongs to the authority
being assigned, not an attempt to make `auth:assignment::create` a subset of an
unrelated payslip-read permission.

## Worked authority check

| Component | Example |
|---|---|
| Maya's assignment authority | Create assignments to Team1 |
| Maya's supporting authority | Read/write within FIN, through a valid direct or group assignment |
| Proposed child authority | Read within FIN AND a supported additional `cert = C17` boundary |
| Result of these checks | Within the recipient, permission, and boundary limits, provided all required facts and dependencies are valid |

The scope example assumes the application supports both keys for the operation
and the endpoint binds them to the same actual data. It does not introduce
standard department/certificate types or prove those relationships from names.
Write in ENG cannot be produced from this FIN route. A registered grant found
by ID is not evidence that Maya possesses a valid supporting assignment.

## Continuing dependency

**Q-099 refinement — AGREED:** retain the specifically selected supporting
assignment and its actual required lineage. For team-held authority, continuing
support is the team's assignment, not Maya's ownership or membership merely
because she acted as administrator. Her current authorization and source access
are still checked when she creates or changes an assignment.

A route explicitly supported through Maya's personal assignment or membership
retains that dependency, including when it is upstream of a team assignment.
No existing personal route is silently converted to team-held support. Missing
required support stops the affected route; stored definitions do not repair it.
See [ownership and lineage](ownership-lineage.md) for the rationale, diagram,
owner-change example, and remaining contracts.

<details>
<summary>Earlier Q-093 wording — unconditional assigner-membership interpretation superseded by Q-099</summary>

Preserve the specific supporting assignment and any membership through which
Maya supplies the selected parent authority. A parent grant ID alone is
insufficient. If required support becomes invalid, that dependent route cannot
supply authority; leaving stored definitions intact does not repair it.

</details>

This is distinct from merely remembering who issued an assignment. Losing an
administrative permission to create future assignments is not automatically the
loss of the explicitly selected source authority. Do not turn unrelated changes
to Maya's rights into revocation of every assignment she ever issued.

An alternate assignment of the same reusable definition is not an automatic
replacement for the selected supporting route. Exact rebinding, freshness,
reactivation, and parent-reference fields remain open.

## Rationale and reconciliation

This supplies a concrete source ceiling instead of treating “may assign to
Team1” as “may assign anything to Team1.” Permission selection and scope AND
construct narrower derived authority while keeping administration separately
authorized. The source must be available through a valid route, not just a
definition the caller can name.

ADMIN-002's earlier provision-without-possession alternative is **superseded
for this dependent-assignment model**: Maya now needs the authority she
distributes. The separation of using authority from administering assignments
remains. [The original rationale](grant-model.md#administration-without-business-access--admin-002-admin-003--q-022)
is preserved as history, not a competing exception to Q-093.

ADMIN-006 / Q-046 remains the distinction between ordinary issuance history and
explicit live dependencies. Q-093's dependent assignments must not be treated as
independent ordinary bindings to bypass their selected source. Migration of old
independent assignments is not decided here.

Bootstrap must establish the explicit supporting authority needed for subsequent
assignment, as well as administrative authority. Q-092's team create/write/delete
grant alone is not enough to distribute unrelated business access. The complete
seed set and trust procedure remain open; no post-bootstrap bypass is introduced.

## Contract gaps retained

Canonical representation of the child-to-parent assignment link and supporting
membership; team hierarchy representation; parent revisions, rebinding and
freshness; full administrative registration/validation; and source-scope binding
when human-relative selectors such as `$self` cross recipient contexts. Identical
scope text is not automatically identical reach. No new lineage field or silent
revision adoption is approved by recording the rule.
