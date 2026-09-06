# Grant lifecycle — canonical discussion

## Q-082 / GRANT-008 — delete is the single permanent-removal operation

Status: **AGREED.** The user questioned whether revoke was technically delete,
then answered “agree” to consolidating the two into one canonical operation:
**delete**. The separate revoke operation is superseded, not an additional API
operation or permission alias. Its permanent-withdrawal meaning survives in delete.

The current lifecycle operations are:

| Operation | Canonical meaning |
|---|---|
| Create | Establish a new grant through authorized administration. |
| Enable | Make an existing disabled grant eligible for evaluation again. |
| Disable | Reversibly stop this grant supplying authority. |
| Delete | Permanently remove this grant from usable authority. Restoring access requires a newly authorized grant. |

Rationale: revoke and delete had the same authorization effect. The distinction
introduced earlier concerned keeping a marked record versus removing it from
the usable collection, not a separate authorization capability. Storage and
retention choices do not justify a second canonical operation. One removal
operation avoids duplicate lifecycle transitions and a redundant revoked state.

Example: disabling G-17 lets an authorized administrator later enable G-17.
Deleting G-17 does not: Finance access must be established through a new grant,
for example G-18, authorized against current administrative bounds. Matching its
old permission and scope does not revive the deleted grant identity. A retained
internal record is not usable authority and cannot be enabled back into existence.

All four operations require appropriate administrative authorization. Enable is
not allow: scope, permission, validity, conditions, recipient dependencies, and
human/proxy limits still apply. Deleting or disabling one grant does not veto an
independent valid route. Q-070's automatic return of a still-valid delegation
when human support returns remains distinct from enabling a disabled grant or
restoring a deleted one.

Q-078's permanence rule is retained under delete; Q-079's reversible suspension
and Q-080's creation/removal remain agreed. Q-081's original three-value status
proposal is **SUPERSEDED, never approved**. The revised representation below
remains proposed; Q-082 answers the operation question, not the complete schema.

Earlier uses of grant “revocation” in Q-069/Q-075 describe authority withdrawal;
the canonical permanent-removal operation is now delete. Their no-stale-use and
execution-time-check guarantees are not weakened by the terminology consolidation.
Business permissions named revoke in domain examples are unrelated operation
names and are not renamed by this grant-lifecycle decision.

## Q-081 revised / GRANT-007 — enabled or disabled administrative status

Status: **PROPOSED, not approved.** Reuse the illustrated `status` field with only
`enabled` and `disabled`. No canonical `revoked` or `deleted` status is required:
delete removes the grant from usable authority; implementation bookkeeping is
outside this grant representation. No physical deletion requirement is implied.

Versioned grant illustration; not the complete grant schema. Validity and
condition details are omitted for focus, not removed from the model:

```json
{
  "version": "1",
  "id": "G-17",
  "recipient": {"type": "group", "id": "finance-readers"},
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"dept": "FIN"},
  "status": "enabled"
}
```

Rationale: this field represents the reversible administrative switch, not a
request decision or a claim that all dependencies are valid. The earlier `active`
spelling remains an illustrative historical convention, not a second canonical
value. Initial-state defaults, full schema validation, expiry/renewal, and exact
administrative operation contracts remain open; HC-05-11 and HC-07-08 stay open.

**Q-081 revised:** Should `status` use only `enabled` and `disabled`, while delete
removes the grant and effective authorization remains evaluated?

<details>
<summary>Historical Q-078–Q-081 record — separate revoke and three-state proposal superseded by Q-082</summary>

The original text below is preserved. Its five-operation table and three-state
proposal do not override the current four-operation model above. Other agreed
meanings survive as explained in Q-082.

## Q-078 / GRANT-004 — explicit revocation is terminal for that grant

Status: **AGREED.** The user answered Q-078 “yes,” then separately added reversible
enable/disable operations and create/delete operations below. Explicit revocation
is terminal for that grant; these additions do not turn revoke into disable.
This is not an audit-system policy or a complete lifecycle wire schema.

Previous status, retained as history: **PROPOSED, not approved** until that answer.

### Agreed example and outcome

G-17 grants Vinay the registered certificate-read permission within Finance.
An authorized administrator explicitly revokes G-17. Later, Finance access is
needed again.

G-17 remains revoked. Restoring access requires an explicitly
authorized new grant, for example G-18, evaluated against the administrator's
current bounds. Do not make G-17 usable again merely by editing its status.
The replacement has its own identity and binding; matching permission and scope
does not make it the old grant. This neither authorizes creation automatically
nor requires every restoration to reproduce the old grant's exact contents.

### Rationale, alternative, and trade-off

Terminal revocation gives “this grant was withdrawn” a stable meaning for
resolution and dependency handling. Reopening that same identity adds a lifecycle
transition and requires every consumer to distinguish its earlier revoked life
from its later active life. The alternative is an explicitly authorized reactivation
operation; it avoids a new grant record but needs those extra semantics. That
alternative is not adopted for explicitly revoked grants.

The trade-off of terminal revocation is a new grant when access must be granted
again. The decision does not demand deletion of the old record or define its
storage/retention policy. It selects authorization meaning, not an audit format.

### Distinctions and remaining decisions

Q-070 remains unchanged: a still-valid delegation may become effectively inactive
when its human loses supporting authority, then work again when support returns.
That temporary lack of support is not explicit revocation of the delegation or
of every underlying grant. Losing a group membership also does not itself revoke
the group's grant for all other members.

Other independently valid grants may still authorize access; revoking G-17 is
not a global deny on Vinay. Whether a replacement grant can satisfy particular
delegation supporting-reference constraints remains separate; no automatic
dependency rebinding is adopted here.

Historical open list at Q-078 proposal: status field names, temporary suspension,
expiry/renewal, record retention, administrative operation encoding, and remaining
lifecycle representation. Q-079 now supports reversible disable/enable; Q-080
supports create/delete. Retention belongs to the external layer under Q-076.
Expiry/renewal, operation encoding, and remaining representation stay open.
HC-05-11 is not closed by these governing operations alone.

**Q-078:** Should explicit revocation be permanent for that grant, with restored
access requiring a newly authorized grant rather than reactivating the old one?

**Answer: yes.** The rationale and distinction from dependent inactivity remain
part of the approved record.

## Q-079 / GRANT-005 — reversible disable and enable

Status: **AGREED — user-proposed addition.** The user added: “for operational
effecieny we can enable/disable a grant.” Disabling a grant stops it supplying
authority without permanently revoking it. An authorized enable operation can
make the same non-revoked grant eligible for evaluation again.

Example: disable G-17 during an operational pause; enable G-17 when the pause
ends. There is no need to create G-18 merely to end that pause. By contrast,
explicitly revoking G-17 under Q-078 ends that grant permanently.

Rationale: operational suspension and permanent withdrawal have different intent.
Using revocation for every temporary pause would require needless replacement
grants. The alternative of reversible revocation would blur Q-078's meaning.

Enable does not mean allow. The grant's own scope, permission, validity, conditions,
recipient dependencies, and applicable human/proxy limits still govern requests.
An enabled but unsupported delegation remains ineffective; returning human support
does not itself enable an explicitly disabled grant. Other complete valid grant
routes are not vetoed by disabling this one. These are consequences of the existing
whole-grant and dependent-authority rules, not new independent authority.

Enable/disable operations require administrative authorization within current
bounds. Exact operation permission names and propagation contracts remain open.

## Q-080 / GRANT-006 — create and delete are lifecycle operations

Status: **AGREED — user-proposed addition.** The user added: “offcouse there can be
delete and create as well.” A grant can be created or deleted through authorized
administration, in addition to being enabled, disabled, or permanently revoked.

Create establishes a new grant binding; the grant being created cannot authorize
its own creation. Delete removes the grant from usable authority. It is not a
temporary disabled state that enable can undo. Any subsequent grant creation must
be authorized in its own right; delete/create cannot revive an explicitly revoked
grant identity. No physical erase or historical retention requirement is adopted.

Rationale: the canonical lifecycle needs both establishment and removal of grant
bindings, not just toggling an existing grant. Operation meanings belong here;
soft deletion, physical storage, and historical records are implementation or
external-layer concerns. No new grant format or administrator bypass is introduced.

### Consolidated operation meanings

| Operation | Meaning | Effect on later authorization |
|---|---|---|
| Create | Establish a new grant through authorized administration. | Evaluated under its configured state and all restrictions; initial-state default is not yet chosen. |
| Disable | Reversibly suspend the same grant. | This grant supplies no authority while disabled. |
| Enable | End administrative suspension of a non-revoked grant. | Eligible for evaluation, not automatically an allow. |
| Revoke | Permanently withdraw this grant. | Cannot be enabled again; restoration needs a newly authorized grant. |
| Delete | Remove the grant from usable authority. | No authority via that grant; physical/historical storage policy is not prescribed. |

## Q-081 / GRANT-007 — representation of administrative state

Status: **PROPOSED, not approved.** Recommend reusing the existing illustrative
`status` field with three canonical administrative values: `enabled`, `disabled`,
and `revoked`. Create and delete are operations, not additional usable-grant
statuses. This does not require deletion to erase the physical record.

The earlier grant examples used `status: active` illustratively, not as an approved
enum. They remain history/working examples until a representation is adopted.
`enabled` is proposed because it does not imply that a grant is currently valid
or sufficient to allow any particular request.

Versioned grant illustration, not a complete finalized schema; optional validity
and condition details are omitted for focus, not eliminated from the model:

```json
{
  "version": "1",
  "id": "G-17",
  "recipient": {"type": "group", "id": "finance-readers"},
  "permissions": ["hrms:employee:certificate::read"],
  "scope": {"dept": "FIN"},
  "status": "enabled"
}
```

Rationale: one administrative-state field avoids combinations of overlapping
booleans such as enabled and revoked. Expiry and unmet dependencies remain
evaluation constraints rather than extra stored statuses. Default state at
creation, full transition validation, administrative API contracts, validity
representation, and the full grant schema remain open. No score is awarded for
closing those broader checkpoints with this partial representation.

**Q-081:** Should the grant use one `status` field with `enabled`, `disabled`, and
`revoked`, keeping create/delete as operations and effective validity derived?

</details>
