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
| 1. Foundations | Identity/authority distinction; scope ownership; responsibility layers; authorization versus business/audit scope. | [Principle catalog](../../docs/principle-catalog.md), [vocabulary](../../docs/authorization-vocabulary.md), [system responsibilities](../../docs/system-overview.md). |
| Foundations reference: canonical terms and JSON | WRITING-002 explanatory coverage; existing identity, scope, record, lifecycle, endpoint and result decisions. No pending format is promoted into a schema. | [Reference with per-section sources](../theory/canonical-terms.md), [consolidated glossary](../../docs/identity-glossary.md), [current record reference](../../docs/grant-record-reference.md), [contract inventory](../../docs/grant-contract-closure.md). |
| 2. Authority and boundaries | PERMISSION-001; SCOPE-006/007/008; complete routes; Q-090 recipient separation; Q-093 distribution checks. | [Permission](../../docs/permission-model.md), [scope](../../docs/scope-model.md), [grant model](../../docs/grant-model.md), [assignment authority](../../docs/assignment-authority.md). |
| 3. Relationships and delegation | Q-091/095 team and grant lineage; Q-099 ownership separation; Q-101 stored/effective distinction; Q-070/127 delegation. | [Lineage](../../docs/authority-lineage.md), [ownership](../../docs/ownership-lineage.md), [bindings](../../docs/parent-grant-bindings.md), [delegation](../../docs/delegation-lifecycle.md). |
| 4. Resolution and enforcement | Q-047/048 request and resolved authority; Q-050-C/D enforcement; Q-051 evaluation failure; Q-074/110/128/129 time boundaries. | [Endpoint authorization](../../docs/endpoint-authorization.md), [endpoint policy](../../docs/endpoint-policy-format.md), [results](../../docs/decision-results.md), [concurrency](../../docs/concurrent-enforcement.md). |
| 5. Canonical records and identity | Q-034/056–059 scope/names; Q-107/118 records and role variant; Q-086/087 identity; Q-050-A versions. | [Core grant records](../../docs/grant-revision-format.md), [role variant](../../docs/role-grant-contract.md), [identity](../../docs/identity-context.md), [JWT](../../docs/jwt-identity-mapping.md), [publication](../../docs/contract-publication.md). |
| 6. Registration and Auth Service | Q-039–042 registration; Q-113–117/124 bootstrap; Q-121–123 platform/catalog; Q-093/100/101/103/110 administration. | [Registration](../../docs/application-registration.md), [bootstrap](../../docs/bootstrap-authority.md), [initial setup](../../docs/bootstrap-initial-assignment.md), [root evolution](../../docs/root-permission-evolution.md), [validator](../../docs/authority-boundary-validation.md). |
| 7. Application integration | Q-049 one permission; Q-050-B/C/E/F policy/material; Q-062–067 results; Q-068/074 boundary enforcement. | [System overview](../../docs/system-overview.md), [endpoint policy](../../docs/endpoint-policy-format.md), [results](../../docs/decision-results.md), [operation enforcement](../../docs/operation-enforcement.md). |
| 8. Lifecycle and scenarios | Q-082/083/109 controls and validity; Q-101–106/111 dependencies/revisions; Q-069/128/129 freshness; Q-070/127 proxies; Q-071–075/130 operations. | [Lifecycle](../../docs/grant-lifecycle.md), [revisions](../../docs/grant-revisions.md), [bindings](../../docs/parent-grant-bindings.md), [freshness](../../docs/authority-freshness.md), [bulk](../../docs/bulk-enforcement.md), [source scenarios](../../docs/use-case-examples.md). |

## Example identity and reuse

The Foundations reference and Chapters 5–7 share G1 revision 2 → A1 → Team1,
then G2 revision 1 → A2 → Team2,
with Nutan receiving Team2 authority through membership. The local permission
is certificate-read/write and the application boundary is FIN, with additional
C17 narrowing in G2. G0 and the administrative root have explicitly stated valid
upstream/trusted-establishment premises; they are not fully encoded root examples.

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
