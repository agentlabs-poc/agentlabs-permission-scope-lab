# Application registration and Auth validation

Q-042 now agrees REGISTRATION-004: all existing grants must pass before
relationship validation is enabled. Otherwise reject activation, report the
incompatible grants, and preserve the previous configuration pending explicit
authorized correction of grants or declarations. Do not silently modify or
delete grants, or grandfather incompatible grants into enabled mode.

The user directed moving to another branch. Remaining registration lifecycle
details are parked, not finalized or excluded from v1. Return to grant
administration, ADMIN-004/005, with Q-043. Earlier Q-042-open prose is history.

Current: REGISTRATION-003 / Q-041 is approved. Each application explicitly
declares upfront whether relationship validation is enabled or disabled.
Enabled means mandatory validation for every grant, not per-grant optionality.
Earlier omission-based proposals below are preserved but superseded by this
explicit registration-level choice.

Latest: REGISTRATION-002 / Q-040 is approved with the user's qualification that
declaring supported permission-scope relationships is an optional feature of
the registration flow. Individual permission and scope registration remains
agreed under Q-039. Omission behavior is proposed separately in Q-041;
older Q-040-open statements below are preserved historical snapshots.

REGISTRATION-001 / Q-039 is agreed. The user refined the shared-definition
proposal: applications register their supported scopes and permissions; Auth
validates before accepting grants and remains abstract and domain-agnostic.
The user approved the refined principle, not a registration payload or API.

## Agreed responsibility boundary

Applications register their supported permission and scope contracts with Auth.
Auth validates grant acceptance against those registrations and canonical
authority rules. It does not interpret application-specific business meaning.

| Auth validates | Application interprets and establishes |
|---|---|
| A permission is registered for the relevant application. | What business operation that permission represents. |
| A scope key is registered for the relevant application. | What resource relationship that boundary key represents. |
| Scope follows canonical syntax and declared constraints. | Trusted facts establishing whether a particular resource is inside the selected boundary. |

The distinction is not that Auth does no reasoning at all. Auth still enforces
canonical validity rules and the caller's authority to create or change a
grant. A structurally valid grant is not automatically an authorized grant.
Administrative permission remains separate from business access (ADMIN-001/002);
the exact grantor bounds and registration-management authority remain open.

Auth need not understand payroll, inspect payslips, query the application
database, or execute application-domain resolvers to perform this registration
validation. Recognizing a validly formed reference does not prove that the
referenced domain object exists. Existence checks and their timing remain open.

## Example and counterexamples

Suppose HRMS registers the permission `hrms:payroll:payslip::read` and the scope
key `dept`. A grant's permission-and-scope fragment could be:

```json
{
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": { "dept": "dept-1" }
}
```

This is not a complete grant or a registration schema. Auth can recognize the
registered permission and key, check the canonical scope shape and declared
constraints, and separately check authority to issue the grant. It does not
establish that dept-1 exists or that payslip P-17 belongs to it.

At the endpoint-owned gate, application bindings provide trusted payslip facts
under the application's declared department relationship. The embedded auth
agent evaluates the complete requirements using shared canonical rules, and
the endpoint enforces the result (ARCH-004/005 and CONTRACT-006).

An unregistered permission or scope key cannot pass registration validation.
An unauthorized caller cannot issue an otherwise well-formed grant. Conversely,
registration of both a permission and a scope key does not by itself establish
that they are a supported combination. Permission-scope compatibility is the
next question, Q-040, not an implicitly approved validation mechanism.

## Optional relationship declarations — Q-040, approved with qualification

The registration flow binds the application's permissions and scopes to any
declared support relationships. Auth validates those declarations without
understanding application business concepts. Declaring the relationships is
an optional feature, not a mandatory registration requirement.

```text
Application registers supported permissions and scope keys
                         |
                         v
Optionally declares their supported relationships in registration
                         |
                         v
Auth uses registered contracts when validating grant acceptance
```

This is a logical dependency flow, not three required API calls or a finalized
registration ordering protocol. It adds no grant fields. Representation and
update transaction mechanics remain open.

For illustration, an application might declare that payslip-read supports
`dept` and `user`, while repository-read supports `project` and `repository`.
Where the declared contract excludes `dept` for repository-read, Auth can reject
that combination through metadata validation without interpreting departments.
These are example mappings, not universal catalogs or finalized array fields.

Optional registration metadata does not remove Q-038's runtime requirement to
evaluate scope under explicitly supported application meanings. Unsupported
relationships cannot become unrestricted access by ignoring scope entries.
Valid registration is also not authority for a caller to issue a grant.

### Q-041 — historical omission-based proposal, superseded

The proposal below was refined before approval: omission does not choose the
mode. The application explicitly declares that choice under REGISTRATION-003.

Recommended interpretation: without relationship metadata, Auth checks registered
permissions and keys, canonical syntax, declared constraints, and grant-issuance
authority, but does not claim to have established their compatibility. The
endpoint-owned gate must still establish supported resource relationships and
satisfy the complete scope before access. If relationship metadata is supplied,
Auth additionally validates against it.

This omission behavior is not yet approved. Optionality per application versus
permission, absent versus empty, partial or invalid declarations, multi-key
combinations, role changes, and metadata removal remain open. Calling the
feature optional does not adopt silent fallback for invalid declarations.

## Explicit application-level choice — REGISTRATION-003 / Q-041, agreed

The application declares the relationship-validation choice upfront in its
registration. Optionality means choosing whether to enable the feature for
the application, not choosing whether an individual grant is checked.

| Declared choice | Auth's obligation |
|---|---|
| Enabled | Every grant must pass the declared permission-scope relationship checks; there is no per-grant bypass. |
| Disabled | Auth still checks registered permissions and scope keys, canonical format, declared constraints, and issuance authority, but does not check permission-scope compatibility. |

In either mode, the endpoint-owned gate must establish that the actual resource
satisfies the complete scope before permitting access. Disabled compatibility
validation does not authorize unsupported runtime relationships or ignoring
scope restrictions.

For example, if the enabled application's declared support contract excludes
`dept` for repository-read, a grant combining them is rejected even if both
identifiers are individually registered. Missing relationship metadata cannot
silently disable an enabled mode or waive its validation requirement.

"Every grant" includes role-based grants and existing grants, not only new
direct grants. Registration or role changes that affect compatibility therefore
need revalidation mechanics preserving this invariant; the approval does not
yet select those mechanics, propagation timing, or a migration algorithm.

No configuration field name or JSON representation is adopted. The choice
must be explicit: there is no agreed omission-based default. Handling a missing
choice, empty/partial declarations, multi-key combinations, role expansion,
authorized mode changes, and compatibility updates still needs detail.

### Q-042 — historical activation question, now answered above

Next question: when enabling relationship validation would leave existing
grants incompatible, should activation be rejected until those grants have
been explicitly corrected? This is not yet decided. The user has agreed that
enabled mode covers all grants; neither grandfathering invalid grants nor
silently deleting, narrowing, or disabling them is adopted by this question.

## Still open

- Registration representation, APIs, storage, distribution, and versioning.
- Who may register or change an application's definitions, and application
  namespace/identity protection. Registration is not authority to self-grant.
- Declaring supported permission-scope combinations, including multi-permission
  grants and role evolution; Q-040 begins this branch.
  Update: Q-040 now settles optional declarations in registration. Q-041 and
  the detailed compatibility mechanics remain open.
  Latest: Q-041 now settles the explicit application-level choice and all-grants
  requirement. Q-042 begins existing-grant activation/change handling.
- Concrete-reference checks, removal/rename behavior, existing grants, and
  synchronization/freshness guarantees.
- Administrative scope containment and exact rejection/failure contracts.

Registration-time grant validation and request-time authorization concern
different lifecycle events; they are not two decision stages for one request.
No prepared handoff, new grant field, permission grammar, or implementation
change is introduced by this agreement.
