# Current grant formats — canonical v1 scope

## Current: recipient-free definitions and separate assignments — Q-090

The canonical minimal formats now live in [grants and assignments](grant-assignments.md).
A grant has no `recipient`; the assignment carries `grant_id`, `recipient`, and
its enabled/disabled status. [Subgroups](subgroups.md) add explicitly dependent,
narrower authority; their combined assignment-aware wire format remains open.

Scope descriptors remain flat, string-valued AND selectors. A derived route adds
constraints to its parent boundary; `{}` cannot discard that parent boundary.
No stored state is migrated, and no contract-version compatibility rule is implied.

<details>
<summary>Pre-Q-090 grant layouts — deprecated recipient-bearing representation, preserved</summary>

All “current” layout labels below describe the `0.0.1` baseline. Their scope,
role-revision, and complete-authority invariants survive unless explicitly
superseded; their recipient-bearing grant format does not.

Role-reference update — Q-089-B / ROLE-003: role-based grants include an explicit
`role_revision` alongside `role_id`. Roles publish immutable revisions; grants
adopt them through authorized boundary validation. Older unpinned role examples
and automatic live-role wording are deprecated, not a latest-revision default.
See [role revisions](role-revisions.md) for the approved excerpts and rationale.

Q-050-A / CONTRACT-010 now settles the shared top-level version field and string
value, initially `"1"`, plus missing/malformed/unsupported-version rejection.
The complete grant schema remains open. Unversioned working examples below are
retained illustrations, not publishable full contracts; earlier version-syntax
open wording is historical after this approval.

Publication status — CONTRACT-009: every published JSON/YAML contract must
include a version. The unversioned examples below remain **working illustrations,
not complete published contracts**. Agreed scope semantics remain current;
version representation and the complete grant schema are not yet settled.
See [contract publication requirements](contract-publication.md).

Q-048 agrees the resolved-grant meaning and membership-based retrieval flow in
the [grant chapter](grant-model.md). Views for a human include direct grants
and grants through valid group memberships, retaining the original bindings and
dependencies. Resolved is not allowed. The transport schema remains open; the
expanded example below is not silently promoted to a complete resolved format.

This chapter updates the working grant examples to use SCOPE-007's canonical
scope format. The earlier [grant examples](grant-examples.md), GRANT-EX-001
through GRANT-EX-006, remain intact and explicitly deprecated as layouts.
Their complete-binding and dependency semantics remain; the live-role update
meaning is now superseded by ROLE-003.

## What changes and what stays

Lifecycle update: Q-082 consolidates permanent removal into delete, superseding
the separate revoke operation. Create, enable, disable, and delete are agreed.
Q-081 revised approves `status: enabled/disabled` in
[grant lifecycle](grant-lifecycle.md). The former three-value
proposal is preserved there as superseded, not a canonical enum.
The older `active` spelling below remains illustrative, not a competing canonical
enum; no historical grant examples are removed.

A grant remains recipient + permissions/role + scope + validity/conditions
(GRANT-001/002, ROLE-001, TERM-004). We are not adding another grant entity.
The current examples consistently use the already illustrated permissions
array for explicit permissions and role_id for a role reference. They do not
introduce an administrator-specific format or revive independent service grants.

Scope syntax, enabled/disabled status values, and explicit `role_revision` selection
are canonical. Full lifecycle/condition schemas, remaining role revision validation,
and resolved-grant transport contracts remain open. The
surrounding JSON below is the current working grant layout, not a claim that
all those branches or application migrations are complete.

## Current versioned illustration — approved status values

This partial illustration uses the agreed version, scope, and status conventions;
it is not a complete grant schema. Missing-field defaults and validity/conditions
are not settled by their omission here. Q-083 in the lifecycle chapter now approves
optional start-inclusive, expiry-exclusive time validity separately; timestamp
validation and complete lifecycle contracts remain open.

```json
{
  "version": "1",
  "id": "G-17",
  "recipient": {"type": "group", "id": "finance-readers"},
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"dept": "FIN"},
  "status": "enabled"
}
```

The unversioned `active` illustrations below are preserved historical layouts,
not current canonical status or complete publishable contracts.

## Scope-format correspondence

These equivalences assume the application defines dept and user with the same
boundary meanings as the corresponding historical examples. They are not a
mechanical migration rule for arbitrary application keys.

| Deprecated scope syntax | Current scope syntax | Meaning preserved |
|---|---|---|
| {"type":"tenant"} | {} | Tenant-wide reach inside the enclosing trusted tenant. |
| {"type":"department","id":"FIN"} | {"dept":"FIN"} | Finance boundary. |
| {"type":"employee_self"} | {"user":"$self"} | The authorizing human's personal boundary through the defined employee relationship. |

Multiple entries combine with AND. Alternatives use separate grants. Missing
or null scope is invalid, never an implicit tenant-wide default. Self remains
human-relative even when the grant recipient is a group or access is exercised
through a dependent proxy.

## Explicit permission grant

The following is G-1 restated with the current representation, not a second
grant or an executed migration. It preserves the historical example's dates.

```json
{
  "id": "G-1",
  "recipient": { "type": "user", "id": "maya" },
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {},
  "status": "active",
  "validity": {
    "not_before": "2026-09-01T00:00:00Z",
    "expires_at": "2026-10-01T00:00:00Z"
  },
  "conditions": []
}
```

The permission still applies only while the complete binding is valid, within
trusted tenant context. Adding download to the permissions array would bind
both operations to exactly the same scope, validity, and conditions. Different
restrictions require a separate binding, not independently scoped array entries.
Lifecycle field names above are retained illustrative conventions, not newly
settled status/condition semantics.

## Group grant with multiple permissions

This abbreviated restatement of G-5 preserves its role in the historical
example: certificate-operators supply read and revoke within Finance. Status,
validity, and conditions are omitted for focus, not removed from the model.

```json
{
  "id": "G-5",
  "recipient": { "type": "group", "id": "certificate-operators" },
  "permissions": [
    "hrms:employee:certificate::read",
    "hrms:employee:certificate::revoke"
  ],
  "scope": { "dept": "FIN" }
}
```

Read plus revoke remain Finance-bound. A separate tenant-wide read grant does
not widen this revoke permission. Group access stays dependent on the source
grant and membership; groups are preferred for human access, not the only
permitted recipient kind. The group's creator is not implicitly authorized.

## Role-referencing grant

**Current — Q-089-B:** the role supplies permissions from the immutable revision
the grant explicitly adopts. Publication alone does not change the grant.

```json
{
  "version": "1",
  "id": "G-17",
  "recipient": {"type": "group", "id": "finance-readers"},
  "role_id": "R-17",
  "role_revision": 1,
  "scope": {"dept": "FIN"},
  "status": "enabled"
}
```

This is the approved role-reference excerpt, not the full grant schema. Resolving
it preserves the source grant and adopted revision; it does not select latest or
create independent authority. The [revision chapter](role-revisions.md) gives the
corresponding role definitions and adoption checks.

<details>
<summary>Earlier unpinned role grant and expanded view — deprecated by Q-089-B; retained unchanged</summary>

G-6 binds the current Certificate Reader role to Maya within Finance. The role
definition supplies its read/download permission bundle, not scope.

```json
{
  "id": "G-6",
  "recipient": { "type": "user", "id": "maya" },
  "role_id": "certificate-reader",
  "scope": { "dept": "FIN" }
}
```

The current role definition affects referencing grants under ROLE-002. It does
not replace the grant's recipient, scope, validity, or conditions. How future
role edits interact with scope-key applicability remains an open validation
branch, not permission to ignore incompatible restrictions.

## Expanded view of that same grant

The following computed view is still G-6. It is not an additional independent
grant or proof that request-specific authorization is complete. This retains
RESOLUTION-003's original example and uses current scope syntax; the final
resolved-grant schema is still open.

```json
{
  "id": "G-6",
  "recipient": { "type": "user", "id": "maya" },
  "permissions": [
    "hrms:employee:certificate::read",
    "hrms:employee:certificate::download"
  ],
  "scope": { "dept": "FIN" },
  "source": { "role_id": "certificate-reader" }
}
```

Preserve any original validity, conditions, and additional dependency evidence
in a real evaluation view even though this abbreviated example omits them.

</details>

## Self and AND-boundary examples

GRANT-EX-007 in [grant examples](grant-examples.md) supplies current group grants
for user:$self, dept:FIN AND user:$self, and explicit tenant-wide {}. These
complete the older examples' personal-boundary and broad-access cases without
adding another scope or grant structure.

## Deprecation coverage and remaining work

| Earlier example | Current counterpart |
|---|---|
| GRANT-EX-001: explicit permission, tenant/department | Explicit grant above plus canonical scope correspondence; GRANT-001's no-field-mixing rule stays. |
| GRANT-EX-002: group recipient | Group grant above; one permission is the same array shape with one entry. |
| GRANT-EX-003: multiple permissions/direct and group | Explicit/group layouts above, each preserving all binding restrictions. |
| GRANT-EX-004: role reference | Role-referencing layout above; {} represents the earlier tenant-wide variant. |
| GRANT-EX-005: expanded role view | Same-grant expanded view above. |
| GRANT-EX-006: employee-group self | GRANT-EX-007, using the defined user:$self boundary. |

No stored grants, application code, or external authorization state have been
migrated. Next settle scope-key definition/governance, return to administrative
bounds, then finish lifecycle and request/resolved-grant contracts. Full v1
handbook coverage remains open.

</details>
