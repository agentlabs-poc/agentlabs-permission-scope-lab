# GitHub-style example: projects, repositories, teams, and members

This second example tests whether the authorization model generalizes beyond payroll.

It is a conceptual code-hosting example using familiar GitHub-style nouns. It does not claim to reproduce GitHub's internal authorization implementation or exact permission names.

**Implementation status: this is a target design, not what `agentlabs-auth` does today.** Its real role/permission grant tables (`authorization_subject_role_bindings`, `authorization_subject_permission_assignments`) bind to a single `membership_id` each—no `subject_type` column, no group/team reference anywhere. The bundle-resolution code filters strictly on `membership_id` and stamps every result `Source: "membership"`; there is no group join in that path at all. A `groups`/`group_memberships` table pair is described in a roadmap doc but exists in no applied migration. The one place "groups" appears in the real codebase is an upstream SSO identity provider's group claim during federation—and Auth's own docs are explicit that "raw upstream groups never grant application access." So everything below—Teams as group principals, `{"principal": {"type": "team", ...}}`—describes what this model should support, not a feature that works yet.

It introduces two different structures:

```text
Resource containment                  Principal membership

Organization                          Team Backend
├── Project Commerce                  ├── user:vinay
│   ├── Repository API                └── user:arjun
│   └── Repository UI
└── Project Internal                  Team Auditors
    └── Repository Payroll            └── user:maya
```

A **Project or Repository is a resource**. A **Team is normally a group principal**. Mixing those concepts creates confusing policies.

## 1. Permission vocabulary

The code-hosting application registers:

```text
codehost:organization::read
codehost:project::read
codehost:project::manage
codehost:repository::read
codehost:repository::write
codehost:repository::admin
codehost:team::read
codehost:team:member::manage
```

Resource depth is permitted where it expresses a real relationship. `team:member` means membership below a team. A repository remains its own resource type even when it is contained by a project.

## 2. Resources created by the application

```text
ORG-ACME · tenant TENANT-001
├── PROJECT-COMMERCE
│   ├── REPO-API
│   └── REPO-UI
└── PROJECT-INTERNAL
    └── REPO-PAYROLL
```

The code-hosting application owns this containment graph and supplies trusted target attributes:

```json
{
  "resource_type": "codehost:repository",
  "resource_id": "REPO-API",
  "tenant_id": "TENANT-001",
  "organization_id": "ORG-ACME",
  "project_id": "PROJECT-COMMERCE"
}
```

Auth does not independently guess containment from a URL or resource name.

## 3. Teams and members

Teams are tenant-owned group principals:

```text
teams
┌──────────────┬────────────┬──────────────┐
│ id           │ tenant_id  │ name         │
├──────────────┼────────────┼──────────────┤
│ TEAM-BACKEND │ TENANT-001 │ Backend      │
│ TEAM-AUDIT   │ TENANT-001 │ Auditors     │
└──────────────┴────────────┴──────────────┘
```

```text
team_memberships
┌──────────────┬────────────┬────────────┐
│ team_id      │ tenant_id  │ principal  │
├──────────────┼────────────┼────────────┤
│ TEAM-BACKEND │ TENANT-001 │ user:vinay │
│ TEAM-BACKEND │ TENANT-001 │ user:arjun │
│ TEAM-AUDIT   │ TENANT-001 │ user:maya  │
└──────────────┴────────────┴────────────┘
```

Removing Vinay from `TEAM-BACKEND` removes authority inherited through that team. It does not delete direct grants Vinay may hold separately.

## 4. Repository Maintainer role

The tenant administrator creates:

```text
Role: Repository Maintainer

Permissions
├── codehost:repository::read
└── codehost:repository::write
```

The role is assigned to the team at one exact repository:

```json
{
  "principal": {
    "type": "team",
    "id": "TEAM-BACKEND"
  },
  "role": "Repository Maintainer",
  "scope": {
    "type": "resource_exact",
    "resource_type": "codehost:repository",
    "resource_id": "REPO-API"
  }
}
```

The results are:

| Team member request | Target | Decision |
|---|---|---:|
| Vinay reads | REPO-API | Allow |
| Vinay writes | REPO-API | Allow |
| Vinay reads | REPO-UI | Deny |
| Maya reads | REPO-API | Deny |

## 5. Project Viewer role with subtree reach

The role explicitly declares both project and repository read capability:

```text
Role: Project Viewer
├── codehost:project::read
└── codehost:repository::read
```

The tenant administrator assigns it to `TEAM-AUDIT` at the Commerce project:

```json
{
  "principal": {
    "type": "team",
    "id": "TEAM-AUDIT"
  },
  "role": "Project Viewer",
  "scope": {
    "type": "resource_subtree",
    "resource_type": "codehost:project",
    "resource_id": "PROJECT-COMMERCE"
  }
}
```

The scope contains resources that the application reports as descendants:

```text
PROJECT-COMMERCE    Allow project read
├── REPO-API         Allow repository read
└── REPO-UI          Allow repository read

PROJECT-INTERNAL     Deny
└── REPO-PAYROLL     Deny
```

The subtree does not invent capabilities. If the role omitted `codehost:repository::read`, project scope alone would not authorize repository reading.

## 6. Direct member access

A tenant administrator can give Maya temporary read access to one repository without adding her to a team:

```json
{
  "principal": {
    "type": "user",
    "id": "user:maya"
  },
  "permission": "codehost:repository::read",
  "scope": {
    "type": "resource_exact",
    "resource_type": "codehost:repository",
    "resource_id": "REPO-PAYROLL"
  },
  "valid_until": "2026-09-30T23:59:59Z"
}
```

The direct assignment uses the same scoped-grant model as a role assignment.

## 7. Tenant configuration screen

```text
Active tenant
Acme · TENANT-001 🔒

Assign access
Principal type   [ Team ▾ ]
Principal        [ Backend ▾ ]
Role             [ Repository Maintainer ▾ ]
Scope type       [ Exact resource ▾ ]
Resource type    [ Repository ▾ ]
Repository       [ API · REPO-API ▾ ]
Validity         [ No expiry ▾ ]
```

Before creating it, the UI explains:

```text
Members of Backend will be able to:
✓ read REPO-API
✓ write REPO-API

They will not be able to:
✗ administer REPO-API
✗ access REPO-UI through this assignment
✗ access repositories outside TENANT-001
```

## 8. Cross-tenant prevention

Auth validates all four tenant roots:

```text
grantor tenant   = TENANT-001
team tenant      = TENANT-001
role tenant      = TENANT-001
resource tenant  = TENANT-001
```

Any mismatch denies assignment creation. A tenant administrator cannot add a member from another tenant to bypass this boundary unless that principal first receives a valid membership in the active tenant.

## 9. Request consumption

Vinay requests `REPO-API`:

```text
Authenticated principal            user:vinay
Expanded group principal           TEAM-BACKEND
Required permission                codehost:repository::read
Assignment scope                   resource_exact:REPO-API
Trusted target                     REPO-API / PROJECT-COMMERCE / TENANT-001
Decision                           ALLOW
```

Vinay requests `REPO-UI`:

```text
Capability via TEAM-BACKEND         ✓
Tenant boundary                     ✓
Exact resource scope                ✗
Decision                            DENY
```

## 10. Schema ownership

```text
AUTH
├── tenants and principal memberships
├── group principals
├── group memberships
├── roles and role permissions
├── principal/group role assignments
├── assignment scopes
├── validity and assignment versions
└── assignment and decision audit

CODE-HOSTING APPLICATION
├── organizations
├── projects
├── repositories
├── containment relationships
├── operation-permission manifest
├── scope target validation
└── target-context resolution
```

This example proves that the core model supports both individual and group principals, exact resources and resource subtrees, without placing IDs inside permission strings.
