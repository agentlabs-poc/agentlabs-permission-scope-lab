# Q-089 — original live-role question, superseded by Q-089-B

Later update: [Q-090](grant-assignments.md) separates grants and assignments;
recipient-bearing JSON retained in this historical discussion is not current
format. Q-089-B's role revision/adoption safety remains, with shared-definition
revision mechanics still open.

**Current decision: Q-089-B / ROLE-003 is AGREED.** Role definitions use immutable
published revisions; grants explicitly adopt a revision after authorization and
boundary validation. No role edit silently updates existing grants. See
[the approved revision/adoption model](role-revisions.md) for canonical excerpts,
the Auth-self-protection context, rationale, and consequences.

The original question below was left undecided at the user's request. Neither
of its alternatives was later adopted; Q-089-B replaces their live-update premise.
The intervening discussion was not recorded while recording was frozen. The
user's explicit “record this commit and push” now authorizes this decision record
and reconciliation checkpoint.

<details>
<summary>Original Q-089 question, examples, rationale, and undecided status — preserved unchanged as history</summary>

# Q-089 — authority to expand a live role

Status: **UNDECIDED — retained for further user discussion.** The user instructed
that the question and rationale be registered and left undecided because he has
further points to discuss. Neither alternative is approved or rejected.
Previous status, retained as history: **PROPOSED, not approved** before this instruction.
This is an administrative authorization
decision in the grants/roles branch, not a business-rule mechanism or a finalized
role wire schema.

## Already agreed and the remaining question

ROLE-002 says existing referencing grants follow the role's current permission
bundle. Each grant retains its own recipient, scope, validity, and dependencies.
Role-edit authorization remains open; live updates alone do not decide who may
make them.

ADMIN-004/005 require ordinary grant assignments to fit the administrator's
applicable authority. This proposal asks how that protection applies when adding
a role permission expands existing assignments without editing their grant records.

## Concrete example and illustrative role JSON

Role R-17 initially contains only `hrms:employee:certificate::read`. A proposed
edit adds write:

```json
{
  "version": "1",
  "id": "R-17",
  "permissions": [
    "hrms:employee:certificate::read",
    "hrms:employee:certificate::write"
  ]
}
```

This is an illustrative role-definition excerpt using the existing role example
fields plus the mandatory contract version, **not approval of a complete role
schema or an endpoint request format**. The rule, not additional fields, is being
proposed. The corresponding current role omits the write permission.

Two existing grants reference R-17:

| Grant | Recipient | Scope | Added authority if the edit succeeds |
|---|---|---|---|
| G-17 | Finance readers group | `{"dept":"FIN"}` | Certificate write within Finance |
| G-18 | Engineering readers group | `{"dept":"ENG"}` | Certificate write within Engineering |

Assume Vinay can edit the role and administer the proposed Finance write
assignment, but cannot administer the Engineering write assignment. Merely
checking that he can edit R-17 would miss the second authority expansion.

## Expanded explanation — what actually changes

The role supplies **which permissions**. Each grant separately supplies **who
receives them and within which scope**. R-17 does not itself contain a Finance
boundary. Reusing it in a Finance grant does not turn it into a Finance-only role.

For example, G-17's illustrative grant excerpt is:

```json
{
  "version": "1",
  "id": "G-17",
  "recipient": {"type": "group", "id": "finance-readers"},
  "role_id": "R-17",
  "scope": {"dept": "FIN"},
  "status": "enabled"
}
```

This reuses the current working role-referencing grant layout; it does not finalize
the full schema. G-18 separately references the same R-17 for Engineering readers
with scope `{"dept":"ENG"}`. Assume the grants, memberships, and other mandatory
requirements are valid for this example.

Before the role edit, both groups derive read permission, each inside its own
grant scope. After adding write to R-17, both derive read and write inside those
same scopes. Neither grant record needs to change: the live role reference supplies
the new permission. Finance does not gain Engineering access; Engineering receives
new write capability inside Engineering. The expansion is in permissions, not
cross-department reach.

Vinay's personal read/write access is not the question. His **administrative**
authority is. In this example it permits providing Finance write access to the
Finance group, but not Engineering write access to the Engineering group.

The Q-089 recommendation would stop the shared-role edit before publishing it,
because it creates both effects. It would not publish the edit and then ask every
application request to recheck Vinay's administrative rights. If appropriately
authorized, the edit becomes the current role definition and ordinary resolution
uses it. Existing human/delegation limits remain applicable.

The simpler alternative deliberately defines role-edit authority as sufficient
authority for those effects wherever the role is already assigned. That is not
inherently an invalid design: it is a stronger meaning for the role-edit privilege.
The question is which administrative authority model we intend, not whether live
role updates should happen at all. The latter is already agreed.

No extra endpoint permission declaration or actual grant-create operation is
adopted by this explanation. The protected role-change endpoint still declares one
permission. The proposal concerns the bounds under which that operation may
expand authority; their exact canonical encoding remains open.

## Earlier recommendation — not adopted; decision remains open

Authorize a live-role permission expansion only when the editor has the required
role-edit authority **and the added authority through all affected referencing
grants fits his applicable administrative bounds**. Keep permission, recipient,
scope, and other relevant assignment bounds associated; do not manufacture wider
administrative authority by combining unrelated grant fields.

In the example, reject Vinay's shared-role edit and leave R-17 unchanged: it would
also provide Engineering write access beyond his administrative authority. An
administrator with appropriate authority across both affected assignments could
perform the change. This is not a requirement that the administrator personally
possess the certificate-write business permission.

Do not silently update only selected referencing grants or narrow their scopes;
that would change the agreed live-role semantics. Separate roles/assignments may
be appropriate when independently administered populations need different
capabilities, but creating or changing those bindings still needs authorization.

## Rationale, trade-off, and alternatives

The same effective authority expansion should not become possible through role
editing when it would be forbidden through direct grant administration. Otherwise,
roles offer an indirect route around the administrative bounds already agreed.

The alternative is to treat permission to edit a role as authority to change
capabilities for every current referencing grant, regardless of the editor's other
assignment bounds. That is simpler but makes role editing a potentially broad
provisioning power that must be intentionally granted and understood. The
recommendation keeps the shared administrative bounds applicable to expansion.

The trade-off is that a shared role spanning separately administered populations
may need an administrator authorized across the affected populations to expand it.
This is an authorization requirement, not a mandatory full grant scan on each edit;
indexing, impact representation, and validation mechanisms are not selected here.

## Architecture and open details

The check belongs to authorization of the Auth role-change operation, before
publishing the changed definition. Application business endpoints continue normal
resolution; they do not inspect who edited the role or implement this admin check.
No additional runtime authorization gate is introduced.

The proposal concerns adding permissions. Permission removal, concurrent role/grant
changes, retained but disabled references, registration compatibility, revision
evidence, and propagation timing still need precise treatment. The rule must not
be assumed to solve those cases or authorize latent expansion without checks.
HC-05-13 remains open. No complete role schema or containment algorithm is adopted.

**Q-089:** Should expanding a live role require authorization for the added
authority through all affected referencing grants, rather than role-edit
permission alone?

**Decision: UNDECIDED at the user's request.** The alternatives remain:

1. Constrain a role expansion by the editor's administrative authority over the
   resulting assignments through affected grants.
2. Define authority to edit the role as sufficient authority for its permission
   changes wherever it is already assigned.

The shared-role example, rationale, and trade-offs above are retained as discussion
material, not canonical rules. Await the user's additional points before advancing
this decision. Existing live-role semantics remain agreed; HC-05-13 remains open.

</details>
