# Grant lifecycle — canonical discussion

## Q-078 / GRANT-004 — explicit revocation is terminal for that grant

Status: **PROPOSED, not approved.** Existing rules say a revoked grant cannot
authorize, and new checks cannot use it after Auth confirms revocation. This
question decides whether the same grant may later be reactivated. It is not an
audit-system policy or a full lifecycle wire schema.

### Example and recommendation

G-17 grants Vinay the registered certificate-read permission within Finance.
An authorized administrator explicitly revokes G-17. Later, Finance access is
needed again.

Recommend that G-17 remains revoked. Restoring access requires an explicitly
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
operation; it avoids a new grant record but needs those extra semantics.

The trade-off of terminal revocation is a new grant when access must be granted
again. The proposal does not demand deletion of the old record or define its
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

Status field names, temporary suspension support, expiry/renewal, record retention,
administrative operation encoding, and the remaining lifecycle representation
stay open. HC-05-11 remains open even if this rule is approved.

**Q-078:** Should explicit revocation be permanent for that grant, with restored
access requiring a newly authorized grant rather than reactivating the old one?
