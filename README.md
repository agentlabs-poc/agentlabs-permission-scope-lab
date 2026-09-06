# Authorization Explanation Bench

**Current revision rules:** [Q-102–Q-106](docs/grant-revisions.md) are recorded
with rationale and examples. [Q-107 JSON](docs/grant-revision-format.md) is approved
at core-shape level; full schemas remain open. All recording and exploration now
belong in the lab. [Scratchpad sources](docs/history/scratchpad-import/README.md)
are preserved as history; no new commit/push is implied.

[Q-112A](docs/direct-human-parent-context.md) reaffirms lineage-supported latest.
The independent parent-revision-field proposal is withdrawn, not canonical.

**Latest pinned discussion: [Q-101 — parent-grant bindings](docs/parent-grant-bindings.md).**
Includes the four-part binding, direct/shared route comparisons, bottom-up
structural changes, current-state enablement, rationale, and 31 review cases.
The local documentation update is authorized; no new commit/push is implied.

Read the [working Authorization Handbook](docs/handbook.md) or open the local
reader at `/` or `/concept.html?doc=concept`. The approved discussion—not the historical
lab prose or simulations—is the source of truth.

The current model uses permissions, flat boundary scopes, complete grants,
human-dependent authority, and one endpoint-owned gate. The endpoint declares
one permission and selected input sources, then keeps actual execution within
the authorized boundaries. There is no prepared handoff or canonical relationship
block in endpoint policy.

Post-`0.0.1`, [grants and assignments](docs/grant-assignments.md) are separate:
grant definitions have no recipient; assignments bind them to recipients.
[Subteams/subgroups](docs/subgroups.md) use explicit dependent authority with
permission subsets and scope AND, not inherited membership. The tag preserves
the previous committed model; full new contracts and migration remain open.

## Review locally

```bash
npm install
npm run dev
```

- [Concept reader](src/content/authorization-concept.md): foundation and rationale.
- [HRMS](src/content/hrms-tenant-setup.md): self, Finance, GET/PUT input examples.
- [Projects](src/content/projects-repositories-teams.md): registered project boundaries.
- [System overview and shared SVG](docs/system-overview.md): current architecture.
- [Discussion tree](docs/discussion-tree.md): whole-handbook progress and open work.
- [Decision log](docs/handbook-roadmap.md): numbered agreements and preserved history.
- [Reconciliation register](docs/reconciliation.md): source status and verification.

The homepage is now the handbook. All guided explorers and the standalone
enforcement trace have been removed from the active site. Recoverable source is
preserved for the [HRMS explorer](docs/history/retired-hrms-explorer-2026-09-05/README.md)
and [previously retired pages](docs/history/retired-pages-2026-09-05/README.md).
The HRMS and Projects handbook chapters remain available. Historical simulations are not
implementations or conformance tests of the current handbook. The
[original README and documents](docs/history/reconciliation-2026-09-05/README.md)
are preserved.

## Verification and current gate

```bash
node --test tests/*.test.mjs
npm run build
git diff --check
```

The handbook remains a working edition; full schemas and several high-impact
behavioral contracts are unfinished. Q-090/Q-091 recording and local reconciliation
are authorized; this update does not itself authorize a new commit/push.
Earlier gate statements in historical snapshots describe their own checkpoints.
