# Decision sources and manuscript traceability

[Handbook contents](../README.md) · [Pending register](pending.md)

The manuscript is organized for learning, not in question-number order. This
map keeps the connection to the approved discussion and its rationale. Existing
source files may retain superseded sections; later explicit decisions govern
where their wording differs. Historical examples are not silently promoted into
current schemas.

Part I combines the general narrative with a source-backed canonical concept
reference. Its JSON is explicitly our model's representation, not a universal
requirement. Part II explains how those concepts are used. The manuscript
introduces no new policy merely by supplying definitions or examples.

| Manuscript chapter | Decision foundation | Sources |
|---|---|---|
| 1. Foundations | Identity/authority distinction; scope ownership; responsibility layers; authorization versus business/audit scope. **Unchanged in this recompilation — none of its sources moved.** | [Principle catalog](../../docs/principle-catalog.md), [vocabulary](../../docs/authorization-vocabulary.md), [system responsibilities](../../docs/system-overview.md). |
| Foundations reference: canonical terms and JSON | Explanatory coverage of every core term; namespace ownership and identifier permanence; scope opacity, exact matching and the platform key; administration as a grant; ownership as a grant; the tenant's two namespaces; delegation lifetime and growth; unreachable nodes; dependents blocking disable and delete; trusted correlations and mount-time policy validation; the authority-loading contract; the contributing chain and the code catalogue; the handler integration contract. | [Reference with per-section sources](../theory/canonical-terms.md), [glossary](../../docs/identity-glossary.md), [record reference](../../docs/grant-record-reference.md), [contract inventory](../../docs/grant-contract-closure.md). |
| 2. Authority and boundaries | Permission identifies an operation; scope selects reach; complete routes; narrowing by conjunction; using versus distributing authority. **Recompiled** for scope values being opaque and matched exactly with subtree excluded, and for administrative authority being a kind of thing a design must choose. | [Permission](../../docs/permission-model.md), [scope](../../docs/scope-model.md), [opacity and exclusion](../../docs/application-registration.md), [grant model](../../docs/grant-model.md), [assignment authority](../../docs/assignment-authority.md), [administration is a grant](../../docs/administrative-authority.md). |
| 3. Relationships and delegation | Different edges, different effects; the issuer versus the continuing source; stored state versus effective authority. **Recompiled** for ownership being a grant rather than an inert relation, a delegation's own window and its tracking of the human, unreachable-versus-unreadable, and dependents blocking disable and delete. | [Lineage and unreachable nodes](../../docs/authority-lineage.md), [groups](../../docs/groups-and-membership.md), [ownership](../../docs/ownership-lineage.md), [delegation lifetime and growth](../../docs/delegation-lifecycle.md), [dependents](../../docs/grant-lifecycle.md), [bindings](../../docs/parent-grant-bindings.md). |
| 4. Resolution and enforcement | Request, resolved request and decision; two ways to establish a data boundary; completed rejection versus inability to evaluate; time in the reasoning; evidence explains the route. **Recompiled** for the unusable route, the open-versus-closed code catalogue as a design note, the explicit background deferral, and response disclosure as the endpoint's duty. | [Endpoint authorization and disclosure](../../docs/endpoint-authorization.md), [decisions, catalogue and unusable routes](../../docs/decision-results.md), [freshness](../../docs/authority-freshness.md), [background deferral](../../docs/background-authorization.md), [audit consumers](../../docs/authority-change-audit.md). |
| 5. Canonical records and identity | Records, roles, identity and format versions. **Recompiled** for namespace ownership, identifier permanence, retirement narrowing rather than closing, scope opacity and the platform key, and the pending list that shrank. | [Core records](../../docs/grant-revision-format.md), [permission namespace and permanence](../../docs/permission-lifecycle.md), [scope](../../docs/scope-model.md), [role variant](../../docs/role-grant-contract.md), [identity](../../docs/identity-context.md), [JWT mapping](../../docs/jwt-identity-mapping.md). |
| 6. Registration, bootstrap and Auth administration | **Rewritten.** Registration and namespace ownership; the tenant's two namespaces and why establishment belongs to the first; administration as an ordinary grant, what bounds each side, why sideways escalation is impossible rather than guarded, and the recorded wart; ownership as a grant; membership synchronization as an ordinary caller; a team's administrative dependants; what an audit consumer may rely on. | [Registration](../../docs/application-registration.md), [namespace ownership](../../docs/permission-lifecycle.md), [bootstrap and the two namespaces](../../docs/bootstrap-authority.md), [administration is a grant](../../docs/administrative-authority.md), [the platform scope key](../../docs/scope-model.md), [ownership](../../docs/ownership-lineage.md), [synchronization](../../docs/groups-and-membership.md), [dependents](../../docs/grant-lifecycle.md), [platform authority](../../docs/application-platform-authority.md), [boundary validation](../../docs/authority-boundary-validation.md), [audit consumers](../../docs/authority-change-audit.md), [write consistency](../../docs/auth-write-consistency.md). |
| 7. Application integration | **Rewritten.** The gate contract — four things supplied, one received, and the order; trusted correlations and mount-time structural validation; nested selection and missing inputs refused; the authority-loading question and answer and what a consumer must do with it; the unusable route; boundary-changing operations as one gate evaluating per boundary; the contributing chain and the code catalogue; response disclosure; the three evidence cases; the background deferral. | [Endpoint policy](../../docs/endpoint-policy-format.md), [handler integration contract](../../docs/handler-integration-contract.md), [authority loading transport](../../docs/authority-resolve-transport.md), [decisions and catalogue](../../docs/decision-results.md), [disclosure](../../docs/endpoint-authorization.md), [evidence cases](../../docs/grant-conditions.md), [composition](../../docs/operation-enforcement.md), [background](../../docs/background-authorization.md), [the open boundary question](../../docs/policy-scope-boundary.md), [concurrent enforcement](../../docs/concurrent-enforcement.md). |
| 8. Lifecycle and scenarios | Publication, adoption and resolution as different events; controls, time and effectiveness; the worked scenarios. **Recompiled** for publication being free while adoption is evaluated differentially, dependents blocking disable and delete, and ownership rotation as an authority change rather than a list edit. | [Revisions](../../docs/grant-revisions.md), [role publication and adoption](../../docs/role-revisions.md), [dependents](../../docs/grant-lifecycle.md), [bindings](../../docs/parent-grant-bindings.md), [validity](../../docs/grant-validity.md), [freshness](../../docs/authority-freshness.md), [bulk](../../docs/bulk-enforcement.md), [delegation](../../docs/delegation-lifecycle.md), [ownership](../../docs/ownership-lineage.md), [direct-human trace](../../docs/direct-human-parent-context.md). |

## What this recompilation was

The first manuscript draft was written before roughly two dozen further decisions
were taken. Rather than patch prose whose sources had moved, each chapter was
recompiled against the decision chapters as they now stand. Chapter 1 survives
untouched because none of its sources moved. Chapters 6 and 7 are rewrites rather
than revisions: each gained material that had no place in the draft at all — the
tenant's two namespaces and administration-as-a-grant in Chapter 6, and the gate
contract, the authority-loading wire and the policy's trusted correlations in
Chapter 7.

Approved prose whose source did not move was kept. The editorial direction is
unchanged: general reasoning stays separate from this framework's representations,
and no pending format is promoted into a schema by being explained.

## Example identity and reuse

The Foundations reference and Chapters 5–7 share G0 revision 1 → A0 → Team0 as the
trusted root, then G1 revision 2 → A1 → Team1, then G2 revision 1 → A2 → Team2,
with Nutan receiving Team2 authority through membership. The local permission
is certificate-read/write and the application boundary is FIN, with additional
C17 narrowing in G2. G0's holder and assignment are named because a resolved
authority answer carries the root step of the chain root-first; the root's
*establishment* is still a stated premise rather than an encoded example, and the
administrative root likewise.

Chapter 8's competing FIN/ENG direct-human diagnostic is labeled a separate
scenario. Its G1 revisions are not an instruction to overwrite the running
example's immutable content. Independent payroll, accounting and temporary-grant
examples have separate IDs and state their support assumptions.

The Foundations reference's disabled G2 control is explicitly a later lifecycle
snapshot, not a second simultaneous control or a contradiction of the earlier
enabled-route example. Its temporary grant repeats the existing independent
temporary example. Role references, identity blocks and GET/PUT policies use the
same approved shapes as the implementation chapters.

## Diagram provenance

`authorization-reasoning.svg` is a new generic logical flow. It does not prescribe
the framework's result envelope or deployment. `model-lineage.svg` is a new visual
of the existing parent-team/parent-grant/assignment pattern, not a new canonical
entity. Both are under `handbook/assets/`.

The Auth internal gate, application request flow and shared-catalog diagrams are
linked directly from the existing approved/reconciled SVG assets under `docs/`.
Their use does not modify the originals, replace the website, or declare a
runtime implementation verified.

## Preservation and acceptance

Original decision text, rejected proposals, scratch imports and historical SVGs
remain where they were. The manuscript's clean reading order is not deletion of
that history. A pending marker does not erase a requirement; an example passing
static validation does not prove runtime authorization correctness.

This first draft can be reviewed chapter by chapter while unresolved items remain
in [the pending register](pending.md). Final acceptance and complete contract
publication remain separate from writing the narrative.
