# Q-101 Binding Documentation Plan

**Goal:** Pin the approved Q-101 discussion in the lab with diagrams, case coverage,
rationale, and reconciled entry points. This is documentation, not an Auth runtime.

**Source:** User-approved discussion and sibling scratchpad
`auth-scratch pad/q101-parent-grant-rules.md`, through Q-101E-3.

**Architecture:** One current chapter owns the binding and lifecycle explanation.
Three SVGs show the four-part binding, direct/shared routes, and structural-change
sequence. Existing chapters retain their earlier prose with explicit supersession
notices. Existing reader asset discovery publishes linked Markdown and SVGs.

**Constraints:** No commit/push; preserve tag 0.0.1 and scratch originals. No runtime
changes, new hierarchy/revision schemas, automatic rebinding, or unrelated decisions.
The latest correction permits disabled bindings, not only removed assignments.

## Tasks

- [x] Add `docs/parent-grant-bindings.md`: approved rules, versioned excerpts,
  four-part relationship, lifecycle distinctions, rationale, numbered case matrix,
  superseded alternatives, and precise remaining contracts.
- [x] Add accessible SVGs under `docs/assets/`: `parent-grant-bindings.svg`,
  `parent-grant-routes.svg`, and `binding-change-lifecycle.svg`.
- [x] Reconcile AGENTS gate, README, handbook/tree/roadmap/reconciliation, affected
  authority/assignment/subgroup/lifecycle chapters, and reader concept content.
  Preserve earlier text rather than deleting the discussion trail.
- [x] Verify JSON and local links, SVG XML and browser rendering, reader tests,
  production build, whitespace, diff scope, and unchanged baseline tag. Review
  for contradictions between grant disablement, assignment disablement, runtime
  ineffectiveness, and structural parent changes.

No completion percentage is recalculated. A case matrix is design coverage, not
proof that a production evaluator implements these rules.
