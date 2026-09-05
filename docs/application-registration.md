# Application registration and Auth validation

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
| A scope key is registered for the relevant application. | What target relationship that boundary key represents. |
| Scope follows canonical syntax and declared constraints. | Trusted facts establishing whether a particular target is inside the selected boundary. |

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

## Still open

- Registration representation, APIs, storage, distribution, and versioning.
- Who may register or change an application's definitions, and application
  namespace/identity protection. Registration is not authority to self-grant.
- Declaring supported permission-scope combinations, including multi-permission
  grants and role evolution; Q-040 begins this branch.
- Concrete-reference checks, removal/rename behavior, existing grants, and
  synchronization/freshness guarantees.
- Administrative scope containment and exact rejection/failure contracts.

Registration-time grant validation and request-time authorization concern
different lifecycle events; they are not two decision stages for one request.
No prepared handoff, new grant field, permission grammar, or implementation
change is introduced by this agreement.
