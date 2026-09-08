# Authorization implementation

Software delivery work lives here, separate from the
[Handbook of Authorization](../handbook/README.md).

All implementation plans live in `plan/`:

- [ABV architecture and rationale](plan/abv-design.md)
- [ABV, reusable CLI and SQLite implementation plan](plan/abv-implementation-plan.md)

Source: [abv/](abv/README.md), an isolated Go module. CP1 foundations and CP2
SQLite persistence are implemented and independently reviewed; see
[checkpoint progress](plan/progress.md). Tenant/application is the
mandatory outer boundary. The planned working CLI calls ABV in process, with
SQLite behind a replaceable storage provider. SQLite persistence is available
internally; actual lineage, both authority gates and working protected CLI
commands arrive together in CP3. The command shell does not expose raw storage.
