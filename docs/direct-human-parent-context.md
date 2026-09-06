# Direct-human parent context — Q-112

## Q-112A — Lineage-supported latest reaffirmed; new field withdrawn

The user corrected the proposed independent parent-revision adoption: “for
Q-112A i remember it was discussed earlier its linage supported latest.”
Treat this as reaffirmation of Q-103, not a request for another selector.

Resolve the parent context from **current valid adopted support in the actual
required lineage**, top-to-bottom. The parent supplies the maximum permission
and scope context; the child narrows it. “Latest” is lineage-supported, not
automatically the newest published revision or a broader revision adopted by
an unrelated recipient. Child/assignment revisions, status, validity, tenant,
and required live support remain mandatory.

The assistant's `parent_grant_revision` proposal is **withdrawn / not adopted**.
Keep `parent_grant_id` as the explicit grant-lineage link. There is no new
independent child-selected parent revision and no added parent-assignment ID.
Original Q-101 direct-human assignment-survival rules remain unchanged.

**Rationale / core-philosophy check:** reuse the agreed live supporting lineage
rather than add a second content selector that could diverge from it. The user
has reaffirmed the governing selection rule; do not keep asking that rule as
though it were undecided. This does not permit historical authority to outlive
required current support or allow recipient adoptions to supply unrelated reach.

**Remaining implementation/evidence work:** define how Auth establishes eligible
support and its current adopted revisions for each route, including a direct
human; validate ambiguous/corrupt input and produce the corresponding evidence
and failure result. The FIN/ENG comparison below supplies candidate revisions
but does not establish either as Nutan's actual supporting lineage. It therefore
does not justify choosing either one, merging them, or adding a canonical field.
The governing rule is reaffirmed; the complete resolver contract is not claimed
finished or proven by this small diagnostic.

All 26 scratchpad sources, including the original experiment and proposed JSON,
are [captured in lab history](history/scratchpad-import/README.md). Originals are
unchanged. Further exploration is lab-only. Archived draft status is historical,
not an alternative current model or approval of the withdrawn field.

<details>
<summary>Earlier Q-112 framing — superseded where it treats the governing selection rule as undecided</summary>

## User clarification recorded; revision-selection mechanism remains open

The user clarified: “as a best practice we should avoid direct assignment. how
ever if direct assignement is done. the rule is simple parent sets the context,
this is both permission and scope, that is the max boundry.”

Prefer human access through group membership and group-held assignments. Direct
human assignments remain supported, not prohibited. This reaffirms GROUP-004
rather than introducing a group-only authorization rule.

For a directly assigned dependent grant G2 with parent G1, the parent supplies
the maximum permission and scope boundary. G2's permissions select a subset;
its additional scope is ANDed with the inherited parent boundary. Direct
assignment does not turn G2 into independent authority or permit broader reach.
Current validity, grant/assignment status, and required upstream support still
apply under the existing rules.

| Established parent context | Child selection | Consequence |
|---|---|---|
| FIN read/write | Read, additional scope empty | FIN read only, not tenant-wide. |
| FIN read/write | Read, additional certificate C17 | FIN AND C17 read, when that boundary is supported for the same data. |
| FIN read | Write | Cannot establish valid write authority. |
| FIN read | Read with additional department ENG | No override; the conflicting constraints cannot authorize FIN-or-ENG access. |

**Rationale / philosophy:** group-based assignment makes shared human access
administration simpler, while a legitimate direct assignment remains governed
by the same dependent-subset model. The parent, not recipient placement, defines
the ceiling. The trade-off is that supporting direct assignments still requires
an unambiguous lineage contract; a best practice does not remove that case.
Membership and ownership are not merged or invented as a consequence.

## What this clarification does not settle

[Q-107](grant-revision-format.md) distinguishes grant identity from immutable
revisions. The parent's identity therefore does not by itself select its
permission/scope content when different supporting assignments have adopted
different revisions. [Q-103](grant-revisions.md) selects the current adopted
revision in actual supporting lineage; a subteam supplies parent-team context.
The equivalent direct-human selection mechanism remains open.

In the motivating example, Team1 has G1 revision 1 (FIN read), TeamX has G1
revision 2 (ENG read), and G2 with parent G1 and empty additional scope is directly
assigned to Nutan. Both parent revisions are assumed otherwise valid under their
own upstream support. The question is which context G2 actually inherits, not
whether it may exceed that context once established. No permanent dependency on
issuer Maya or assignment A1 is inferred; [Q-101's direct-human case](parent-grant-bindings.md)
remains intact.

The assistant's recommendation remains to use current adopted authority in the
actual required supporting lineage, not automatically the latest published
revision or a union of unrelated team holdings. This is not approval of a new
parent-revision field, source-assignment ID, ownership record, or universal
tenant-wide search for supporting assignments. If the required support cannot
be established, it cannot serve as proof of authority.

Q-112 is recorded as a boundary clarification, **not closure of the direct-human
parent-revision-selection checkpoint**. No canonical representation is added,
and no earlier assignment-survival rule is deprecated by implication.

</details>
