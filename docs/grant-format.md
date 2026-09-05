# Current grant formats — canonical v1 scope

This chapter updates the working grant examples to use SCOPE-007's canonical
scope format. The earlier [grant examples](grant-examples.md), GRANT-EX-001
through GRANT-EX-006, remain intact and explicitly deprecated as layouts.
Their agreed binding, role, and dependency semantics are not deprecated.

## What changes and what stays

A grant remains recipient + permissions/role + scope + validity/conditions
(GRANT-001/002, ROLE-001, TERM-004). We are not adding another grant entity.
The current examples consistently use the already illustrated permissions
array for explicit permissions and role_id for a role reference. They do not
introduce an administrator-specific format or revive independent service grants.

Scope syntax is canonical. Final lifecycle/status/condition schemas, role
revision encoding, and resolved-grant transport contracts remain open. The
surrounding JSON below is the current working grant layout, not a claim that
all those branches or application migrations are complete.

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
