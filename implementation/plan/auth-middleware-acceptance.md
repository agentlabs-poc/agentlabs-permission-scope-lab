# AUTH-MW-03 — first-slice acceptance obligations

This matrix fixes observable behavior from the existing handbook before SDK code.
It does not invent Auth transport or mark unimplemented model paths supported.
Authority supplied by a deterministic test adapter is explicitly fixture evidence.
AUTH-MW-04 now permits a read-only adapter over a disposable existing-schema ABV
SQLite database. It must prove real stored assignment/membership/lineage/control
behavior, no writes from evaluation, context isolation and consistent loading.
An invalid/unavailable database is evaluation failure, not an empty authority set.

| ID | Scenario | Expected observable behavior |
|---|---|---|
| MW01 | Missing/invalid server policy | Protected handler never runs. |
| MW02 | Caller submits alternate permission/policy | Cannot alter the endpoint's fixed permission or input bindings. |
| MW03 | Required selected body field absent, same name present in query | No source fallback; handler never runs. |
| MW04 | Missing/invalid/unsupported identity or wrong requested tenant/app | No protected effect; never trust a body identity or infer scope from a permission prefix. |
| MW05 | Complete eligible FIN-read route and FIN/C17 request | Allow with supporting grant references; handler remains constrained to evaluated request. |
| MW06 | FIN-read grant plus ENG-write grant, FIN-write request | Deny; no cross-route permission/scope mixing. |
| MW07 | First route inapplicable, another complete route authorizes | Allow on a valid complete route; no first-failure shortcut. |
| MW08 | Parent FIN plus child C17 | Both restrictions survive, including same-key conflicts; child {} does not erase FIN. |
| MW09 | Valid group-derived self scope | Self resolves to authorizing human, not group or caller-selected owner. |
| MW10 | Assignment pins revision1 while newer content exists | Runtime uses actual adopted lineage, not newest published content. Evidence producer and consumer obligations must be explicit. |
| MW11 | Required parent/control/membership support unavailable | That route cannot authorize; another complete valid route may still work. No orphan authority. |
| MW12 | Complete valid empty authority set | Completed deny; malformed/incomplete set is not treated as empty. |
| MW13 | Authority lookup timeout/incomplete evidence | Evaluation error, no decision field in agreed error variant; no handler effects. |
| MW14 | Invalid/mixed/truncated evaluator result | Stop; neither guessed deny nor allow. |
| MW15 | Evaluator deny/error | Both approved message fields and code preserved for response adapter; no fabricated private-only reason semantics. |
| MW16 | Query says FIN/C17, stored C17 is ENG | No ENG data disclosed; handler must not retry by certificate ID alone. |
| MW17 | Submitted PUT department is FIN, current row is ENG | Submitted value is not proof of current authority; unsupported move composition fails closed. |
| MW18 | Explicit FIN collection vs all-department request | FIN request may allow; broad request denied with FIN-only authority, not silently filtered. |
| MW19 | Authority reduction between two checks | Second request cannot reuse withdrawn support; fixture proof is not a production freshness protocol. |
| MW20 | Same allowed synchronous request, unrelated subsequent authority reduction | Preserve approved Q-129 semantics, not a blanket cancellation guarantee. Actual data boundaries still bind at use. |
| MW21 | Clock crosses a supplied validity boundary during evaluation | Expired authority cannot become an allow due to stale local time comparison; exact evidence/time responsibilities fixed in M0. |
| MW22 | Unsupported proxy/direct-recipient evidence path or unknown constraint | Explicit failure, not silently dropped restriction or unrestricted human access. |
| MW23 | Cancellation / bounded work limit | Stop; no partial authorization or protected effects. |
| MW24 | API-side binary dependencies | No Auth SQL/store or ABV mutation coordinator embedded; shared canonical types alone do not violate placement. |

## Integration test discipline

Use known fixture route sets and a small in-memory application data store. Record
exact handler call/effect counts, not just status codes. Prove identity/material
cannot be changed between checked input and protected effect through adapter misuse
covered by the contract. Constrained endpoint code remains an application trust
responsibility; a wrapper is not proof against arbitrary malicious handler code.

Test configured authority providers as trusted service adapters, never treat JSON
submitted by the requesting client as grant evidence. No full production Auth
transport, network freshness proof, JWT verifier or middleware HTTP status mapping
is claimed by passing these local tests.

Use seeded SQLite authority for the integrated AUTH-MW-04 walkthrough, retaining
small in-memory provider fixtures for pure failure/clock tests. Apply an authorized
control change through the existing ABV test harness between requests; verify the
next adapter read/evaluation reflects the change. Do not create arbitrary SQL
writes as a new public authority-management operation for the sake of the demo.

Each implementation brief selects its owned cases and exact test command. Full
first-slice handoff reports covered, deliberately unsupported and unresolved cases
separately, with rationale. No percentage based on merely listing test IDs.
