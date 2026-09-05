# Endpoint-completion: candidate cases

These are illustrative tests of the handbook model, not newly agreed permission,
scope, condition, or grant types, and not claims about current implementation.

Related decisions: CONTRACT-002, CONTRACT-003, CONTRACT-004.
Related agreed classification rule: CONTRACT-005 in handbook-roadmap.md.

## Cases beyond self-access

Each case assumes the needed application facts are unavailable to middleware and
the declared integration cannot yet turn the restriction into a complete,
enforceable access predicate. These assumptions matter: the business noun alone
does not determine an endpoint's resolution mode.

| ID | Example request and authority | Application facts needed | Why completion may be needed |
|---|---|---|---|
| EC-001 | Read certificate C-17 under a Finance-department scope. | C-17's actual tenant and department. | A caller-supplied department path segment does not prove the certificate is in that department. |
| EC-002 | Read repository R-8 under a Project A subtree scope. | R-8's actual project and containment path. | An opaque repository ID does not establish ancestry; moves can change it. |
| EC-003 | Read records of the user's direct reports, where that relationship scope is supported. | Trusted user-to-employee mapping and the relevant employee reporting relationship. | A generic Auth group is not assumed to encode the business reporting relationship. |
| EC-004 | Approve an expense under a rule preventing approval of one's own submission. | The expense's actual submitter identity and its mapping to the actor. | Possession of approve permission does not establish that the actor satisfies this separation-of-duties condition. This rule itself remains illustrative. |
| EC-005 | Approve an invoice under an explicitly configured monetary approval ceiling. | The stored authoritative amount and relevant currency/amount basis. | Middleware must not trust a caller's claimed amount as the authoritative invoice amount. Monetary conditions are illustrative, not an adopted grant schema. |
| EC-006 | Move a document from Folder A to Folder B under the declared source/destination access rules. | The document's actual current parent, the destination's tenant/ancestry, and any other facts the declared operation requires. | The decision covers both the actual source and proposed destination. Exact required permissions are not assumed here. |
| EC-007 | Bulk-download a supplied list of certificate IDs under a department scope. | Actual scope-relevant attributes for the selected resources, or a query restriction that establishes permitted membership. | One authorized item cannot justify all items. All-or-nothing versus partial-result behavior remains a separate decision. |

## Cases that do not automatically require endpoint completion

- An application database read merely to retrieve data already restricted by a
  complete authorization predicate is enforcement, not automatically additional
  authorization resolution.
- For a tenant-scoped operation whose supported authorization is completely
  established in middleware, applying the authorized tenant restriction in the
  repository is still mandatory but need not imply prepared mode.
- An object's existence or a business rule applicable to everyone, such as a
  malformed document being unpublishable, is not automatically an unresolved
  authorization rule. A rule that varies access by the actor or their authority
  can be an authorization condition and needs to be declared accordingly.
- Conditions in the table might be expressible as mandatory predicates under a
  supported implementation. If no authorization condition remains unresolved,
  their endpoint may qualify for middleware-complete mode. Conversely, merely
  expressing a condition in SQL does not by itself establish that it is fully
  resolved; the operation contract must make the distinction explicit.

## Mode-selection test (CONTRACT-005, agreed)

For every authorization form supported by the endpoint, can middleware establish
the decision and complete mandatory enforcement restrictions using its trusted
inputs and declared evaluator contracts?

- If yes, middleware-complete mode is a candidate; the handler still enforces.
- If application facts are required to finish deciding the authorized access,
  use endpoint-completion mode for the whole endpoint.

Keep endpoint mode fixed as agreed in CONTRACT-002. These cases do not justify
dynamic mode switching or independent handler-specific policy semantics.

No endpoint-coverage percentage has been measured. Determine coverage through a
later audit of actual operations and supported scopes.
