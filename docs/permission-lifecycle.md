# Registered permission lifecycle — Q-125

**Q-128 follow-up:** [all confirmed authority reductions have no stale-cache grace for new checks](authority-freshness.md).
Effective retirement cannot be bypassed by an older catalog/root representation.
Exact confirmation/propagation mechanics remain open; grant records are unchanged.

## Q-126 — Stable authorization meaning (approved)

**APPROVED.** An existing permission identifier must not be repurposed for a
materially different authorization meaning, including after retirement. A
different operation requires a new identifier. Descriptions and display labels
may be corrected without changing the operation's authorization meaning.

For example, `hrms:payroll:run::approve` means approving a payroll run. Changing
that same identifier to also authorize transferring money to employees would
silently expand existing authority. Payment execution needs a separate permission,
such as `hrms:payroll:run::pay`. These names illustrate the approved grammar;
they do not register new permissions or assign them to any real recipient.

**Rationale:** immutable grant content and adopted role revisions retain
permission identifiers. They would not protect against silent privilege changes
if the identifier could later stand for a different operation. Stable meaning
preserves what the administrator selected and what the recipient actually received.
Retirement does not make the name available for an unrelated operation.

**Core-philosophy check:** this preserves explicit authority selection, role/grant
revision intent, no implicit aliases, and application ownership of domain meaning.
Auth can preserve identifier history and validate registration contracts, but
does not infer business semantics or prove that new application code implements
the declared operation. The application platform administrator remains responsible
for the definition's meaning; application integration must enforce that meaning.

**Trade-off:** a materially different operation requires explicit naming and
grant/role selection instead of silently inheriting old grants. Cosmetic wording
corrections do not require inventing another permission identifier. The rule is
not a ban on bug fixes or internal implementation changes preserving the same
authorization meaning.

No new permission field, semantic-comparison engine, automatic grant migration,
or alias is introduced. This settles reuse for a *different* meaning; it does
not silently approve same-meaning restoration of a retired permission or settle
scope-definition evolution and exact lifecycle API contracts.

The earlier identifier-reuse gap below is narrowed by Q-126; remaining restoration
and lifecycle questions are distinct from repurposing an identifier.

## Q-125 — Retirement with existing grant references (approved)

**APPROVED.** The user agreed that an authorized application permission retirement
may proceed even when existing grants still reference that permission. Editing
every reference first is not a prerequisite for retirement.

Once retirement takes effect, the application's computed root no longer supplies
the permission. An existing grant reference cannot authorize that retired
operation. Auth does not silently rewrite or delete stored grants or assignments.

Retirement requires the application-management authority described in Q-121;
it is not authorized merely by holding the business permission or by administering
an unrelated tenant grant. Q-123's one shared application catalog applies: this
is not selective permission retirement by tenant application version.

## Grant example and consequence

Before retirement, assume this revision has an otherwise valid assignment and
supporting root route, and all permission/scope definitions are registered:

```json
{
  "version": "1",
  "grant_id": "G-PAYSLIP-DELETE",
  "revision": 1,
  "parent_grant_id": "G-HRMS-ROOT",
  "permissions": ["hrms:payroll:payslip::delete"],
  "scope": {"dept": "FIN"}
}
```

This is an immutable grant-revision example using existing fields, not a
retirement API or the unfinalized computed-root encoding. Tenant is implied.

```text
Before retirement
HRMS catalog supplies delete → root supplies delete → child can supply FIN delete
                                                    subject to all other checks

After effective retirement
HRMS catalog no longer supplies delete → root no longer supplies delete
                                      → stored child cannot authorize delete
```

The stored revision and assignment are retained unchanged. Their mere existence
or enabled status cannot restore the retired permission. An operation requiring
it must not be allowed. A failure to fetch retirement/catalog evidence is still
an evaluation failure, not proof that retirement occurred or permission to allow.

## Rationale, trade-off, and core-philosophy check

Old grant references should not force an application to keep supporting a
capability. Requiring a full reference migration first would make retirement
depend on every existing assignment and grant owner completing a cleanup task.
The approved alternative allows retirement and withdraws the authority it supplies.

This follows the live parent ceiling: stored child content cannot preserve
permission that the valid source no longer supplies. Retaining the records
preserves immutable content and explicit assignment lifecycle; it does not
grandfather access. The trade-off is loss of access for operations relying on
that permission when retirement becomes effective.

No automatic grant-disable mutation, assignment deletion, new status value,
`revoke` operation, or migration permission is introduced. The existing Q-042
rule for activating permission/scope compatibility validation is unchanged;
retirement does not authorize silently enabling an incompatible configuration.

## Remaining contract boundaries

The retirement request/evidence format, effective visibility and concurrency,
permission-identifier reuse/restoration, and scope-definition evolution remain
open. This decision specifies loss of the retired permission; it does not settle
every consequence for other still-supported permissions in a mixed grant.
Any such outcome must respect the agreed complete-lineage rules, not silently
rewrite the stored permission selection.

Sources: [permission meaning](permission-model.md),
[application registration](application-registration.md),
[publishing authority](application-platform-authority.md),
[computed root and shared catalog](root-permission-evolution.md), and
[grant lifecycle](grant-lifecycle.md).
