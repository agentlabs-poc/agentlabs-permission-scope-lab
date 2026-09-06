# Q-088 — minimal bootstrap grants and later self-assignment

Status: **AGREED.** The user answered the clarified Q-088 “yes, agreed.”
Previous status, retained as history: **USER DIRECTION RECORDED; clarified rule
awaiting confirmation** until that answer. This
is a bootstrap authority discussion, not a new grant format. The user clarified:
some grants are created through bootstrap, they may remain, they should be minimal,
and the bootstrapped user can subsequently assign more grants to himself.

## Agreed shape

Bootstrap creates the minimal ordinary grants needed to establish administration.
Those grants can remain after setup and use the normal grant lifecycle. They do
not have to expire or be deleted merely because setup has finished.

The bootstrapped human may explicitly assign additional grants to himself **when
his current administrative authority permits the whole assignment**. This preserves
the agreed distinction between permission to provide access and permission to use
that access. It also preserves recipient, assignable-permission, scope, validity,
and other applicable administrative bounds; new grants cannot authorize their own
creation or enlarge those bounds without supporting current authority.

## What “bypass” meant

The earlier wording meant allowing an operation solely because someone was the
bootstrap user, without a currently applicable grant authorizing the operation.
It did **not** mean retaining bootstrap-created grants. For example, allowing
Vinay to create grants after his only administrative grant has been disabled,
just because he performed setup, would be such a special exception.

The distinction is between ordinary authority created during setup and permanent
special treatment of the setup identity. No separate bootstrap-user privilege,
automatic deletion of seed grants, or automatically expiring bootstrap account
is proposed.

## Concrete example

Suppose the seed authority permits Vinay to create Finance certificate-read grants
for eligible human recipients, including himself. These are plain-language
administrative bounds; their exact scope encoding is not finalized here.

1. After bootstrap, Vinay can administer those assignments. That seed alone does
   not give him certificate-reading access.
2. He explicitly creates a Finance certificate-read grant for himself. The
   evaluator checks that his administrative authority permits that assignment.
3. The new grant can supply Finance read authority through normal evaluation.
4. He cannot instead assign Engineering access or Finance write access using seed
   authority that permits only Finance read assignments.

If tenant-wide provisioning authority is intentionally seeded, it can permit
correspondingly broad self-assignment within its bounds. That is powerful authority
even if represented by just one grant. “Minimal” must concern the authority needed
to initialize administration, not merely a small number of grant records. It does
not silently prohibit legitimate root administration or prescribe exact seed bounds.

## Rationale and remaining work

Keeping seed grants ordinary avoids a second privilege system. Keeping the seed
minimal limits initial authority without preventing authorized administration.
Making self-assignment explicit preserves the separation between administration
and use. Checking existing authority prevents “can create a grant” from being
interpreted as “can create any grant with any reach.” These are authorization
rules; no business-rule or external audit-system design is added.

The approved administrative model already supplies the whole-assignment rule.
This clarification applies it to bootstrap rather than introducing a new
universal possession ceiling: Vinay need not already possess the business access
he is authorized to assign. Exact seed contents, bootstrap control, administrative
scope representation, recovery, and repeated-initialization behavior remain open.
HC-05-09 remains open; no completion credit is claimed.

**Q-088 clarified:** Bootstrap seeds minimal ordinary grants that may remain;
the bootstrapped human can explicitly assign himself additional grants within
his current administrative authority. Is this the agreed shape?

**Answer: agreed.** Seed minimal ordinary grants, allow them to remain under the
normal lifecycle, and permit explicit self-assignment within current administrative
bounds. The user did not approve a temporary-only seed policy, automatic removal
after setup, or unrestricted self-assignment. The rationale and example above
are part of this decision; the full bootstrap procedure remains open.

## Original question — retained with terminology clarification above

**Q-088:** Should bootstrap establish the starting administrative grants, with
all subsequent administration governed by normal authorization and no permanent
bootstrap bypass?
