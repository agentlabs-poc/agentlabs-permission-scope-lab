# Authorization groups and membership

Q-045 approves GROUP-001-A and GROUP-003. The user also reaffirmed PROCESS-003:
recording must capture rationale, not just the approval. This chapter therefore
preserves the choices, reasons, examples, safety consequences, and open details.

## Agreed policies

1. **Auth owns authorization groups and memberships.** Applications may
   synchronize business membership into Auth or keep business groupings separate.
   Application business membership is not automatically authorization membership.
2. **Groups contain humans only.** Service accounts and agents are not
   first-class group members. Their access remains human-dependent and bounded
   by the authority available through that human.

Team and group are synonymous (TERM-001). Group-based human grants remain
preferred, while direct human grants remain supported (GROUP-004). Those
earlier agreements are retained rather than reopened by Q-045.

## Why Auth owns the authorization membership

The shared authorization layer needs one authority for whether a human is a
member of an authorization group. Allowing each application to infer that
membership independently from its own business data would leave the same grant
dependent on potentially different membership interpretations.

An application's departments, project teams, or reporting relationships are
business facts and may serve purposes unrelated to access. Optional sync lets
an application deliberately connect those facts to authorization membership;
keeping them separate lets business organization and access administration
evolve independently. Neither choice makes arbitrary business membership an
implicit access grant.

The earlier alternative, GROUP-001-B, allowed application-owned authorization
membership through resolver integrations. Q-045 selects Auth ownership instead.
This does not prevent applications supplying ordinary business facts required
for scope evaluation, and does not prescribe a network call for every request.

## Why membership is human-only

The agreed automation model is a dependent subset of human authority, not an
independent source of access. First-class automated membership would create
another route that would need its own dependency rules to avoid bypassing the
human's restrictions. Human-only membership keeps the group route through the
human, with the proxy's additional limits enforced on top.

For example, Vinay belongs to `payroll-readers`. His agent may exercise permitted
delegated access derived through Vinay; the agent does not join the group. If
Vinay loses that supporting membership, the group-derived route is no longer
available to his agent. Other genuinely applicable routes are not erased by
an unrelated membership change. Exact freshness mechanics remain open.

## Examples and counterexamples

| Situation | Consequence under the agreed model |
|---|---|
| Vinay is in an application Finance department but not the Auth payroll group | Department membership alone does not make the payroll group's grants applicable. Other valid direct/group authority is evaluated separately. |
| The application deliberately syncs selected Finance employees into the Auth group | Authorization uses the resulting Auth membership under its freshness contract; sync does not itself invent permissions or widen grant scope. |
| Business and authorization groups remain separate | An application business-group change does not automatically change Auth membership. |
| An agent requests its own first-class group membership | That membership route is not supported. Its access must remain through the human-dependent model. |
| An employee-group grant has `{"user":"$self"}` scope | Self resolves for each human, not for the group collectively; an automated proxy retains that human anchor. |

A configured sync does not imply instantaneous revocation. Until changes reach
Auth and relevant evaluation state, stale membership is a risk the future sync
and freshness contract must address. This chapter does not claim that default
denial can detect a change the evaluator has not learned about.

## Lifecycle and authority limits

Membership changes require their own administrative authorization. Permission
to assign a grant to a group does not itself permit joining or editing that
group. Group ownership also does not make a group's creator implicitly entitled
to every permission granted to it.

The group grant and membership remain dependencies of group-derived authority
(RESOLUTION-004). Proxy limits remain in force independently of group membership.
These policies do not permit independent service-account authority.

Nested-group mechanics, membership synchronization guarantees, exact lifecycle
representation, freshness, and delegation encoding remain open. Q-045 settles
ownership and member types, not these mechanisms or the whole identity glossary.

## Q-077 / GROUP-005 — direct human membership only in v1

Status: **PROPOSED, not approved.** Human-only authorization membership is already
agreed. This settles the separately open question of transitive membership through
group nesting, not whether agents may join groups.

Recommend no nested authorization groups in v1: only explicit human-to-group
membership links participate in grant resolution. Vinay may belong to several
groups, but putting Finance inside Employees must not automatically make him an
authorization member of Employees. To use Employees' grants, establish his own
authorized membership there. An application may manage this through deliberate
sync, as already allowed; no implicit business-hierarchy inheritance is introduced.

Rationale: direct membership keeps applicable group grants and their dependencies
explicit, without recursive traversal, cycle rules, or additional transitive
revocation paths. The alternative is nested groups with transitive human access,
which reduces membership administration but needs those additional semantics.
The trade-off of exclusion is maintaining multiple explicit memberships.

No new membership wire format, implicit grant, independent proxy authority, or
restriction to only one group is proposed. Membership lifecycle, authorized sync,
and freshness remain open; HC-05-12 is not closed by this decision alone.

**Q-077:** Should v1 use only direct human-to-group memberships, with no nested
group membership inheritance?
