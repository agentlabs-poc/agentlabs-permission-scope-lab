# Demonstrations — `application_registry`

Captured CLI output. The two binaries were run against real SQLite stores, the
session captured, and each SVG generated from that capture line by line — nothing
here was typed by hand, layout included.

| File | Shows |
|---|---|
| `demo-1-registry-contract.svg` | The domain alone: registration add-only and shape-checked, install requiring the application, listing by tenant and by slug and refusing to answer unfiltered, uninstall reporting whether it did anything, suspend reversible — and the two record types sharing `application_registry_l1_records`, told apart by `key2` |
| `demo-2-composition-root.svg` | The two domains composed: the gate shut, then opened by writes made in the *other* domain's store, then shut again by suspend and by disable — and both stores' tables listed to show there is no shared one |
| `composition-root.svg` | Not a capture — the block diagram of the seam: Auth-AL declares the interface, the registry satisfies it structurally, and only `implementation/wiring` imports both |

The second demonstration earned its place twice. Its first version matched on an
error string and reported Auth-AL's own missing record as the registry refusing —
a demo that lied. Rewritten to print the port's answer and Auth-AL's separately,
it then found a real defect: the cutover had left Auth-AL still requiring a row in
its own `applications` table, so the gate opened and the read failed anyway.

Regenerate after any contract change. A demonstration that describes behaviour the
code no longer has is worse than none.
