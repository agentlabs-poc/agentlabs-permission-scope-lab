# Permission — definition and retained naming explanation

## Reading status

The user asked to retain the earlier Markdown's canonical permission explanation.
This chapter restores its detail in the active handbook, aligned with the later
approved scope and endpoint model. The
[original explanation](history/reconciliation-2026-09-05/src/content/authorization-concept.md.txt)
remains unchanged, including its earlier proposals. Retention does not reinstate
deprecated scope formats, additional canonical entities, or runtime claims.

The operation-versus-reach definition is agreed under PERMISSION-001. **Q-056 /
PERMISSION-002 adopts the naming convention with variable namespace depth**:
the user confirmed “this is right” and requested the deeper explanation from
earlier Markdown. The previous pending-grammar status is retained as history:
Q-056 was proposed because PERMISSION-001 had explicitly left grammar open.

## Canonical meaning — PERMISSION-001

A permission identifies an operation on an application-defined resource type.
It names **what can be done**, not **who can do it** or **which data they can
reach**. A permission string alone is therefore not complete authority.

For example, reading one's own payroll ledger and reading Finance payroll
ledgers use the same permission when the underlying operation is the same.
The different reach comes from each grant's scope, not a different read name.

| Operation name, using the retained naming convention | Grant scope fragment | Reach, where the application supports the key |
|---|---|---|
| `hrms:payroll:ledger::read` | `{"user":"$self"}` | The authorizing human's own ledger boundary. |
| `hrms:payroll:ledger::read` | `{"dept":"FIN"}` | Finance ledger boundary. |
| `hrms:payroll:ledger::read` | `{}` | No narrower scope restriction within the trusted tenant. |

These are comparisons, not complete grant schemas. Recipient, validity,
membership/delegation dependencies, and other mandatory restrictions still apply.
Permission and scope remain bound within each complete grant; a wider read grant
cannot supply reach to a narrower write grant.

Rationale: separating operation from reach avoids multiplying permission names
for every employee or department. The application can keep a stable operation
catalog while grants express changing access boundaries. Self, department IDs,
and tenant reach do not belong in the permission name.

Counterexample: naming a permission “read my ledger” and treating it as proof of
ownership does not enforce self access. The endpoint must bind authorization to
the actual data it returns. Likewise, a matching Finance grant and request value
does not prove a certificate belongs to Finance; actual execution must enforce
that boundary under CONTRACT-012.

## Canonical naming explanation — Q-056 / PERMISSION-002, agreed with depth

The earlier handbook described this convention:

```text
<namespaced-noun>::<verb>
```

Everything before `::` names the application-defined noun; everything after it
names the operation. Colons within the noun organize its namespace:

```text
hrms:payroll:ledger::read
hrms:payroll:ledger::post
hrms:payroll:salary_earning::create
agentforge:repository::write
```

In the first example, `hrms` identifies the application, `payroll` the domain,
`ledger` the resource type, and `read` the operation. Namespace depth may vary;
this explanation does not require every application to have a department or an
identical domain hierarchy. Namespace organization avoids ambiguous operation
names across applications; it does not itself grant access or establish trust.

### Namespace depth — restored from the earlier Markdown

The earlier “Resource-type depth” section used:

```text
<application>:<domain>:<resource>[:<subresource>...]::<verb>
```

It also included shorter names such as `hrms:payroll::read`. The variable-depth
convention can therefore be explained using the user's domain terminology as:

```text
<app>:<domain>[:<subdomain-or-resource>...]::<verb>
```

Square brackets mean optional additional namespace segments in this explanatory
notation; they are not literal characters in a permission name. Use `:` between
namespace segments and retain `::` before the final verb, as in the earlier file.
The user requested more depth, not a change to that separator.

```text
hrms:payroll::read
hrms:payroll:ledger::read
hrms:payroll:ledger:entry::read
hrms:payroll:ledger:entry:attachment::read
```

The application defines what its deeper segments mean: domain, subdomain,
resource, or subresource. Auth does not hardcode these meanings or require every
application to have the same number of levels. In the final example, `hrms` is
the application, the noun is `payroll:ledger:entry:attachment`, and `read` is the
operation. A name like `salary_earning` is a single segment, whereas `ledger:entry`
has two. No numeric maximum depth or character-validation rules are selected here.

Rationale: real application vocabularies have different depths. A fixed
app/domain/resource-only shape would flatten meaningful distinctions or require
artificial names. Variable depth retains a common readable convention without
making Auth interpret the application's domain structure.

Naming depth, application data relationships, and scope reach remain different
things. Names describe operations; they do not establish facts about a particular
ledger or entry. Deeper naming alone does not settle permission inheritance.

The original parent/child inheritance and wildcard examples were explicitly
working proposals, not adopted rules. They remain in the archive, not the active
canonical contract. Their v1 treatment is a separate open decision. No wildcard
support follows from retaining namespaced permission examples.

### Why retain this explanation?

The namespace and operation separator make permission examples understandable
across HRMS, code hosting, and other applications. A shared convention could make
catalogs easier to review than unrestricted application-specific spellings.
The alternative is to let each application choose any registered identifier;
that keeps naming flexible but sacrifices a common readable structure.

**Q-056 — answered yes, with deeper namespace explanation:** retain
`<namespaced-noun>::<verb>` with variable-depth application namespaces. Detailed
character validation, namespace ownership, catalog
evolution, aliases, and hierarchy/wildcard behavior are not bundled into this
decision. Existing shorthand examples are not silently renamed by this approval;
their full naming reconciliation remains an editorial follow-up.

### Q-057 / PERMISSION-003 — parent names do not automatically grant child permissions

Status: **PROPOSED, not approved.** Recommend that granting
`hrms:payroll:ledger::read` not automatically grant
`hrms:payroll:ledger:entry::read`. These name separate permissions unless an
explicit, separately agreed authority mechanism supplies both; ordinary roles
can already list both permissions.

Rationale: adding a more specialized operation later should not silently expand
existing access simply because its name shares a prefix. The alternative,
automatic inheritance, shortens broad assignments but makes namespace changes
affect existing authority. The earlier Markdown proposed the non-inheritance
rule but explicitly left its adoption open.

This question does not settle wildcard support, permission aliases, or the
application-level effects covered by a particular registered operation. Scope
and actual-use enforcement remain mandatory for any granted permission.

## How permission participates in authorization

Applications register supported permissions with Auth (REGISTRATION-001). Auth
validates registered references without interpreting application business meaning.
The application supplies that meaning and the endpoint's declaration.

Each protected method/route declares exactly one required permission
(CONTRACT-008); the HTTP method alone is insufficient to identify every business
operation. Grants can supply multiple permissions, and a human can have multiple
direct and group-derived grants. The evaluator considers complete applicable
authority routes without mixing their fields. The endpoint enforces the resulting
boundary against actual output and effects. See [grants](grant-model.md),
[scope](scope-model.md), and [endpoint policy](endpoint-policy-format.md).
