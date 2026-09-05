# Handbook reconciliation implementation plan

> Execute locally in the user's current workspace. The explicit commit/push
> freeze overrides earlier checkpoint-commit instructions. No runtime
> authorization migration or new canonical decision is included.

**Goal:** Make the reader, logical diagram, and current handbook navigation
reflect the approved discussion through Q-050-F.

**Architecture:** One endpoint-owned gate combines canonical authority with
application meanings and mandatory enforcement. There is no prepared handoff,
canonical relationship block, or additional canonical resource wrapper.

**Tech stack:** Markdown, SVG, existing TypeScript/Marked/Vite reader.

**Spec:** [Endpoint policy](../../endpoint-policy-format.md),
[decision log](../../handbook-roadmap.md), [scope](../../scope-model.md).

## Global constraints

- Preserve old documents and SVG before updating their active versions.
- Approved discussion is authoritative; Q-050's remaining publication details,
  decision results, update/move rules, and freshness remain open.
- Exactly one permission per endpoint; versioned policy with explicit inputs.
- Required input presence and application-owned value validation stay distinct.
- No commit, push, new deployment, or interactive evaluator rewrite.

## Local execution checklist

- [x] Archive original reader documents, overview, entry points, and diagram
  under `docs/history/reconciliation-2026-09-05/`; compare bytes with HEAD.
- [x] Reconcile `src/content/*.md`: explain flat AND scopes, complete grants,
  human membership/dependency, one endpoint gate, and bounded execution with
  versioned grant/GET/PUT examples. Link rationale and preserved originals.
- [x] Update `docs/system-overview.md` and `docs/assets/authorization-system.svg`:
  show Auth and application responsibilities, required policy/input validation,
  complete authority evaluation, mandatory constrained execution, and failure
  paths. Do not require a relationship resolver or eager lookup.
- [x] Update `src/concept.ts` descriptions and render the shared SVG URL; add
  responsive image styling. Label `src/main.ts`, `src/projects-explorer.ts`,
  and `enforcement-trace.html` as historical without changing their algorithms.
- [x] Reconcile `README.md`, `docs/handbook.md`, current discussion-tree status,
  and reconciliation register. Preserve detailed decision/rationale history.
- [x] Verify JSON examples, local links, archive byte equality, SVG rendering,
  and `npm run build`; run `git diff --check` and report actual local changes.
  Confirm HEAD unchanged and leave all changes uncommitted.
