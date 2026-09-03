# Authorization Explanation Bench

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
