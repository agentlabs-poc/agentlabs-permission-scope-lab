# 3. Relationships, inheritance and delegation

[Contents](../README.md) · [Previous](02-authority-and-boundaries.md) · [Next](04-resolution-and-enforcement.md)

**Definitions and representations alongside this chapter:**
[teams, membership and ownership](canonical-terms.md#teams-and-administration),
[parent/child grant and assignment JSON](canonical-terms.md#dependent-relationships),
[delegation](canonical-terms.md#delegation),
[enablement, validity and orphans](canonical-terms.md#lifecycle).
Pending relationship formats are distinguished from approved grant records.

## Draw different relationships as different edges

Authorization often depends on relationships, but not every relationship has the
same effect. A membership connects a person to a group. An assignment connects
authority to a recipient. A parent relationship can impose an inherited ceiling.
An ownership relationship can confer administrative responsibilities. Delegation
connects an actor's authority to another source under specified limits.

Using one vague term such as access relationship for all of these makes ordinary
changes difficult to predict. Removing a group owner should not be interpreted
as deleting the group's business grants unless the model explicitly says so.
Adding a group member should not silently make that person an administrator.

| Relationship | Question it answers | A conclusion it does not establish by itself |
|---|---|---|
| Membership | Which group-held authority is applicable to this person? | The person may change the group. |
| Assignment | Which recipient receives this authority definition? | The grant is enabled and supported. |
| Grant lineage | Which parent constrains this derived authority? | A stored parent definition is currently usable. |
| Team lineage | What boundary must the child team's authority respect? | Child members are also parent-team members. |
| Ownership/administration | Who may perform specified management operations? | The owner's personal business scope belongs to the team. |
| Delegation | On whose authority is this actor operating, and under what limits? | The actor has independent authority if required support disappears. |

The exact effects depend on the model. Some systems intentionally inherit group
membership through nesting. The framework in Part II does not: its team hierarchy
constrains authority without implicitly adding human memberships.

## A worked dependent hierarchy

Consider a parent team with Finance read/write authority. A child team receives
a derived read grant with an additional certificate restriction. Nutan receives
the child team's authority through membership. The following diagram illustrates
the **specific model used in Part II**, not a mandatory arrangement for all systems.

![Model example: separate team and grant parentage connected by assignments, with explicit membership](../assets/model-lineage.svg)

There are two ceilings to consider. The child grant stays within its parent
grant. The child team's total authority stays within the parent team's authority.
Checking only one narrow child grant is insufficient if another assignment can
give the same child team unrelated, broader access.

The example deliberately does not make Team1 synonymous with the Finance
department. A team is an authorization relationship; FIN is an application
boundary selected by a grant. A different application may have no department
concept at all.

## The issuer and the continuing source are different

Maya may need authority to create an assignment at a particular moment. What
keeps that assignment supported afterward is a separate lifecycle question.
An implementation should not infer permanent dependence on Maya merely because
her name appears in the creation history.

Equally, an actual personal dependency must not be discarded just because the
recipient is a team. If a route genuinely requires a person's current authority,
that support remains relevant. The design must distinguish established live
dependencies from historical attribution.

This distinction matters during ownership rotation. Replacing Maya with Om may
change who can administer Team1 without changing Team1's assigned business
authority. Importing Om's broader personal grants would be a separate authority
change, not a harmless consequence of editing an owner list.

## Delegation preserves the source and the limits

A delegated actor needs a clear authority anchor. Suppose Vinay may read and
write Finance records but delegates only read to Agent A. The agent's usable
authority cannot exceed either the source authority or the delegation's limits.
If Finance support disappears, an old delegation record cannot manufacture it.

Attribution should retain the actual actor as well as the supporting identity.
Otherwise a downstream component may treat a restricted program as if the human
were acting with unrestricted authority. Preserving a familiar human identifier
for compatibility does not remove the need to enforce the agent's limits.

Whether a system supports independent machine principals, delegation chains,
multiple authority anchors or automatic restoration is a design choice. Part II
chooses human-dependent agents and service accounts, direct delegation only in
v1, and restoration of still-valid delegated access when required human support
returns. Those choices must not be presented as universal definitions of service
accounts or delegation.

## Stored state and effective authority

An enabled assignment is a stored administrative state. Effective authority is
the outcome of all required current conditions. An enabled descendant may become
ineffective because its parent was disabled, without having its own state edited.

This separation lets a design represent a temporary interruption without
rewriting an entire graph. It also requires precise restoration rules. Returning
support and explicitly enabling a disabled record are different events. A system
must say which one is needed in each case.

An orphaned dependent route lacks required parent support. It may remain stored
for inspection or authorized repair while providing no access. Automatic deletion
or rebinding is not inherent in the word orphan. Nor does a failed lookup prove
that a route is orphaned: inability to inspect a relationship differs from an
established absence.

**Chapter takeaway:** the graph is part of the authorization model. Draw the
relationships precisely, then explain which edges are required now, which are
only historical, and which changes can affect downstream authority.

[Next: resolution, decisions and enforcement](04-resolution-and-enforcement.md)
