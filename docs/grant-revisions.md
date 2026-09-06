# Shared-grant revisions, adoption, and current lineage — Q-102–Q-106

**Q-122 approved refinement:** [computed root permission coverage](root-permission-evolution.md)
is selected for legitimate roots; catalog additions do not require automatic
root revision publication/adoption. This supersedes the mechanism-open wording
below, not ordinary explicit grant/role adoption or immutable-content rules.
Root encoding and applicable catalog/version binding remain open.

**Root-evolution qualification — Q-120A:** [automatic root growth](root-permission-evolution.md)
is now intended for legitimate application capability upgrades, without a
separate manual root-update/adoption gate. Live versus materialized revision
mechanics remain open. Do not infer in-place mutation of published content or
automatic selection/adoption for ordinary dependent grants. The earlier blanket
root-growth assumption is qualified; original decisions/rationale remain below.

**AGREED at the rule level described below.** The user approved these decisions
and requested recording in the lab by default. Earlier scratch notes are retained
as discussion history. This chapter is current for the approved revision rules;
the [Q-107 core JSON](grant-revision-format.md) was subsequently **approved**.

These rules extend the existing [role-revision principle](role-revisions.md)
to reusable grant authority while preserving [Q-101's dependent bindings](parent-grant-bindings.md).
Tenant is implied. No runtime migration, new owner powers, or complete schema
is adopted by this recording.

## Q-102 — Each assignment adopts its revision explicitly

**Agreed in principle.** Different recipients of the same reusable grant may
adopt different published revisions. Publishing a revision does not automatically
change an existing assignment. Adoption is an explicit authorized operation
with complete current boundary/dependency validation.

```text
G1 revision 1: FIN read
G1 revision 2: FIN read/write

Assignment A1 → Team1: explicitly adopts revision 2
Assignment AX → TeamX: remains on revision 1
```

The different versions may coexist because the assignments adopted them at
different times. Q-104A/Q-105 below restrict new creation and explicit upgrades
to the latest published revision; they do not auto-upgrade old assignments.

**Rationale:** authoring a reusable definition and changing a recipient's authority
are separate actions. Independent adoption avoids silently changing everyone
attached to a shared grant. The accepted cost is revision-selection state and
validation across assignments that can hold different revisions.

**Core-philosophy check:** no silent expansion; separate administrative and
source-boundary checks; live parent ceilings remain mandatory. Retaining an
older revision cannot preserve authority its required current support no longer
permits. Revision selection cannot bypass grant-wide or assignment disablement.

## Q-102A — Maintain advisory update candidates

**Agreed at advisory-rule level.** Make assignments with available newer grant
revisions discoverable to authorized administrators, with ongoing suggestions.
The view can show the assignment/recipient, adopted revision, available revision,
and differences in permissions or scope that need review.

| Assignment | Adopted | Available | Change to review |
|---|---|---|---|
| G1 → TeamX | Revision 1 | Revision 2 | Adds FIN write |

An older revision is not automatically invalid. A newer revision is not
automatically suitable or safer: it may broaden, narrow, or otherwise change
authority. The suggestion does not prove that adoption will pass current checks.
Auth can expose the authorized information; an administration layer presents
suggestions. No notification channel, delivery schedule, mandatory UI design,
or new request-time authorization dependency is prescribed.

**Rationale / philosophy:** visibility helps complete deliberate adoption work
without turning a reminder into authority. No automatic adoption or implicit
enablement occurs. Current validity must be checked when the action is taken,
not inferred from an earlier suggestion list.

## Q-103 — Resolve current adopted lineage from top to bottom

**Agreed.** “Lineage latest” means the revision **currently adopted in the
actual required supporting lineage**, not the newest revision merely published
or held by an unrelated team. For subteam authority, use the revision adopted
by the required supporting parent-team assignment.

```text
Valid upstream authority
           ↓
G1 effective authority at its required supporting assignment
           ↓
G2: permissions subset; parent scope AND additional scope
           ↓
G3: constrained by G2's effective authority
```

This is logical evaluation order, not a prescribed network-fetch sequence.
It differs from bottom-up assignment disable/removal for structural changes.

Example: Team1 holds G1 revision 1 (FIN read), while TeamX holds revision 2
(FIN read/write). TeamY under Team1 receives derived G2. Its parent ceiling
comes from Team1's revision 1, not TeamX's revision 2. If Team1 explicitly and
validly adopts revision 2, that becomes the current parent ceiling for this
route. A read-only G2 does not gain an unselected write permission.

Parent adoption can change inherited scope even when a child keeps additional
scope `{}`. Therefore validate the complete resulting authority and affected
bindings; unchanged child JSON is not proof of unchanged effective reach.
Ineffective or orphaned required support cannot produce effective descendants.

**Rationale / philosophy:** establish the parent's actual authority before
deriving the child's. This preserves live dependency, non-expansion, explicit
adoption, and the team ceiling without an independent child selection that can
substitute a broader parent revision. Own adopted content remains subject to
current support, not an independent frozen entitlement.

Q-104 resolves duplicate G1 holdings at the same recipient. [Q-112A](direct-human-parent-context.md)
reaffirms lineage-supported latest for the direct-human discussion as well;
the independent parent-revision-field proposal is withdrawn. Eligible support
discovery and validation/evidence contracts remain to be completed for these
routes; that work does not reopen the agreed governing selection principle.

## Q-104 — One current assignment per grant and recipient

**Agreed.** Within the trusted tenant, a given grant identity has at most one
current assignment to a given recipient type/identity. Retained disabled
assignments count; creating a duplicate is not a way to bypass disablement or
explicit adoption. Historical/deleted records are not current assignments.

| Situation | Result |
|---|---|
| A1 assigns G1 revision 1 to Team1 | One current assignment. |
| Create A9 assigning G1 revision 2 to the same Team1 | Reject duplicate; use the authorized adoption operation on A1. |
| Assign G1 to different TeamX | Permitted subject to current creation and authority checks. |
| A1 is disabled; create another G1-to-Team1 assignment | Still a duplicate. |

**Rationale / philosophy:** one adopted revision and one assignment state per
grant/recipient pair make parent-team support lookup deterministic. Reuse across
different recipients survives. The deliberate restriction excludes simultaneous
duplicate assignments of the same grant to the same recipient. This is not a
database-index prescription or permission to recreate deleted support without
Q-101's dependency and explicit-enablement checks.

## Q-104A / Q-105 — Latest on creation and explicit upgrade

**Agreed.** A new assignment must select the latest published grant revision.
An explicit upgrade of an existing assignment must also select the latest,
not an intermediate revision. Existing assignments can remain unchanged.

| Operation | Revision behavior |
|---|---|
| Create an assignment | Must select latest published revision and pass current checks. |
| Publish a newer revision | Existing assignments keep their adopted revisions; suggest review. |
| Explicitly upgrade an assignment | Must select latest published revision and pass current checks. |
| Re-enable a disabled assignment | Not an upgrade; retain adopted revision and validate current reality. |

Example: G1 revision 1 is read, revision 2 adds write, and latest revision 3
also adds delete. Team1 currently uses revision 1. It may stay there, but if
it upgrades it must select revision 3. If its permitted authority cannot support
revision 3, reject the upgrade and leave the assignment unchanged, subject to
its normal current validity. Do not silently select revision 2 instead. New
assignment creation likewise cannot fall back to an older revision to pass checks.

**Rationale / philosophy:** a simpler, consistent selection rule at explicit
creation/upgrade time; no silent changes to existing recipients. Latest is not
proof of entitlement. The accepted trade-off is loss of intermediate-version
flexibility even where an intermediate revision would fit current bounds.
No automatic parent upgrade, disabled-record enablement, or rollback operation
is introduced. [Q-110](auth-write-consistency.md) subsequently agrees that a
conflicting change before the protected write stops the attempt, without stale
revision selection or silent substitution. Exact persistence mechanics remain open.

## Q-106 — Immutable revision content, live administrative controls

**Agreed.** Published grant authority content is immutable. Changing permission
selection (or the adopted role reference/revision), additional scope, or parent
grant reference requires a new revision, not editing an already adopted one.
Q-109 subsequently places optional grant validity in immutable revision content;
condition placement remains to be specified. See [grant validity](grant-validity.md)
for the approved excerpt, adoption consequences, and expiry-shortening trade-off.

Grant-wide enable/disable remains a separate live control over the grant identity
and all its revisions. Assignment enable/disable is separate per recipient route.
Neither toggle publishes a new authority revision or changes an adopted revision.

```text
G1 revision 1: FIN read             — unchanged published content
G1 revision 2: FIN read/write       — explicit adoption required

Disable G1 → no assignment of either revision supplies G1 authority
```

**Rationale / philosophy:** revision selection cannot protect recipients if
adopted content can be edited beneath them. At the same time, suspension must
remain effective without asking every recipient to adopt a new disabled version.
Immutable content is not immutable effective access: current parent authority,
scope accumulation, memberships, and lifecycle controls still constrain it.
Publishing a new parent reference cannot bypass four-part structural guards
when that change becomes effective for an assignment.

## Approval trail and remaining work

| Reference | User decision |
|---|---|
| Q-102 | “yes” to independent adoption, then proposed upgrade suggestions. |
| Q-102A | “proceed” after the advisory recommendation and philosophy check. |
| Q-103A/B | Requested lineage-latest, top-to-bottom; confirmed current adopted, not merely published. |
| Q-104/A | Agreed uniqueness and required latest revision for new assignments. |
| Q-105 | Explicitly agreed with latest-only upgrades and the stated trade-off. |
| Q-106 | “yes” to immutable content with separate live controls. |

The [Q-107 core JSON](grant-revision-format.md) was separately **approved** after
review of all three blocks; the earlier unreviewed-draft status is history.
Q-108 defers assignment validity in v1; Q-109 approves optional validity in grant
revision content, with explicit adoption for changes. Remaining work includes
complete schemas and API validation, eligible-support discovery and evidence,
timestamp/clock contracts, concurrent publication
and adoption, revision lifecycle/retention, and implementation conformance.
No new completion percentage or closure of those entire bundles is claimed.
