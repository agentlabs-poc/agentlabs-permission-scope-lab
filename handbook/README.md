# Handbook of Authorization

**First manuscript draft · 8 September 2026**

Authorization connects a person or program's authority to a specific operation
and its effects. A permission check is part of that connection, but it cannot
stand alone: identity must be established, the permitted boundary must be known,
supporting authority must remain usable, and execution must respect the result.

This handbook explains those ideas and shows how to put them into practice.
It is written for application developers, platform engineers, architects and
reviewers who need a shared vocabulary for reasoning about access.

## Two parts, two kinds of statements

**Part I — Foundations: concepts and canonical representations** explains the
core vocabulary, design principles, relationships and failure patterns. The
narrative is a general guide; the accompanying concept reference places our
approved JSON beside each relevant definition, with field explanations and
examples. Framework-specific choices are labeled, not presented as universal
authorization theory. Concepts without separate records and formats still
pending are identified explicitly.

**Part II — Implementation guide** applies that reasoning to the dependent-grant
model developed in this repository. Its rules are choices of this framework,
not universal requirements for every authorization system. “Implementation” here
means applying the canonical records through service responsibilities and integration—not a
claim that a working Auth engine or SDK is supplied or verified.

For example, separating identity from authority is a broadly useful distinction.
Human-only authorization groups, one permission per endpoint, flat string-valued
scopes and human-dependent service accounts are specific choices identified in
the Foundations reference and applied in Part II. A reader can use the general
reasoning without adopting those choices. Separating theory from implementation
does not mean postponing a concept's JSON until the implementation chapters.

## Reading order

| Chapter | The question it answers |
|---|---|
| **Part I — Foundations: concepts and canonical representations** | |
| [1. Foundations](theory/01-foundations.md) | What is authorization, and what makes an authorization boundary trustworthy? |
| [Foundations reference: canonical terms and JSON](theory/canonical-terms.md) | What does each core term mean, why does it exist, and how is it represented? |
| [2. Authority, permissions and boundaries](theory/02-authority-and-boundaries.md) | What can an actor do, within which limits, and through which complete authority route? |
| [3. Relationships, inheritance and delegation](theory/03-relationships-and-delegation.md) | How does authority reach people and programs, and what keeps it supported? |
| [4. Resolution, decisions and enforcement](theory/04-resolution-and-enforcement.md) | How do evidence and rules become a decision that actually constrains execution? |
| **Part II — Implementation guide: the repository's model** | |
| [5. Canonical records and identity](implementation/05-canonical-model.md) | How are identity, adopted content and live controls assembled into usable authority routes? |
| [6. Registration, bootstrap and Auth administration](implementation/06-auth-service.md) | How is authority established and safely distributed inside Auth Service? |
| [7. Application integration](implementation/07-application-integration.md) | How does an endpoint declare, evaluate and enforce authorization? |
| [8. Lifecycle and worked scenarios](implementation/08-lifecycle-and-scenarios.md) | What happens as authority changes, and how does the model apply across applications? |

Start at Chapter 1 and use the Foundations reference alongside Chapters 1–4.
Its term index covers identity, authority, relationships, lifecycle, requests,
resolution, decisions and responsibility layers. Each term's representation is
classified as approved core JSON, a nested value, a concept without a separate
record, or a pending format. For an endpoint integration task,
read Chapters 2, 4, 5 and 7. For grant administration, read Chapters 3, 5, 6 and 8.
Reviewers should follow both flows: authority entering the system and an actual
operation consuming it.

## How to read examples

Maya administers authority, Nutan receives access, and Vinay sometimes acts
through an agent. Team1 and Team2 illustrate explicit team and grant relationships.
FIN is an application department boundary, not another name for a team. A
certificate C17 or payslip P17 is application data, not a new canonical entity.

Every published framework-contract example carries a string `version`. Blocks
are labeled when they are only approved core shapes or excerpts. A format version
does not mean that all validation, transport or recovery behavior has been
settled. Example identifiers do not imply that records have been created in a
running system.

## Draft status and unresolved decisions

The manuscript presents approved behavior and explanatory reasoning. It does
not settle remaining decisions by omission. Relevant chapters include a pending
notice and link to [the pending register](appendices/pending.md). “Pending” means
unresolved and deferred from this writing pass—not approved, implemented or
excluded from the intended v1 scope.

The [source map](appendices/source-map.md) links framework rules to their recorded
decisions and rationale. Existing discussion chapters and archives remain the
source material; this manuscript does not overwrite them or silently promote a
historical proposal. The current website is not replaced by this draft.

[Editorial plan and coverage](editorial-plan.md) ·
[Pending decisions](appendices/pending.md) ·
[Decision sources](appendices/source-map.md)

## Presentation companion

[Core Concepts & JSON — editable PowerPoint](slides/authorization-core-concepts.pptx)
explains the vocabulary in 30 slides with JSON, diagrams and presenter notes.
Revision/adoption mechanics are intentionally left aside; required fields remain
in the examples. See the [presentation guide](slides/README.md) for previews,
sources and representation caveats.
