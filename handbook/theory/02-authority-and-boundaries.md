# 2. Authority, permissions and boundaries

[Contents](../README.md) · [Previous](01-foundations.md) · [Next](03-relationships-and-delegation.md)

**Definitions and JSON alongside this chapter:** [permission](canonical-terms.md#permission),
[scope and self](canonical-terms.md#scope-and-self),
[grant, assignment and role](canonical-terms.md#grant-records),
[version and adoption](canonical-terms.md#version-and-adoption).
The reference labels our canonical representations separately from general reasoning.

## Permission identifies an operation

A permission names what may be done. Reading a payslip, approving an accounting
entry and writing repository content are different operations. A permission name
does not, by itself, identify who may perform the operation or which records it
covers.

Separating operation from reach avoids creating a new operation name for each
employee, department or repository. “Read payslip” can be the same operation
whether its permitted reach is one person's payslips or an entire department's.
The boundary supplies that distinction.

Good names help people review a catalog, but naming structure is not authority.
A shared prefix need not imply inheritance. If a system does give a permission
prefix special meaning, that behavior is a policy choice which must be explicit.
Part II's convention does not infer access from shared prefixes.

## Scope describes permitted reach

This handbook uses **scope** to mean a selector of an authority boundary. It
answers “within which limits?” rather than “which operation?” Other systems may
use the same word differently; readers should check the local contract instead
of assuming that every field named scope has identical semantics.

For a payslip-read operation, a boundary might select records belonging to an
employee or department. A system's surrounding context may already impose a
tenant boundary. Combining these limits gives the effective reach; an inner
restriction cannot erase a mandatory outer boundary.

The scope definition owns the meaning of its selectors. A department identifier
has meaning because the application defines the department relationship, not
because a generic grant engine recognizes a familiar word.

| Authority statement | Operation | Boundary |
|---|---|---|
| Vinay may read his own payslips | Read payslip | Payslips belonging to Vinay |
| A payroll reader may read Finance payslips | Read payslip | Finance department |
| A repository maintainer may write one repository | Write repository content | The identified repository |

These are conceptual statements, not a proposed wire format.

## A grant and its assignment answer separate questions

A **grant** describes authority: operations together with their boundary and
applicable restrictions. An **assignment** connects that authority to a recipient.
This separation is useful when the same authority definition can be reused.
Systems may store the connection in different ways; the conceptual distinction
remains valuable when examining access.

A definition in a catalog does not prove that anyone receives it. Conversely,
an assignment cannot supply more authority than the definition and its required
support permit. A membership may make a group's assignment applicable to a
human, but it does not become an additional unrestricted grant.

A **role** can organize a bundle of permissions for reuse. The bundle does not
necessarily define membership, scope or authority to distribute it. In a design
where a grant selects a role, the grant's boundaries still matter. Treating
“has the role” as the whole authorization explanation can hide those limits.

## Evaluate complete authority routes

Suppose Maya has these two independent grants:

| Grant | Permission | Scope |
|---|---|---|
| Finance writer | Write | Finance |
| Engineering reader | Read | Engineering |

They do not establish Engineering write. Taking write from one and Engineering
from the other manufactures a combination that neither grants. Permissions,
boundaries and restrictions must remain associated with the route that supplies
them.

Multiple complete routes may legitimately authorize different operations or
different portions of an explicitly defined batch. That is different from
mixing fragments within a route. A system must define its combination semantics:
alternative grants, mandatory outer limits and any explicit prohibitions cannot
be collapsed into an unordered bag of fields.

The model in Part II uses alternative complete positive routes and does not
support explicit deny grants in v1. This is a chosen combination policy, not a
claim that other authorization systems cannot have prohibitions.

## Narrowing authority

A dependent-authority model constructs a child within a parent's ceiling.
For the model explored here, the relationship is:

```text
Child operations are selected from the parent's effective operations.
Child boundary is the parent boundary AND additional child restrictions.
```

This is a useful construction because it preserves the parent's restrictions
rather than requiring an arbitrary proof that one independently written query
contains another. Equality can be permitted: no additional restriction means
the inherited boundary remains unchanged, not that all restrictions disappear.

Conjunction is not overwriting. If an exact-department predicate requires FIN,
adding a predicate requiring ENG does not replace FIN. Under that meaning, both
cannot hold for the same record. An implementation must not erase one constraint
while flattening the result into an object.

The construction still needs an established parent context. If the parent has
several possible effective versions or supporting relationships, the system must
know which ones are eligible. The algebra of narrowing cannot decide that
relationship on its own.

## Using authority and distributing authority

Being allowed to read Finance payslips is not the same as being allowed to give
another person access. Being allowed to administer a group is not the same as
being allowed to assign arbitrary business authority to it.

Authorization administration therefore raises two questions: may this actor
perform the administrative operation, and is the authority being distributed
within the allowed distribution boundary? A system must define that distribution
boundary. Part II chooses parent-supported issuance, where both checks are
required. Other designs should state their own rule explicitly rather than
infer it from the word administrator.

**Chapter takeaway:** a useful authorization explanation preserves operation,
reach, recipient and support together. Do not let convenient packaging erase
the relationships that make the authority valid.

[Next: relationships, inheritance and delegation](03-relationships-and-delegation.md)
