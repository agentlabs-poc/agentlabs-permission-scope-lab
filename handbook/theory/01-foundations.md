# 1. Foundations

[Contents](../README.md) · [Next: authority and boundaries](02-authority-and-boundaries.md)

## What Foundations must establish

Foundations includes both the conceptual narrative in Chapters 1–4 and the
[canonical terms and JSON reference](canonical-terms.md). Its purpose is to
explain every core concept, not merely introduce the system before sending the
reader elsewhere for definitions.

For each term, the reference explains meaning, purpose, relationships, rules,
representation and an example/counterexample. Approved JSON is shown next to
the concept and its fields are explained. A concept without its own record is
identified as such; an unsettled format is explicitly pending, not invented.

Use this learning sequence:

1. [Identity and enclosing context](canonical-terms.md#identity): who acts, whose
   authority supports the request, and which outer boundary applies.
2. [Permission, scope and registration](canonical-terms.md#permission): the
   operation, permitted reach, and application-owned meanings.
3. [Grants, assignments and roles](canonical-terms.md#grant-records): authority
   content, recipients and explicit content adoption.
4. [Teams and dependent relationships](canonical-terms.md#teams-and-administration):
   membership, ownership, lineage, delegation and continuing support.
5. [Lifecycle and roots](canonical-terms.md#lifecycle): enabled versus effective,
   validity, orphans and trusted initial authority.
6. [Requests through enforcement](canonical-terms.md#endpoint-declaration):
   endpoint policy, material, resolved views, decisions and actual effects.

The generic explanation remains useful independently of this framework. JSON
in the reference describes our chosen representation, not a universal standard.
Implementation chapters then explain how services and endpoints use it.

## Authorization connects authority to an effect

Suppose a payroll application receives a request to read a payslip. Recognizing
the caller is only the beginning. The application must establish whether that
caller may perform this kind of read, whether the particular data falls within
the caller's permitted reach, and whether the response actually stays within
that reach. Authorization is the reasoning and enforcement that connect these
requirements to the protected effect.

An effect can be a returned document, a changed row, a created account, an
exported total or a grant assigned to another person. Read access matters even
when nothing is modified: revealing information is itself a protected effect.
Administration matters because it changes what later operations can do.

A useful review question is therefore not only “Was permission checked?” It is:
**What authority justifies this exact effect, and where are its limits enforced?**

## Authentication and authorization answer different questions

Authentication establishes an identity under a trust arrangement. Authorization
determines what an established actor is entitled to do in the relevant context.
A recognized identity does not supply every permission, and a document claiming
to contain permissions does not authenticate whoever presents it.

Consider Vinay and a program acting on his behalf. A system may need to know both
who actually sent the request and whose authority supports it. Collapsing those
identities can hide restrictions on delegated access. The precise credential
format is an implementation matter; the distinction between caller and authority
source is part of the reasoning.

Similarly, a tenant identifier in a URL is a request claim. It must be reconciled
with trusted context. A caller should not acquire a different administrative
boundary merely by substituting another tenant's identifier.

## State the boundary before choosing the mechanism

An authorization design needs an answer to four questions:

1. **Operation:** what protected action is being requested?
2. **Reach:** which data or effects may that authority cover?
3. **Support:** why does this actor have that authority now?
4. **Enforcement:** how will execution remain within the established limits?

These questions are independent of whether the implementation uses tables,
policy documents, a central service or embedded evaluation. A central service
can make decisions consistently while an application still returns the wrong
row. An embedded evaluator can be close to the data while consulting stale or
untrusted authority. Deployment location does not prove correctness.

## Principles that make the design explainable

A principle states a design constraint, such as not treating missing evidence
as authority. A policy states the applicable rules for particular operations
and contexts. An implementation realizes those rules through records, evaluation
and enforcement. Keeping these levels separate helps distinguish an architectural
requirement from one possible storage or API choice.

### Grant the authority needed for the responsibility

Least privilege is a design discipline: establish the operations and reach a
responsibility needs, and avoid unrelated authority. It is not the rule that
every grant must be smaller than another by an arbitrary amount. An administrator
may legitimately need broad authority within an explicitly established domain.
The important questions are why that authority exists and how its boundary is
maintained.

### Do not manufacture authority from missing information

If a required check cannot be completed, the absence of a negative answer is not
permission to proceed. A timeout is not evidence that the caller is entitled to
access. It is also not proof that no entitlement exists. Distinguishing failure
to evaluate from a completed rejection improves diagnosis without allowing an
unchecked operation.

### Make authority-changing actions explicit

Changing membership, assigning a grant and adopting new authority can have
different effects. Treating them as interchangeable “user updates” makes it hard
to explain who gained access and why. A design should identify which operations
can change authority and what bounds those operations must respect.

### Keep decisions connected to execution

An allow answers a particular authorization question. It does not authorize a
later operation on different data. If a check covers Finance records, the
subsequent query cannot become an unrestricted lookup merely because the check
has already passed. The same requirement applies to mutations and derived output.

### Preserve distinctions rather than hiding them in names

A team named Finance need not be the Finance department. Someone called an owner
need not automatically possess every business permission. An enabled record need
not supply effective access if its required support is missing. Names help people
read the model, but explicit relationships and rules establish authority.

## General guide versus chosen model

Different systems can meet these principles with different representations.
Some model authority through permission bundles, others through attributes or
relationships, and many combine these techniques. This book does not claim that
one storage format is a universal theory of authorization.

The Foundations reference documents one concrete dependent-grant vocabulary and
its approved representations; Part II explains their use. Its human-only groups,
human-dependent automated accounts, explicit grant revisions and endpoint policy
shape are deliberate choices. Understanding their rationale is useful even when
another implementation makes a different choice.

The subject here is authorization. Establishing a record's department and
constraining a read belong in that subject. Designing payroll calculations,
approval workflows or an external audit-retention system does not. Those systems
may consume authorization outcomes without becoming part of the authority model.

**Chapter takeaway:** begin with the protected effect and its authority boundary.
Credentials, services, roles and tables are ways to implement that reasoning,
not substitutes for it.

[Next: authority, permissions and boundaries](02-authority-and-boundaries.md)
