# Grants and assignments — Q-090

**Later revision rules — Q-102–Q-106:** [grant revisions](grant-revisions.md)
adds explicit per-assignment adoption and one current assignment per tenant/
grant/recipient, including disabled assignments. New assignments and explicit
upgrades select latest published revision; existing assignments may remain.
The [core JSON representation](grant-revision-format.md) is approved under Q-107;
the earlier examples below remain the pre-revision-format checkpoint.

**Current refinement — Q-101:** [parent-grant bindings](parent-grant-bindings.md)
uses `parent_grant_id` plus actual assignments/team context; no additional
parent-assignment lineage reference is required for the settled cases. Older
wording below requiring one specific assignment identity as the child's source
is superseded, not the need for real assignment records or required parent-team
support. Assignment disablement and grant-wide disablement are separate controls.
Relevant bindings may be removed OR disabled before parent changes; re-enabling
a retained assignment must validate the changed reality.

Later refinements: [Q-093](assignment-authority.md) requires administrative
assignment authority plus a supporting parent route available to the assigner.
[Q-095](authority-lineage.md) defines “sub-” as a relationship, not a new entity:
child grants and all authority of child teams must remain within their respective
parents. Old no-possession-ceiling wording below describes Q-090's checkpoint;
it is superseded for Q-093's dependent-assignment model.

Status: **AGREED.** The user instructed: “now lets make the sub-team/sub-group
canonical and grant without recipient canonical.” This promotes the discussed
grant/assignment separation into the handbook. Tag `0.0.1` preserves the prior
baseline at `247e8392bb09885a9ff1c8ce94e5205a279e6852`.

## Canonical separation

- A **grant** is a reusable authority definition: permissions (or an adopted role
  revision) and their scope. It has no `recipient`.
- An **assignment** binds a grant to a recipient. Its `grant_id` references the
  definition; its `recipient` identifies the human or group receiving it.
- A definition alone gives nobody access. Creating it and authorizing its
  assignment are separate responsibilities.
- An authorization route retains the assignment, grant, applicable membership,
  and every required dependency and restriction. Separation does not permit
  mixing a permission from one grant with another grant's broader scope.

Tenant is an implied mandatory boundary, not a scope key. Team/group membership
determines who receives authority; application scope keys such as `dept` determine
its reach. A group is not implicitly a department and does not own one shared scope.

## Canonical examples

These are the approved minimal shapes, not complete lifecycle or migration
schemas. Their top-level `version` is the contract version, not a definition
revision. Definition revision/adoption fields remain open.

Reusable grant:

```json
{
  "version": "1",
  "id": "G-PAYSLIP-FIN",
  "permissions": [
    "hrms:payroll:payslip::read",
    "hrms:payroll:payslip::write"
  ],
  "scope": {
    "dept": "FIN"
  }
}
```

Assignment to Team1:

```json
{
  "version": "1",
  "id": "A-TEAM1-PAYSLIP-FIN",
  "grant_id": "G-PAYSLIP-FIN",
  "recipient": {
    "type": "group",
    "id": "team1"
  },
  "status": "enabled"
}
```

Another authorized assignment can reuse the same definition for another
recipient. Disabling this assignment withdraws Team1's route without disabling
other assignments of the definition. Neither definition creation nor possession
of business access supplies assignment or membership-management authority.

## Database representation

An implementation may store assignments in a database table linking each
`grant_id` to a user or team/group. Separate rows let different recipients reuse
the same grant definition. The assignment links the complete grant—permissions
and scope—not a standalone scope detached from its permissions.

The canonical JSON describes the relationship; it does not require storing JSON
documents or prescribe a physical table schema. Database operations must still
enforce assignment authorization, trusted tenant isolation, and applicable
dependencies. Human-to-group membership is a separate relationship.

Rationale: keep the shared contract clear while allowing ordinary relational
storage and recipient-based retrieval without duplicating grant definitions.

## Resolution and reuse

```text
Verified human
  ├── direct assignments ─────────────────────┐
  └── valid human-to-group memberships        │
          └── assignments to those groups ────┤
                                             ↓
                              grant definitions + restrictions
                                             ↓
                              complete dependent evaluation views
```

Keep assignment identity as well as grant identity. Two assignments of one
definition are different authority routes; removing one must not be masked by
the definition remaining available. A computed view is neither a new assignment
nor an allow. `$self` remains anchored to the authorizing human, not to the
definition, group, creator, or assigner. Reuse does not freeze it to one person.

For [subgroups](subgroups.md), retain the specific supporting parent assignment.
The continued existence of a reusable parent definition is not evidence that
the parent team still holds the supporting authority.

## Rationale and consequences

Separating reusable authority from recipient bindings makes reuse and
recipient-based retrieval explicit. It also distinguishes authoring authority
definitions from distributing actual access. This is a logical contract decision,
not a required database layout or evidence of a performance improvement.

Shared changes must not silently expand existing assignments. The immutable
role-revision and explicit adoption principles of [Q-089-B](role-revisions.md)
remain; a safe revision/adoption contract for reusable grant definitions is still
needed. These minimal examples must not be interpreted as approval to mutate a
shared definition in place and automatically propagate broader authority.

Historical Q-090 checkpoint, superseded in part by Q-093 as stated above:
the earlier rule that a valid ordinary human/group binding survives its issuer's
later loss of issuance authority remains. That is different from an explicitly
dependent subgroup route or a human-dependent proxy. This decision does not
approve a personal-business-access ceiling for administrators.

## Deprecations and open contracts

- TERM-004's “grant and assignment are the same logical record” rule is
  **deprecated**. They now represent distinct concepts.
- GRANT-001/002 and ROLE-001's recipient-bearing layout is **deprecated as a
  representation**. Permission/scope association, role meaning, and complete
  authority validation remain required across the definition and assignment.
- Old recipient-bearing examples remain historical; they are not an alternative
  current schema. [Previous formats](grant-format.md) are preserved, not migrated.
- Still open: definition revision/adoption fields, full role-based definition
  excerpts, placement of time validity/conditions and definition lifecycle,
  administrative permission/scope encoding, migration/version compatibility,
  and combined subgroup assignment dependency fields.
- Assignment provenance must survive resolution, but this decision does not add
  fields to the minimal evaluator result or design an audit system.

Human-only group membership, dependent agents/service accounts, one permission
per endpoint, the endpoint-owned gate, and application boundary enforcement remain.
