# Editorial plan — Handbook of Authorization

**WRITING-001 · Approved writing direction, not new authorization policy**

The manuscript lives at repository-root `handbook/`, as requested. Its purpose
is to explain authorization in a readable progression while preserving a clear
separation between general theory and this repository's concrete model. Existing
`docs/` chapters remain intact as decision sources, including superseded history.

## WRITING-002 — approved Foundations correction

The user clarified that Foundations must explain all core canonical terms and
their JSON, then approved the proposed correction. The first draft's introductory
foundation and later-only record chapter did not fully meet that purpose.

Foundations now includes Chapters 1–4 plus the indexed
[canonical concept reference](theory/canonical-terms.md). Every concept family
has a definition, purpose/rationale, relationships and rules, representation
status, and example/counterexample. Approved core JSON appears beside the
concept with field explanations. A nested value is not invented as a standalone
record; a pending wire format stays pending. Implementation chapters explain
how services, endpoints and evaluators use these concepts.

This refines WRITING-001's placement of JSON, not the approved authorization
model. Generic reasoning remains distinct from our framework's representations.
The original Chapter 5 examples are retained as an integration walkthrough and
cross-checked against the new reference, rather than deleting prior explanations.
Existing decision documents, historical proposals and diagrams remain unchanged.

## Narrative and content plan

| Stage | Content to establish | What the reader can reason about afterward |
|---|---|---|
| 1. Foundations | Identity versus authority; trust boundaries; least privilege; failure behavior; responsibility versus mechanism. | Why authentication alone cannot authorize a protected effect. |
| Foundations reference, alongside 1–4 | Canonical vocabulary with approved JSON, field meanings, relationships, rationale, counterexamples and explicit representation gaps. | Recognize what each term means and how it is represented before implementing it. |
| 2. Authority and boundaries | Permission, scope, grant, assignment, role; complete routes; conjunction and alternatives; administrative authority. | Whether an operation is supported without combining unrelated permission and scope fragments. |
| 3. Relationships and delegation | Distinct relationship types; parent-dependent authority; live support; ownership versus issuance; explicit versus effective state. | What should and should not change when a member, owner or supporting relationship changes. |
| 4. Evaluation and enforcement | Request claims, trusted material, resolution, decision, enforcement; lookups versus constrained execution; time and evidence. | Why a successful check is insufficient if actual output escapes its boundary. |
| 5. Concrete model | Framework choices; permission grammar; scope rules; grant/control/revision/assignment/role records; identity and JWT. | How to read current versioned examples without treating excerpts as full schemas. |
| 6. Auth Service | Registration, compatibility mode, shared catalog, bootstrap, administrative and source-boundary gates; Maya's worked assignment. | How legitimate initial authority becomes bounded distributed authority. |
| 7. Applications | Endpoint declaration; path/body sources; GET and PUT; decision variants; actual-data enforcement; failure handling. | How to design and review an authorization-aware endpoint. |
| 8. Changes and scenarios | Revision/adoption, disablement, time validity, orphans, structural guards, freshness; HRMS, code hosting, ticketing and accounting. | How the same model behaves across state changes and application domains. |

The progression is deliberate: readers encounter a concept with its relevant
JSON and field explanation, relationships before their lifecycle consequences,
and decision semantics before an endpoint implementation. The final chapter
reconnects those ideas through scenarios instead of adding another vocabulary.

## Chapter writing pattern

Each chapter provides definitions in ordinary language, a worked explanation,
the reason for the distinction, at least one failure pattern or counterexample,
and a short transition to the next chapter. Foundations and Part II both identify
framework-specific rules and link their decision sources. Pending decisions are called out near
the affected behavior, not hidden only in an appendix.

No requirement is invented simply to make an example complete. In particular,
there is no new direct-human parent selector, computed-root wire field,
relationship block in endpoint policy, generic condition engine, owner-transfer
permission or automatic support-repair rule.

## Diagram plan

| Diagram | Purpose | Location |
|---|---|---|
| General authorization flow | Separate established identity, authority/evidence, decision and constrained execution; no mandated Auth service. | `assets/authorization-reasoning.svg`, Chapter 4 |
| Model lineage | Show grant parentage and team parentage connected by assignments, with membership and administration distinguished. | `assets/model-lineage.svg`, Chapters 3 and 6; explicitly labeled model-specific in Chapter 3 |
| Auth request through both internal checks | Explain administration versus proposed-authority validation before persistence. | Reuse `../docs/assets/auth-service-authority-gate.svg`, Chapter 6 |
| Application request flow | Follow client, authentication, endpoint, embedded evaluation, authority loading and constrained data access. | Reuse `../docs/assets/authorization-system.svg`, Chapter 7 |
| Shared catalog and isolated roots | Explain the chosen catalog-to-root behavior without combining tenant authority. | Reuse `../docs/assets/shared-root-catalog-flow.svg`, Chapter 6 |

Reusing approved SVGs preserves their established meaning and avoids maintaining
competing system diagrams. New SVGs use native vector text/shapes, accessible
titles/descriptions and readable labels. The book links repository assets; it is
not yet a standalone export bundle.

## Validation and delivery checklist

- [x] Write the eight substantive chapters and introductory reading guide.
- [x] Keep general theory distinct from framework-specific policy throughout.
- [x] Include versioned core examples, rationale, counterexamples and scenario outcomes.
- [x] Preserve pending items with source links and explicit local caveats.
- [x] Add or reuse the planned SVGs and inspect new diagrams visually.
- [x] Check Markdown links, JSON parsing, SVG XML, navigation and chapter coverage.
- [x] Run the existing reader build/tests as a repository regression check.
- [x] Verify existing decision documents and archived sources were not rewritten.

These are manuscript-delivery checks, not new policy decisions or evidence of a
working authorization implementation. A complete first draft is distinct from
final handbook acceptance. No commit, push, site replacement or runtime migration
is included in the writing request.

### First-draft verification — 8 September 2026

This is the preserved verification snapshot before WRITING-002's expansion;
its file/example/link counts are not current totals.

The manuscript has eight chapters, a reading guide, this plan and two appendices.
Checks parsed all 12 Markdown files and 18 versioned JSON examples, verified
123 local file links, checked the running G1/A1 and G2/A2 revision references,
and confirmed five distinct SVG references. Both new SVG files pass XML parsing.
Initial browser inspection of both new diagrams prompted shorter input labels
and repositioned lineage labels; a repeat screenshot attempt stalled, so the
final label adjustments were checked in source rather than in a fresh capture.

The existing production build succeeded and all 10 reader tests passed. Those
are website regression checks, not authorization-engine tests or a claim that
the manuscript has been integrated into the website. Existing `docs/`, `src/`
and `tests/` files were unchanged, and the root README and working-instruction
edits are additive. Final design acceptance and the pending contracts remain
separate from these first-draft delivery checks.

### WRITING-002 verification — 8 September 2026

- [x] Add the Foundations concept reference with an indexed vocabulary and
  definitions, rationale, relationships, representation status and counterexamples.
- [x] Put approved JSON and field explanations beside the concepts; distinguish
  nested values, concepts without records and formats still pending.
- [x] Connect the four narrative chapters to the relevant reference sections and
  explain the handoff to implementation without removing earlier examples.
- [x] Cross-check repeated immutable grant examples and assignment selections;
  preserve existing decision sources, runtime files and diagrams.

Fresh checks parsed 13 Markdown files and 35 versioned JSON examples, verified
221 local links including 38 fragment links, and confirmed consistent repeated
content for eight distinct immutable grant examples. A 73-term presence inventory
supplemented the source-backed editorial review; a text match alone is not proof
of explanatory completeness or schema approval. Both existing manuscript SVGs
still pass XML parsing and no diagram content changed in this expansion.

The production build and all 10 existing reader tests passed again. These remain
repository regression checks, not tests of a deployed Auth engine or website
integration of the manuscript. The work remains uncommitted; the pending register
and final handbook acceptance are not closed by this editorial correction.
