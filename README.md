# Authorization Explanation Bench

Read the [working Authorization Handbook](docs/handbook.md) or open the local
reader at `/` or `/concept.html?doc=concept`. The approved discussion—not the historical
lab prose or simulations—is the source of truth.

The current model uses permissions, flat boundary scopes, complete grants,
human-dependent authority, and one endpoint-owned gate. The endpoint declares
one permission and selected input sources, then keeps actual execution within
the authorized boundaries. There is no prepared handoff or canonical relationship
block in endpoint policy.

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
behavioral contracts are unfinished. The user has explicitly reopened the
commit/push gate for the verified checkpoint and continued handbook work.
Earlier freeze notices in historical snapshots describe the review period.
