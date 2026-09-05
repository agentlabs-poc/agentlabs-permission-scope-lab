# Authorization Explanation Bench

Read the [working Authorization Handbook](docs/handbook.md) for the chapters
being developed in the current discussion. ~~The original lab concept page has
not yet been reconciled with these decisions.~~
Reconciliation checkpoint: the original lab pages now carry historical-status
and deprecation notices. The linked working chapters hold the current decisions;
the original prose and interactive implementation have not been rewritten.

The [agreed handbook roadmap and decision log](docs/handbook-roadmap.md) tracks the
work to finalize the shared authorization foundation. Start at its "Resume here"
section for the next discussion. The [discussion tree](docs/discussion-tree.md)
maps every branch, its conclusions, and the remaining work.
The [working grant chapter](docs/grant-model.md) develops the agreed concepts
with rationale, examples, and explicitly open details.
The [current grant formats](docs/grant-format.md) use canonical v1 scope;
earlier layouts remain available as explicitly deprecated examples.
The [cross-domain use cases](docs/use-case-examples.md) exercise Git hosting,
ticketing, HRMS, and accounting. The [reconciliation register](docs/reconciliation.md)
separates current decisions, preserved history, and implementation gaps.

A local explanation bench for exploring scoped authorization through documented examples and an interactive request evaluator.

The lab separates:

1. **Capability** — the stable `<namespaced-noun>::<verb>` permission.
2. **Reach** — the scope attached when that permission or role is assigned.
3. **Target** — the trusted resource and ownership context resolved for a request.

An authorization decision is the intersection of all three. The simulation is deliberately in-memory and is not an authorization implementation.

> Historical feature description below: "complete concept guide" describes the
> earlier lab artifact, not completion of the current Authorization Handbook v1.

The bench contains a complete concept guide, worked HRMS and projects/repositories examples, an Auth/Authz responsibility map, and an exportable design-question board. Its HRMS and Projects/Repositories guided explorers start with familiar stories and walk each request through five explained stages. See [the working design](docs/design.md).

```bash
npm install
npm run dev
```
