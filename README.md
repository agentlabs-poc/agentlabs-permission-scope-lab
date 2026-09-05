# Authorization Explanation Bench

Read the [working Authorization Handbook](docs/handbook.md) for the chapters
being developed in the current discussion. The original lab concept page has
not yet been reconciled with these decisions.

The [agreed handbook roadmap and decision log](docs/handbook-roadmap.md) tracks the
work to finalize the shared authorization foundation. Start at its "Resume here"
section for the next discussion. The [discussion tree](docs/discussion-tree.md)
maps every branch, its conclusions, and the remaining work.
The [working grant chapter](docs/grant-model.md) develops the agreed concepts
with rationale, examples, and explicitly open details.

A local explanation bench for exploring scoped authorization through documented examples and an interactive request evaluator.

The lab separates:

1. **Capability** — the stable `<namespaced-noun>::<verb>` permission.
2. **Reach** — the scope attached when that permission or role is assigned.
3. **Target** — the trusted resource and ownership context resolved for a request.

An authorization decision is the intersection of all three. The simulation is deliberately in-memory and is not an authorization implementation.

The bench contains a complete concept guide, worked HRMS and projects/repositories examples, an Auth/Authz responsibility map, and an exportable design-question board. Its HRMS and Projects/Repositories guided explorers start with familiar stories and walk each request through five explained stages. See [the working design](docs/design.md).

```bash
npm install
npm run dev
```
