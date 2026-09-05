# Handbook reconciliation register

## Checkpoint following 675b434 — Q-044 administration rules

Q-044 approves ADMIN-004/005 at the governing-rule level: ordinary grant model,
explicit administrative operation, complete requested-assignment checks,
associated bounds, and normal validation. The grant chapter preserves the
Finance JSON example and all counterexamples; no administrative scope fields
or containment algorithm are adopted. Earlier proposed-status text is retained
with current qualification. Q-045 is open for group-policy consolidation.

Docs-only checks passed: `git diff --check`, 22 Markdown files, 131 local links,
44 JSON blocks, 120 unique decision/question references in the tree, and all
five approved rules present in the chapter. No application source changed.

## Checkpoint following c631a3c — Q-043 vocabulary and explanation audit

TERM-005 records the approved vocabulary correction after the user's tenability
check. Current chapters and the SVG use permissions, boundaries, request material,
and ordinary application facts without an extra canonical entity. The earlier
administration framing is withdrawn, not approved under another name.

[Detailed reasoning](authorization-vocabulary.md) includes five operation cases
and security counterexamples. The [explanation audit](explanation-audit.md)
checks actual chapter coverage across recent discussions and identifies open
work. [Exact superseded wording](history/q043-vocabulary.md) and the prior SVG
preserve history; the earlier lab pages remain explicitly historical.

ADMIN-004/005 remain open and registration lifecycle detail stays parked. This
checkpoint does not change grant JSON, application behavior, or security rules.

Verification: the documentation scan passed for 22 Markdown files, 130 local
links, 43 JSON blocks, and 118 unique decision/question references present in
the tree. Fourteen current chapters and the current SVG contain none of the
retired vocabulary. Changed original Markdown lines remain in their source or
the deprecated record; the old SVG is preserved byte-for-byte. `git diff --check`
and `npm run build` passed. The updated SVG was rendered and visually checked.
These checks support the documentation checkpoint, not production authorization
correctness or completion of all handbook branches.

## Checkpoint following 61a8967 — Q-042 and branch return

Q-042 agrees REGISTRATION-004: existing grants must pass before relationship
validation is enabled; otherwise reject activation, report incompatibilities,
and preserve the prior configuration pending explicit authorized correction.
User directed ending this detour. Remaining registration lifecycle mechanics
are parked, not finalized. Discussion returns to stage 5, ADMIN-004/005, starting
with Q-043's canonical administration model. No withdrawn admin fields or new
administration rules are adopted by moving branches.

Docs-only checks passed: `git diff --check`, 19 Markdown files, 96 local links,
42 JSON blocks, and 117 unique decision/question references in the tree.
REGISTRATION-004 is AGREED; Q-043 is OPEN. No application source changed.

## Checkpoint following edd7635 — refined Q-041 approval

REGISTRATION-003 records the explicit application-level choice approved in
Q-041: enabling relationship validation makes it mandatory for every grant,
including existing and role-based grants. Disabled mode retains other Auth
validation and runtime scope enforcement. Missing relationship metadata cannot
silently turn off an enabled mode.

The earlier omission-based Q-041 proposal is preserved and explicitly
superseded; it is not the approved rule. No field name, missing-choice default,
automatic grant mutation, or revalidation algorithm is adopted. Q-042 remains
open: enabling validation when existing grants are incompatible.

Docs-only verification passed: `git diff --check`, 19 Markdown files, 96 local
links, 42 JSON blocks, and 115 unique decision/question references present in
the tree. REGISTRATION-003 is AGREED and Q-042 remains OPEN. No application
source or implemented authorization behavior changed.

## Checkpoint following b516ff8 — qualified Q-040 approval

REGISTRATION-002 records Q-040's approval with support-relationship declarations
as an optional registration feature. Registration binds those relationships;
Auth stays domain-agnostic. Q-041 proposes omission behavior and remains open.
Neither mandatory relationship metadata nor "omitted means all combinations
supported" is adopted. Individual registration, canonical checks, issuance
authority, and runtime scope requirements are not made optional.

The registration chapter, log, and tree preserve the earlier proposal and
explain the user's qualification. No schema or implementation is adopted.

Checks for this docs-only checkpoint passed: `git diff --check`, 19 Markdown
files, 96 local links, 42 JSON blocks, and 113 unique decision/question references
present in the tree. The new Q-041 remains OPEN. No application source changed.

## Checkpoint following 8e1c16f — Q-039 approval

REGISTRATION-001 records the approved refinement of Q-039: applications register
supported permissions and scope contracts; Auth validates grant acceptance
against them and canonical authority rules without interpreting application
business meaning. [Application registration](application-registration.md)
contains the detailed rationale, example, counterexamples, and remaining gaps.

Earlier Q-039-open statements below describe the preceding checkpoint. They are
preserved, not the current status. Registration format, authorized registration
management, versioning, reference checks, and compatibility declarations remain
open; Q-040 is the next proposal. No new schema or implementation is adopted.

Verification for this checkpoint: `git diff --check` and `npm run build` passed.
The documentation scan checked 19 Markdown files, 96 local links, 42 JSON
blocks, and 111 unique decision/question references present in the tree.
The build remains a packaging check, not proof of authorization implementation.

## Checkpoint following d581fba

User requested reconciliation, commit/push, and preservation by deprecation or
strikethrough rather than deletion. They also requested worked use cases across
Git/ticketing/HRMS/accounting. This checkpoint is documentation reconciliation,
not handbook-v1 completion, a new policy decision, or an authorization-engine
migration. The [decision log](handbook-roadmap.md) records agreement status.

The discussion subsequently added explicit approvals Q-036 through Q-038 to
this checkpoint: the two responsibility layers, embedded-agent integration,
and application-owned scope meanings with endpoint fact bindings. Q-039 remains
an open proposal about sharing definition contracts; recording it is not approval.

## Current reference map

| Topic | Current agreed basis | Current explanation |
|---|---|---|
| Scope meaning and ownership | SCOPE-003/006 | [Scope boundaries](scope-model.md): boundary selector; scope owns its meaning, grant binds its use. |
| Application meanings and scope compatibility | Q-038; conceptual core of SCOPE-004 | Applications define supported boundary relationships; endpoints bind trusted facts rather than infer meanings from matching names. Grant-validation mechanics remain open. |
| Scope format and combination | SCOPE-007/008 | Required flat string-valued object; AND within scope, alternatives through complete grants; explicit {} tenant-wide; missing/null invalid. |
| Grant binding, roles, dependencies | GRANT-001/002/003, ROLE-001/002, RESOLUTION-003/004 | [Current grant formats](grant-format.md) and [grant chapter](grant-model.md). Complete lifecycle/transport schemas remain open. |
| Endpoint gate and material | CONTRACT-006/007, INPUT-001 | [Endpoint authorization](endpoint-authorization.md): one gate, explicit permission/material/source declaration, no prepared handoff. |
| Logical architecture | CONTRACT-006/007 and ENFORCEMENT-002 | [System overview](system-overview.md) and [SVG](assets/authorization-system.svg). Diagram is not a deployment or network-call requirement. |
| Two layers and embedded auth agent | ARCH-004/005; Q-036/037 | Auth supplies authority material; the application supplies domain meanings and facts. Shared canonical evaluation spans both at one endpoint-owned gate. The existing diagram remains compatible; it does not yet label the layers. |
| Authority combination and automation | DECISION-001/002, AUTHORITY-002, DELEGATION-002 | Positive alternative grants retain dependencies; service/agent authority remains human-dependent and bounded. |
| Administrative separation | ADMIN-001/002/003 | Business access and access administration are separate; self-assignment must be explicit, authorized, and audited. Exact bounds remain open. |

## Preserved history and reconciliation actions

| Earlier material | Status and handling |
|---|---|
| [Original lab handbook](../src/content/authorization-concept.md) | Historical notice added. Input, grant-layout, typed-scope, and audit-receipt sections carry targeted cautions. Original substantive text and examples remain. |
| [Original HRMS setup example](../src/content/hrms-tenant-setup.md) | Historical notice added. Scope catalogs/tenant_self and storage/bootstrap assumptions are not promoted to current rules. Personal business-access possession is not a universal administration prerequisite. |
| [Original projects/repositories example](../src/content/projects-repositories-teams.md) | Historical notice added. Old layouts and external implementation assertions remain preserved, not silently revalidated. |
| [Earlier lab design](design.md) | Historical notice and deprecated typed-vocabulary annotation added; original proposal retained. |
| [Two-mode authorization flow](authorization-flow.md) and [completion cases](endpoint-completion-cases.md) | Already deprecated; retained without rewriting their historical explanation. Cases may still motivate fact/enforcement requirements. |
| [Grant examples](grant-examples.md), EX-001 through EX-006 | Deprecated layout labels retained. [Current equivalents](grant-format.md) and EX-007 use canonical scope. |
| Roadmap resume notes | Stale mode/repeated-evaluation wording struck through or explicitly annotated, with current CONTRACT-006/007 references. Historical decision text is preserved. |
| [Handbook mind map](discussion-tree.md) | Current point and return paths remain explicit. Cross-domain scenarios are available evidence, not a claim of closing validation stage 10. |

Strikethrough/deprecation preserves what was said while making it clear that it
is not the current rule. A current chapter's old introductory "unreconciled"
wording is struck and qualified: historical prose is now classified and
annotated, not fully rewritten or migrated. The lab's earlier "complete concept
guide" description does not mean the new handbook is complete.

## Reading older decision rationale

The log records rationale at the time of each decision. An old sentence saying
"this does not settle X" is not a veto on a later explicit agreement:

| Earlier snapshot | Later resolution |
|---|---|
| GRANT-001 did not yet settle grant/assignment vocabulary. | TERM-004 settled the single-binding terminology. |
| GRANT-003 / DECISION-001 did not yet settle positive/deny combination. | DECISION-002 excludes explicit deny grants from v1. |
| SCOPE-001/003 and SELF-001 did not yet settle encoding. | SCOPE-007 settled flat scope syntax and the $self token; relationship definitions and timing remain open. |
| ADMIN-001 did not yet settle whether administrators need business access. | ADMIN-002 separates administration from personal business access. |
| Q-031 / SCOPE-008 still described key-value syntax as proposed. | Q-034 approved SCOPE-007. |
| CONTRACT-002 had settled two modes. | CONTRACT-006 later deprecated them; CONTRACT-007 specifies the single declaration's material sources. |
| ARCH-004 / Q-036 left embedded-agent integration as a follow-up. | Q-037 approves ARCH-005's integration responsibility, without fixing deployment or interface syntax. |
| Q-028 and scope chapter introductions left explicit resource relationships open. | Q-038 agrees application-defined meanings and supported resource relationships. SCOPE-004 remains only partially agreed because validation mechanisms are still open. |

This is cross-referencing already approved decisions, not retroactive approval
of proposals such as scope-key governance or administrative containment.

## Cross-domain example coverage

[Worked use cases](use-case-examples.md) adds four scenario groups per domain:

- Git hosting: project boundaries, separate read/write routes, role/tenant reach,
  and human-dependent agent restrictions.
- Ticketing: queue boundaries, requester self, AND selection, and the difference
  between request assertions and actual ticket facts.
- HRMS: employee self through groups, shared permissions/scope, direct human
  grants with AND boundaries, and membership removal preserving other routes.
- Accounting: read versus approve, missing facts, tenant-wide ledger access,
  and current role-definition changes.

Each gives grant JSON, endpoint input sources, assumptions, and expected
outcomes. Permission names and key meanings are fictional application examples,
not a new platform-wide catalog. These are documentation consistency scenarios,
not claims that the interactive engine implements them.

## Explicit remaining gaps

- ~~Scope-key meanings, ownership, and definition lifecycle: current discussion.~~
  Q-038 settles application ownership and explicit boundary meanings. Shared
  definition governance (Q-039), concrete definitions, and lifecycle remain open.
- Administrative scope containment, bootstrap, and grant lifecycle: return point.
- Exact endpoint declaration, request, result, and resolved-grant schemas.
- Group ownership/human-only membership proposal consolidation where the log
  still says OPEN/PROPOSED; do not silently mark these agreed during cleanup.
- Collections, bulk/aggregate policy, moves, approval conditions, concurrency,
  revocation freshness, and audit guarantees.
- Original code and other-repository implementation claims: no revalidation
  or migration is implied by preservation notices.
- Full scenario suite, external primary-source checks when requested, and v1
  publication: still unfinished. No completion percentage is asserted.

## Verification scope

This checkpoint checks reference/link consistency, JSON syntax and canonical
scope shapes in current examples, preservation of historical source material,
and the scenario expectations under an illustrative bounded model. Building
the lab verifies that Markdown notices still package with its existing code;
it does not verify implementation of the current authorization contract.

### Checks run for this checkpoint

- `git diff --check`: passed.
- Documentation scan: 18 Markdown files, 91 local links, 41 valid JSON blocks,
  and 109 unique decision/question references, each present in the tree.
- Example structure: 16 unique scenario groups and 12 grants with canonical
  flat string-valued scopes and one permission source per grant.
- Historical preservation: all original lines in the lab design and three
  original application-content pages remain in order; annotations are additive.
- Illustrative-model check: 39 assertions across all 16 scenario groups passed.
  This temporary check used already trusted, normalized facts and the documented
  grant fixtures; it did not test fact retrieval, caches, application handlers,
  or a production authorization engine, and adopts no evaluator implementation.
- `npm run build`: passed (TypeScript and Vite). This checks that the existing
  application still packages with the Markdown notices, not that it implements
  the new authorization model.
