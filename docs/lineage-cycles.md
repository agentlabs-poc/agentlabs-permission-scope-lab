# Lineage cycle integrity — Q-111

**AGREED AT RULE LEVEL.** The user answered “agree” after reviewing the loop
examples, disabled-record rule, rationale, and restructuring trade-off.
This moves from revision lifecycle to the integrity
of the two parent relationships: grants and teams. [Q-095](authority-lineage.md)
defines dependent child authority; [Q-101](parent-grant-bindings.md) permits
structural changes when affected bindings are removed or disabled. Neither is
permission to create circular support. Exact cycle-validation contracts remain
open in the existing handbook; approval settles the prohibition, not those
implementation/representation contracts.

<details>
<summary>Earlier proposal status — superseded by explicit Q-111 agreement</summary>

Previous status: PROPOSED / NOT APPROVED. The recommendation was to reject
self-parenting and ancestor loops in both grant and team lineage, even while
the relevant records or bindings are disabled. That rule is now agreed.

</details>

## Agreed rule and examples

A grant must not be its own parent or become its own ancestor through a chain
of parent grants. Apply the same rule separately to team parentage. Reject a
proposed structural change that would create such a loop, even when relevant
grants or assignments are disabled. Disabling removes usable authority, not the
need for structurally valid parent relationships.

Here each arrow means “depends on parent,” not membership or ownership:

```text
G1 → G2 → G3 → G1       REJECT: grant ancestor loop
Team1 → Team2 → Team1  REJECT: team ancestor loop
G1 → G1               REJECT: direct self-parent
```

The recommendation is a graph-integrity rule, not a new entity, scope key,
assignment field, or ownership relationship. Valid non-circular parent changes
remain possible under Q-101's binding and boundary checks.

## Rationale, philosophy, and trade-off

Required support cannot be justified by a chain that eventually cites itself.
Rejecting loops preserves top-down resolution and dependent subset authority.
Disabling a loop would hide its immediate access effect but leave invalid
structure for a later enable/adoption attempt. The trade-off is that even an
inactive edit must preserve an acyclic hierarchy; administrators may need
additional valid intermediate steps for restructuring.

Alternative considered: allow circular structures while disabled and reject
only at enablement. This is not recommended because disabled records remain
real dependencies and would require separate valid/invalid structural modes.
No new draft mode or unlimited hierarchy-depth guarantee is proposed.

## Revision and validation limits

Grant parent references live in immutable revisions. Evaluate the relevant
proposed/adopted lineage, not an artificial union of every historical revision's
parent edges. An invalid self-parent reference cannot become acceptable by
publishing a new revision; an adoption or reparenting must not create a cycle in
the resulting relevant lineage. Current supporting revisions still follow
Q-103's actual-route rule. Full revision-aware graph selection, direct-human
support selection, depth bounds, and error representation remain open.

This proposal does not make cycles into bootstrap roots, orphan grants, or
independent authority. It also does not claim that absence of grant/team cycles
alone proves every combined assignment/dependency case safe. Defensive runtime
handling of corrupt stored graphs must not treat a loop as proof of support;
its exact decision/error contract remains to be specified.

**Q-111, answered “agree”:** approve rejecting self-parenting and ancestor loops in grant and team
lineage, including when the relevant records or bindings are disabled?
