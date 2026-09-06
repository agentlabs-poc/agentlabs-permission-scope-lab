# Initial bootstrap assignment — Q-115 agreed

**AGREED AT ARRANGEMENT LEVEL.** The user specified: “we should create a user,
create a adminstrator group, user can be any ligitimate user.” Trusted setup
creates the selected legitimate human user and an Auth administrators group,
explicitly assigns the initial root authority to the group, and explicitly adds
the human as a member. No particular username, email, or special root identity
is required. Being legitimate does not automatically entitle every user to
administrator membership: trusted setup selects the initial administrator.

```text
Trusted setup + implied tenant
  ├── Register permission/scope contracts → establish valid initial root grants
  ├── Create administrators group ← explicit root-grant assignment
  └── Create selected legitimate human user
         └── Explicit group membership → valid administrators-group authority
```

This shows dependencies, not a prescribed API or transaction sequence. Registration
precedes grant acceptance; the group and human must exist for their links to be
established. Root authority supplies maximum intended registered permissions and
scope `{}` within the tenant under Q-114. Definitions or group names alone grant
no access. No runtime users or groups are created by this documentation change.

Assignment excerpt using existing approved Q-107 fields (not a full setup schema):

```json
{
  "version": "1",
  "id": "A0",
  "grant_id": "G0",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "bootstrap-admins"},
  "status": "enabled"
}
```

Assume G0 is validly established, enabled and time-valid, revision 1 is latest
when assigned, and the human has explicit membership. IDs are illustrative.

**Rationale:** group-held support need not be replaced merely because a human
administrator changes (Q-099). Authority comes through explicit membership and
assignment, not a special identity. This applies the group-based-access preference;
direct assignments remain supported generally. Membership is not ownership.

**Core-philosophy check:** registration, trusted establishment, tenant isolation,
administrative/source checks, status, validity, revisions, and lineage constraints
remain mandatory. Only humans join groups; services/agents remain dependent on
humans. Neither the administrator nor the group receives an evaluator bypass.

**Trade-off / remaining contracts:** this group's powerful membership must be
protected. Trusted human selection/verification, user and membership APIs, root
evidence, partial setup, and recovery remain open. Q-116 below now settles the
outcome of repeating completed initialization. No
new user/membership JSON fields, mandatory two-person setup, last-administrator
deletion ban, or automatic subgroup hierarchy is approved here.

<details>
<summary>History — original Q-115 proposal, subsequently agreed and clarified above</summary>

**PROPOSED / NOT APPROVED.** [Q-114](bootstrap-authority.md) settles registration
before grant acceptance and maximum intended initial tenant authority. The
remaining choice here is the default recipient arrangement—not whether initial
authority is maximum or whether registration is required.

## Recommended default

Hold the initial root authority in an Auth-owned administrators group and make
the first human administrator an explicit member. This applies the existing
preference for group-based access to the concrete bootstrap arrangement. Direct
human assignment remains supported; this proposal does not prohibit it globally.

```text
Registered permission/scope contracts
                 ↓
Trusted setup establishes G0: maximum intended tenant authority
                 ↓ explicit assignment A0
        Bootstrap administrators group
                 ↓ explicit human membership
               Vinay
```

Versioned proposed seed-assignment excerpt using existing Q-107 fields:

```json
{
  "version": "1",
  "id": "A0",
  "grant_id": "G0",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "bootstrap-admins"},
  "status": "enabled"
}
```

Assume valid trusted G0 establishment, registered contracts and required
compatibility checks, revision 1 latest at assignment creation, valid grant
status/time, and explicit membership for Vinay. This is not a complete bootstrap,
root, group-creation, or membership API schema. The bootstrap trust procedure
must establish the initial group/assignment/membership dependencies; it is not
an instruction to invoke an unprotected ordinary group API before authority
exists. The example group ID has no privileged meaning to the evaluator.

## Rationale and alternatives

- **Recommended: group-held authority.** Replacing a human administrator need
  not replace the root assignment merely because its original recipient was
  that human. This follows Q-099's separation of team-held continuing support
  from the identity of the administrator who acted.
- **Alternative: direct initial assignment.** Fewer initial membership records,
  but any later change of recipient/support must be explicitly managed. It is
  still valid under the general model and is not silently migrated by this proposal.

**Core-philosophy check:** root authority still comes from registered contracts
and trusted establishment. Membership supplies the group's valid assigned
authority; it does not supply ungranted powers by virtue of ownership or a group
name. Subsequent group/membership/assignment changes still require their actual
administrative permission and source/binding checks. Only humans are members;
agents/service accounts remain human-dependent. All tenant, scope, validity,
revision/adoption, orphan, and structural constraints continue to govern actual
lineage. Group ownership is not conflated with membership.

**Trade-off and remaining risk:** admitting a human to this group gives that
human the group's powerful initial authority. Its membership administration is
therefore security-critical. Recovery after losing all administrators, protected
group lifecycle, exact membership/owner records, and initial transaction/failure
contracts remain open. This proposal does not require two initial administrators,
two-person approval, or a ban on disabling/deleting the last administrator.
It does not automatically make every other group a child of this group, or
guarantee that unrelated personally dependent routes survive membership changes.

**Q-115:** should group-held root authority with explicit initial-human membership
be the default bootstrap arrangement?

</details>

## Q-116 — Repeating completed bootstrap (agreed)

**AGREED.** The user answered “Q116 agree.” Once initial setup completes for a
tenant, repeating bootstrap reports already initialized without changing grants
or administrator membership. The original proposal is retained below as history;
its recommendation is now the agreed rule. This does not prohibit ordinary
authorized administration or later application registration.

**Accepted consequence:** bootstrap cannot reset intentionally changed root
authority, re-enable disabled grants/assignments, restore removed membership, or
add a different initial administrator. Recovery must not be achieved by replaying
completed initial setup. Exact recovery and completion-evidence contracts remain
open; this approval adds no canonical field or wire error code.

<details>
<summary>History — Q-116 proposal, approved above</summary>

**Context:** Q-113 left repeated initialization open. This question concerns
initial setup invoked again for the same tenant after it completed, not normal
later administration or registration of another application.

**Recommendation:** report already initialized without changing authority. Do
not recreate/reset root grants, re-enable disabled records, or add another human
to the administrators group through repeated initial setup.

**Example:** initial setup creates Maya's user and the administrators group and
explicitly establishes her membership. A later setup invocation naming Nutan
must not add Nutan, even if he is legitimate. Normal authorized membership
administration can explicitly add him.

**Rationale / core-philosophy check:** initial establishment must not become a
continuing alternate authority-change route. This preserves explicit membership,
assignment, disablement, adoption, and bounded administration. No-bypass is already
agreed; the new proposed choice is the repeated-initialization outcome.

**Alternative / trade-off:** reapplying seed configuration may ease recovery but
could restore intentionally withdrawn authority. With the recommended no-mutation
behavior, recovery needs a separately governed process. Partial failure,
concurrency, reliable completion evidence, and recovery contracts remain open.
No status field, wire error code, or recovery API is introduced.

**Q-116:** once initial setup completes, should repeated bootstrap report already
initialized without changing grants or administrator membership?

</details>

## Q-117 — Authority visibility during incomplete setup (proposed, not approved)

**Context:** Q-116 settles repeating completed setup, not a crash during setup.
Q-110 preserves checked authority through an ordinary Auth write. Bootstrap
must additionally establish the initial registry, root grants, group, user,
assignments, and membership coherently. This question concerns when the new
bootstrap authority becomes usable, not how many database transactions to use.

**Recommendation:** make initial authority usable only when the entire intended
bootstrap arrangement has been validated and its successful establishment is
durably recorded. Persisted partial setup records must not expose a usable
partial administrator route. Publishing that authority and establishing completion
must be indivisible from the perspective of authorization consumers.

**Example:** the user, group, and root grant are stored, but setup fails while
establishing membership or recording completion. The stored records do not yet
give the human bootstrap authority. A trusted continuation must validate the
current intended arrangement before activation. Conversely, if setup completed
but the response was lost, Q-116 governs the retry: no repeated authority mutation.

**Rationale / core-philosophy check:** fail-closed establishment prevents a setup
that appears unfinished from already exposing powerful authority. Registration,
explicit assignment/membership, tenant bounds, and checked-state consistency
remain mandatory. No business-logic condition, new grant status, endpoint-policy
field, or evaluator bypass is introduced by this proposal.

**Alternative / trade-off:** exposing each completed write immediately could
make setup simpler, but allows partial authority before coherent completion.
The recommendation requires coordinated persistence/visibility and reliable
completion evidence. It does not require deleting partial user/group records or
one transaction spanning an external identity provider. Storage mechanics,
trusted continuation, concurrent attempts, and recovery remain open. This rule
concerns new bootstrap authority, not unrelated existing access of the human.

**Q-117:** should initial administrator authority become usable only after the
complete bootstrap arrangement is validated and durably established?
