# Authorization vocabulary and request material

Q-090 now distinguishes a reusable recipient-free **grant definition** from its
recipient-bearing **assignment**. A **subteam/subgroup** under Q-091 receives
explicit dependent authority, not transitive membership. See
[grant/assignment meanings](grant-assignments.md) and [subgroups](subgroups.md).
This supersedes TERM-004's previous treatment of grant and assignment as one
logical record, without introducing another request-boundary entity.

TERM-005 / Q-043 is agreed. The user asked to remove an unnecessary term from
the vocabulary, then requested a tenability check before editing. After the
check, the user agreed to describe authorization through permissions, scope
boundaries, requests, and request material, without another mandatory entity.
[Exact deprecated wording](history/q043-vocabulary.md) preserves the earlier
explanations and proposal. This chapter records the reasoning, not just a rename.

## Canonical explanation

> Authorization determines whether the requested operation is permitted within
> the applicable scope boundary, using sufficient trusted request material.

This is a summary, not a replacement for the full authority rules. Grant
applicability, validity, conditions, tenant boundaries, and human-dependent
delegation limits still apply. Permission and scope must remain associated in
each complete grant; material resolution cannot amplify authority.

| Existing concept | What it contributes |
|---|---|
| Grant | The recipient's permission or role binding, scope, validity, conditions, and dependencies. |
| Permission | The operation that authority covers. |
| Scope | The boundary defining that authority's reach within the enclosing tenant. |
| Request | The operation being requested, under the server-owned endpoint declaration. |
| Request material | Identified inputs, verified context, and trusted facts needed for evaluation. |
| Resolution and enforcement | Establish sufficient evidence, decide under complete applicable grants, and bind actual execution to that decision. |

These concepts do not require an additional canonical entity, wrapper, or
record between the request and scope evaluation. Calling a payslip, repository,
or proposed certificate a resource is ordinary application vocabulary, not a
new universal authorization object or mandatory JSON representation.

## Boundary is not the same as the requested work

Suppose an applicable grant permits payslip-read with:

```json
{ "dept": "FIN" }
```

This specifies authorized reach. It does not say which payslip the caller is
requesting. The request identifies P-17, and trusted material establishes the
department relevant to that payslip under the application's defined meaning.
Evaluation checks both permission and boundary. The scope does not replace the
request, and knowing a payslip identifier does not establish authority.

This is why removing the extra term is tenable: the request and its material
already supply the distinction that the earlier explanations were trying to
express. Merely deleting that distinction would not be tenable.

## Case-by-case tenability check

| Operation | Expression using the existing concepts |
|---|---|
| Read P-17 | The request identifies the payslip; trusted material establishes department/ownership; applicable scope requirements are checked. |
| List payslips | The declared listing operation must constrain returned records to the authorized boundary; no single-record abstraction is required. |
| Create a certificate | Material describes the proposed certificate and required relationships; an existing certificate need not exist. |
| Move a certificate | Material distinguishes current department from proposed department; the operation's boundary requirements must be explicitly defined. |
| Create a grant | Material contains the proposed recipient, permissions, scope, and conditions, to be evaluated against administrative authority. |

These are expressibility checks, not completed enforcement designs or tests
against the lab engine. Collection/query, current-versus-proposed state,
multi-record, and administrative containment contracts remain open. The table
does not decide a move policy or introduce administrative scope fields.

## Security counterexamples

Checking ownership of P-17 and returning P-18 is not valid authorization for
the actual read. The evidence and decision must remain bound to what execution
really does. This requirement is already ENFORCEMENT-002; vocabulary changes
do not remove it. Freshness and concurrency mechanics remain to be settled.

Similarly, receiving `dept=FIN` in a path is not proof that P-17 belongs to
Finance. Identified request values and authoritative facts have different
roles. The endpoint declaration and application bindings establish those roles;
matching field names do not prove a relationship (INPUT-001, CONTRACT-007,
and Q-038).

Removing the term therefore does not mean that any bag of matching key/value
pairs is sufficient, or that the handler can ignore which data it reads or
changes. It removes an unnecessary vocabulary abstraction while preserving
trust, boundary evaluation, and actual-use enforcement.

## Effect on the administration branch

The earlier Q-043 proposal relied on an extra canonical abstraction. That
framing is withdrawn, not renamed into a different mandatory object. The
administration discussion can instead ask whether the requested grant creation
or change fits the administrator's authority, using the proposed grant's
material. ADMIN-004/005 and their exact scope representation remain open.

This decision settles the vocabulary correction and its tenability, not a new
administrative grant schema. Registration lifecycle detail remains parked as
the user requested. The current branch is still grant administration.
