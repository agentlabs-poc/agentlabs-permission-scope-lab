# Team administration — Q-092

Status: **AGREED.** The user specified that team creation covers teams and
subteams, that initial create/write/delete authority includes membership changes,
and that assigning grants needs a separate grant. After the revised JSON and
boundary explanation, the user answered “agreed, move next.”

## Initial administrative grant

```json
{
  "version": "1",
  "id": "G-TEAM-ADMIN",
  "permissions": [
    "auth:group::create",
    "auth:group::write",
    "auth:group::delete"
  ],
  "scope": {}
}
```

This recipient-free definition supplies authority only through an authorized
assignment. In the bootstrap example, it is assigned to Maya using the
[canonical assignment model](grant-assignments.md). Tenant is implied.

| Permission | Authorized operation, within the administrative scope |
|---|---|
| `auth:group::create` | Create teams or subteams. |
| `auth:group::write` | Update team details and manage human membership. |
| `auth:group::delete` | Delete teams. |

`scope: {}` makes this tenant-wide team administration, not authority limited
to teams Maya created. A narrower administrative scope requires explicit
definition; ownership is not an implicit boundary or evaluator bypass.

## Assignment is separate

[Q-093](assignment-authority.md) now settles the assignment rule: administrative
operation/recipient authority plus a valid parent authority route available to
the assigner. The concrete dependency contract remains open. [Q-095](authority-lineage.md)
also requires all authority assigned to a child team to remain within its parent
team's authority; wider rights of the administrator do not bypass that ceiling.

These permissions do not authorize attaching or changing business grants or
their assignments. Creating a subteam does not automatically establish inherited
authority. Establishing or changing its dependent authority route requires
separate assignment authorization, with its applicable boundaries validated.
Ordinary team writes cannot be used to bypass that check by changing authority
relationships as though they were descriptive metadata.

Membership administration does intentionally distribute the team's existing
access: adding Nutan to a team makes its valid assignments applicable through
her membership. Separate grant-assignment authority prevents changing the team's
assigned authority; it does not prevent an authorized membership administrator
from distributing that already-assigned access. This is powerful authority, not
mere contact-list editing. Membership alone still does not authorize administration.

## Rationale and preserved alternatives

The coarse create/write/delete model keeps team administration simple and includes
membership management in write. Separating assignment authority keeps team
structure and membership operations distinct from choosing the permissions and
scope a team receives. Each protected endpoint still declares one required
permission; bundling three permissions in a grant does not change that rule.

Earlier proposal, superseded for this initial bootstrap grant: seed only
`auth:group::create`, then separately discuss membership-management authority.
The user expanded the initial grant to the three permissions shown above.
Earlier scratch examples using `auth:group:membership::add` and `::remove`
remain historical proposals; they are not newly approved aliases for write.

## Still open

[Q-099](ownership-lineage.md) separates explicit owner changes from the selected
team-held supporting lineage. It does not decide whether `auth:group::write`
authorizes ownership transfer; that operation's exact permission remains open.
Ownership is not an implicit grant-assignment permission.

The full assignment/dependency contract following Q-093's governing rule;
narrow team-administration scope encoding and descendant coverage; owner records
and automatic owner assignment; detailed deletion/dependency handling; bootstrap
trust, exact full seed set, and repeated initialization. This is not a complete
bootstrap procedure, runtime implementation, or closure of administrative containment.
