# Worked use cases — Git hosting, ticketing, HRMS, and accounting

**Q-090/Q-091 update:** recipient-bearing examples below are deprecated layouts,
preserved for their scenarios and enforcement rationale. Use the current
[grant/assignment split](grant-assignments.md) and [subgroup model](subgroups.md).
The old “current working layout” wording describes the `0.0.1` baseline, not a
second canonical contract. Scenario migration and final conformance remain open.

These 16 fictional scenario groups exercise already agreed concepts. They are
not claims about GitHub or any other product's implementation, and do not adopt
new permissions, scope-key catalogs, or application policies for AgentLabs.
Scope uses canonical SCOPE-007 syntax; surrounding grant examples retain the
current working layout. See [current grant formats](grant-format.md).

**Role update — Q-089-B:** role-based grants now adopt explicit immutable
revisions. Older role JSON without `role_revision` is historical/incomplete, not
a latest-revision default. The original live-update outcome of UC-ACCOUNT-004 is
deprecated below and replaced with publication-versus-adoption cases. See
[role revisions](role-revisions.md).

## Common assumptions and reading rules

- Each case is independent. Only the grants named in that case apply; necessary
  group membership is valid unless the case explicitly changes it.
- The actor is Vinay unless named otherwise. The trusted enclosing tenant is
  T-1. Validity and conditions are satisfied unless stated otherwise; their
  fields are omitted for focus, not removed from the grant model.
- Each application's scope keys below are defined for these examples only.
  The endpoint supplies trustworthy material for those meanings. They are not
  arbitrary request-field equality checks or a finalized registration system.
- One endpoint-owned gate obtains sufficient material, uses the shared
  evaluator, and enforces the outcome. There is no prepared handoff.
- ALLOW means the specified operation is supported under the stated facts and
  constraints. DENY means no listed complete grant covers it or a mandatory
  boundary prevents it. A required-material failure also stops execution, but
  is not misreported here as a completed policy denial.
- Returning data or mutating state waits for authorization. Necessary internal
  fact lookups remain tenant/request constrained. A path or body assertion
  alone does not establish an existing resource's relationships.

## Git hosting

For these cases, project selects repositories belonging to that project, and
repository selects the exact repository. These are app-defined boundaries.
The repository-reader role currently contains git:repository::read only.

```json
[
  {
    "id": "UC-GIT-G1",
    "recipient": { "type": "group", "id": "backend" },
    "permissions": ["git:repository::read", "git:repository::write"],
    "scope": { "project": "PROJECT-A" }
  },
  {
    "id": "UC-GIT-G2",
    "recipient": { "type": "group", "id": "backend" },
    "permissions": ["git:repository::read"],
    "scope": { "repository": "REPO-9" }
  },
  {
    "id": "UC-GIT-G3",
    "recipient": { "type": "group", "id": "repository-auditors" },
    "role_id": "repository-reader",
    "scope": {}
  }
]
```

Example endpoint declarations, expressed as explanation rather than a new schema:

| Method/route | Required permission | Selected request material | Additional material when needed |
|---|---|---|---|
| GET /git/projects/{project}/repos/{repo} | git:repository::read | Path project and repo. | Actual repository tenant, project, and ID. |
| PATCH /git/repos/{repo} | git:repository::write | Path repo; identified body description for the proposed edit. | Actual repository tenant and project; applicable authority. |

| Case | Applicable authority and request | Expected result and reason |
|---|---|---|
| UC-GIT-001 | G1: read REPO-1 through PROJECT-A; actual repository project is A. Repeat with a repository actually in B while the path still claims A. | ALLOW the first; DENY the second. Requested project is not proof of actual containment. |
| UC-GIT-002 | G1 + G2: REPO-9 is in PROJECT-B. Request read, then write. | ALLOW read through G2; DENY write. G1's write permission cannot borrow G2's repository scope. |
| UC-GIT-003 | G3: read a repository in T-1, then a repository in T-2. | ALLOW within T-1; DENY cross-tenant. Empty scope is tenant-wide, never globally unrestricted. The role supplies read, not write. |
| UC-GIT-004 | A Vinay-dependent agent uses the human's G1, additionally restricted to REPO-1. REPO-1 and REPO-2 both belong to A. Later Vinay loses the supporting backend membership. | REPO-1 is covered before membership loss; REPO-2 is outside the delegation. After supporting membership loss, neither is covered by this route. No independent agent grant survives it. |

The agent restriction in UC-GIT-004 is a semantic premise under AUTHORITY-002;
this example deliberately does not invent a delegation JSON schema or authorize
who may create that delegation. No other supporting route is assumed.

## Ticketing

For these cases, queue is the ticket's assigned queue. The user boundary means
the ticket's requester human, not its assigned support agent or its creator by
guesswork. This example definition demonstrates scope-owned meaning; other
applications must explicitly define their own relationships.

```json
[
  {
    "id": "UC-TICKET-G1",
    "recipient": { "type": "group", "id": "queue-a-support" },
    "permissions": ["support:ticket::read", "support:ticket::update"],
    "scope": { "queue": "QUEUE-A" }
  },
  {
    "id": "UC-TICKET-G2",
    "recipient": { "type": "group", "id": "employees" },
    "permissions": ["support:ticket::read"],
    "scope": { "user": "$self" }
  },
  {
    "id": "UC-TICKET-G3",
    "recipient": { "type": "group", "id": "queue-b-self-service" },
    "permissions": ["support:ticket::read"],
    "scope": { "queue": "QUEUE-B", "user": "$self" }
  }
]
```

| Method/route | Required permission | Selected request material | Additional material when needed |
|---|---|---|---|
| GET /tickets/{ticket} | support:ticket::read | Path ticket. | Stored tenant, queue, requester identity; trusted human context. |
| PATCH /tickets/{ticket} | support:ticket::update | Path ticket; identified body status. | Current stored ticket boundaries and validated proposed change. |

| Case | Applicable authority and request | Expected result and reason |
|---|---|---|
| UC-TICKET-001 | G1: update a ticket in A, then one in B. | ALLOW A; DENY B. Both permissions in the grant remain queue-A-bound. |
| UC-TICKET-002 | G2: read Vinay's own ticket in B, then Maya's ticket in B. Also try updating Vinay's ticket. | ALLOW own read; DENY another human's read and the update. Being inside self scope does not supply update permission. |
| UC-TICKET-003 | Only G3: read Vinay's ticket in B, Vinay's ticket in A, and Maya's ticket in B. | Only the first is allowed: queue B AND requester Vinay must both hold. No independent G2 route is assumed. |
| UC-TICKET-004 | G1: update T-9, whose stored queue is B. Body includes status and a caller assertion claiming queue A. | DENY. Only declared body inputs supply request material; a caller's queue assertion cannot substitute for the existing ticket's authoritative queue. |

Whether an undeclared body field is ignored or rejected by request validation
is not settled here; either way it cannot manufacture authorization facts.

## HRMS

For these cases, dept denotes the resource's explicitly established department.
The user boundary selects the human's own resources, using a trusted
user-to-employee relationship for payslips and the defined personal relationship
for certificates. Relationship timing remains a separate lifecycle decision.

```json
[
  {
    "id": "UC-HRMS-G1",
    "recipient": { "type": "group", "id": "employees" },
    "permissions": ["hrms:payroll:payslip::read"],
    "scope": { "user": "$self" }
  },
  {
    "id": "UC-HRMS-G2",
    "recipient": { "type": "group", "id": "finance-payroll-readers" },
    "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::download"],
    "scope": { "dept": "FIN" }
  },
  {
    "id": "UC-HRMS-G3",
    "recipient": { "type": "user", "id": "maya" },
    "permissions": ["hrms:employee:certificate::read"],
    "scope": { "dept": "FIN", "user": "$self" }
  }
]
```

| Method/route | Required permission | Selected request material | Additional material when needed |
|---|---|---|---|
| GET /hrms/payslips/{payslip} | hrms:payroll:payslip::read | Path payslip. | Actual tenant, department, owner-employee, and trusted human mapping as needed. |
| GET /hrms/payslips/{payslip}/download | hrms:payroll:payslip::download | Path payslip. | Same applicable boundary facts; GET alone does not determine permission. |
| GET /hrms/certificates/{certificate} | hrms:employee:certificate::read | Path certificate. | Actual tenant, department, and personal-boundary relationship. |

| Case | Applicable authority and request | Expected result and reason |
|---|---|---|
| UC-HRMS-001 | G1: Vinay maps to E-5. Read an E-5 payslip, then an E-18 payslip owned by Maya. | ALLOW the first; DENY the second. Employees is the recipient group; self resolves per human, not to all group members. |
| UC-HRMS-002 | G2: read/download Finance payslips, then request an Engineering payslip. | ALLOW both operations in Finance; DENY Engineering. The scope is shared by the grant's permissions. |
| UC-HRMS-003 | Maya with G3: read her own Finance certificate, another human's Finance certificate, and her own Engineering certificate. | Only the first is allowed: Finance AND Maya's personal boundary. Direct human grants remain supported despite the group-preferred practice. |
| UC-HRMS-004 | G1 + G2: Vinay loses finance-payroll-readers membership but retains employees membership. Read another person's Finance payslip, then his own payslip. | The group-wide Finance route disappears. Another person's payslip is denied; his own remains covered by G1. Removing one route does not erase an independent valid route. |

## Accounting

For these cases, dept is the invoice's established department. The ledger-reader
role initially contains accounting:ledger::read only. There is no implied
monetary limit, self-approval policy, or approval hierarchy: those were not
adopted merely because this is an accounting example.

```json
[
  {
    "id": "UC-ACCOUNT-G1",
    "recipient": { "type": "group", "id": "finance-accountants" },
    "permissions": ["accounting:invoice::read"],
    "scope": { "dept": "FIN" }
  },
  {
    "id": "UC-ACCOUNT-G2",
    "recipient": { "type": "group", "id": "finance-invoice-approvers" },
    "permissions": ["accounting:invoice::approve"],
    "scope": { "dept": "FIN" }
  },
  {
    "id": "UC-ACCOUNT-G3",
    "recipient": { "type": "group", "id": "ledger-auditors" },
    "role_id": "ledger-reader",
    "scope": {}
  }
]
```

| Method/route | Required permission | Selected request material | Additional material when needed |
|---|---|---|---|
| GET /accounting/invoices/{invoice} | accounting:invoice::read | Path invoice. | Actual invoice tenant and department. |
| POST /accounting/invoices/{invoice}/approve | accounting:invoice::approve | Path invoice; no authorization-relevant body fields in this example. | Actual invoice tenant and department; required business validation remains separate. |
| GET /accounting/ledgers/{ledger} | accounting:ledger::read | Path ledger. | Actual tenant and the grant's explicitly adopted role revision. |

| Case | Applicable authority and request | Expected result and reason |
|---|---|---|
| UC-ACCOUNT-001 | G1: read a Finance invoice, then approve it. | ALLOW read; DENY approve. Being inside the boundary does not supply another operation. |
| UC-ACCOUNT-002 | G2: approve a Finance invoice, then an Engineering invoice. Repeat when the required department fact cannot be established. | ALLOW Finance; DENY Engineering. Missing required material stops execution as a failure, not as an allow or prepared result. |
| UC-ACCOUNT-003 | G3: read a ledger in T-1, then a ledger in T-2. | ALLOW within T-1; DENY cross-tenant. Tenant-wide {} does not override the enclosing tenant. |
| UC-ACCOUNT-004 — DEPRECATED by Q-089-B | Historical: G3 remains, but an authorized role edit replaces ledger-read with ledger-export. No other read grant applies. Request ledger-read again. | Historical, no longer canonical: DENY read because the grant used current role permissions. The replacement below requires explicit revision adoption; publishing alone does not withdraw read. |

Historical explanation, superseded by Q-089-B: UC-ACCOUNT-004 assumed evaluation
had obtained the current authoritative role version. Propagation timing remains
open, but a newer published revision is no longer the grant's permission source
unless the grant explicitly adopts it.

### UC-ACCOUNT-004 replacement — publishing is not adopting

Current versioned grant excerpt:

```json
{
  "version": "1",
  "id": "UC-ACCOUNT-G3",
  "recipient": {"type": "group", "id": "ledger-auditors"},
  "role_id": "ledger-reader",
  "role_revision": 1,
  "scope": {},
  "status": "enabled"
}
```

Revision 1 contains ledger-read. Revision 2 contains ledger-export instead of
ledger-read. Other grants do not supply either permission in these cases.

| Variant | Change | Expected result under Q-089-B, assuming all other checks succeed |
|---|---|---|
| Publication | Publish revision 2; G3 stays on revision 1. | Read remains allowed; export is not supplied. Publication neither adds nor removes permissions from G3. |
| Adoption | An authorized, boundary-validated change makes G3 adopt revision 2. | Read is no longer supplied; export is supplied within G3's unchanged tenant scope. Failed adoption leaves revision 1 selected. |

These are authorization semantics examples, not runtime or propagation tests.

## Coverage and deliberate gaps

| Agreed concept | Cases |
|---|---|
| Permission separate from scope | UC-GIT-002, UC-TICKET-002, UC-ACCOUNT-001. |
| AND within one scope | UC-TICKET-003, UC-HRMS-003. |
| Alternative complete grants; no field mixing | UC-GIT-002, UC-HRMS-004. |
| Implicit tenant and explicit {} | UC-GIT-003, UC-ACCOUNT-003. |
| Human-relative self and group dependency | UC-HRMS-001/004, UC-TICKET-002. |
| Human-dependent agent subset | UC-GIT-004. |
| Adopted role-revision expansion | UC-GIT-003's role example needs explicit revision selection; UC-ACCOUNT-004 now distinguishes publication and adoption. |
| Declared inputs versus actual resource facts | UC-GIT-001, UC-TICKET-004. |
| Required-material failure prevents execution | UC-ACCOUNT-002. |

These are documentation consistency cases, not tests against an implemented
authorization engine. Open branches still include administrative scope
containment, bootstrap, exact declaration and resolved-data schemas, scope-key
governance, collection/aggregate/bulk policy, amount limits, approval separation,
and freshness/concurrent changes. Do not infer those mechanisms from these cases
or introduce new scope fields to fill the gaps without discussion.
