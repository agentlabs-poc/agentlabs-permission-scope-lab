# Permission Quest

A local concept lab for exploring scoped authorization with payroll examples.

The lab separates:

1. **Capability** — the stable `<namespaced-noun>::<verb>` permission.
2. **Reach** — the scope attached when that permission or role is assigned.
3. **Target** — the trusted resource and ownership context resolved for a request.

An authorization decision is the intersection of all three. The simulation is deliberately in-memory and is not an authorization implementation.

The payroll content is an interchangeable scenario pack. The game also contains adversarial levels, an Auth/Authz responsibility map, a live decision trace, and an exportable design-question board. See [the working design](docs/design.md).

```bash
npm install
npm run dev
```
