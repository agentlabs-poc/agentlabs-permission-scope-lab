# Authorization implementation

Software delivery work lives here, separate from the
[Handbook of Authorization](../handbook/README.md).

All implementation plans live in `plan/`:

- [ABV architecture and rationale](plan/abv-design.md)
- [ABV, reusable CLI and SQLite implementation plan](plan/abv-implementation-plan.md)

Source: [abv/](abv/README.md), an isolated Go module. CP1 is implemented and
independently reviewed; see [checkpoint progress](plan/progress.md). Tenant/application is the
mandatory outer boundary. The planned working CLI calls ABV in process, with
SQLite behind a replaceable storage provider. Persistence and mutation commands
are not supplied by the initial CP1 command shell.
