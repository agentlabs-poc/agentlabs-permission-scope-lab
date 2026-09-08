# Authorization implementation

Software delivery work lives here, separate from the
[Handbook of Authorization](../handbook/README.md).

All implementation plans live in `plan/`:

- [ABV architecture and rationale](plan/abv-design.md)
- [ABV, reusable CLI and SQLite implementation plan](plan/abv-implementation-plan.md)

Planned source: `abv/`, an isolated Go module. No runtime implementation exists
here yet. Tenant/application is the mandatory outer boundary; the first CLI
calls ABV in process, with SQLite behind a replaceable storage provider.
