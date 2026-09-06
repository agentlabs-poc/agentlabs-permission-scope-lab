# Direct-human parent context — Q-112

## C01-D2 — Trace of the existing records

**Diagnostic completed; no new selection policy approved.** The user agreed to
trace the actual lookup and evidence using existing fields. That approval does
not choose FIN, ENG, all parent holdings, a new parent field, or a new permanent
dependency. This section supplies the concrete trace requested after C01-D1.

### What changed between the original cases

The [original Q-101 example](history/scratchpad-import/q101-direct-human-dependency.md.txt)
used one G1 definition, with FIN read/write, at both Team1 and TeamX. Removing
A1 did not imply losing that unchanged authority where the stated valid support
remained. It did not test competing adopted content revisions.

[Q-102](grant-revisions.md) later permitted different assignments to adopt
different revisions. The [earlier Q-112 diagnostic](history/scratchpad-import/q112-parent-context-check.md.txt)
already exposed the FIN/ENG comparison. It was captured in the lab, not lost in
the scratchpad. Q-112A subsequently reaffirmed actual lineage-supported latest
and withdrew the proposed independent parent-revision field. It did not specify
the direct-human eligibility lookup. Do not present this as a newly discovered
subset-rule defect or repeat the withdrawn proposal.

### Records supplied to the trace

These are versioned content excerpts using approved fields, not complete API or
bootstrap contracts. Assume valid G0 upstream support, trusted common tenant,
registration, enabled controls and applicable time validity. G1 revision 1 was
adopted at Team1 before revision 2 existed; retaining it is valid under Q-102.

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 1,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"dept": "FIN"}
}
```

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 2,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"dept": "ENG"}
}
```

```json
{
  "version": "1",
  "grant_id": "G2",
  "revision": 7,
  "parent_grant_id": "G1",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {}
}
```

The relevant assignment facts, stated without inventing a new relationship format:

| Assignment | Recipient | Adopted content | State |
|---|---|---|---|
| A1 | Team1 | G1 revision 1 | Enabled and otherwise valid |
| AX | TeamX | G1 revision 2 | Enabled and otherwise valid |
| AN | Nutan, directly | G2 revision 7 | Enabled; its parent context is being investigated |

Maya's valid source memberships and separate administrative authority explain
issuance permission. They are not an implicit permanent Nutan-to-Maya or
Nutan-to-A1 relationship. No ownership record participates. The example does
not establish a parent-team relationship for Nutan or a membership requirement
for her direct assignment.

```text
Nutan ← AN → G2 revision 7
                 │ parent_grant_id = G1
                 ▼
          Reusable identity G1
          ├── A1 → Team1 adopts revision 1 → FIN read
          └── AX → TeamX adopts revision 2 → ENG read

Established: both holdings exist and adopt different content.
Not established: which holding is eligible support for Nutan's direct route.
```

### Lookup trace and limits

| Step | What the existing records establish | What they do not establish |
|---|---|---|
| Read AN | Nutan adopts G2 revision 7. | A parent revision numbered 7. |
| Read G2 revision 7 | Parent identity is G1; select read; add no local scope constraint. | A particular G1 holder or eligible set of holders. |
| Inspect G1 holdings | A1 adopts revision 1; AX adopts revision 2. | That either is Nutan's actual supporting lineage merely because it exists. |
| Validate both candidate holdings | Each is individually valid under the stated premises. | Recipient-route eligibility; source validity and eligibility are different tests. |
| Supply A1 to the local narrowing calculation | FIN read AND no additional constraint gives FIN read. | Proof that supplying A1 was authorized for Nutan's route. |
| Supply AX to that calculation | ENG read AND no additional constraint gives ENG read. | Proof that supplying AX was authorized for Nutan's route. |

In the separate **subteam** case, Team2's parent Team1 supplies the missing
context: the required G1 holding is at Team1, uniquely selected by Q-104. A1's
adopted revision 1 then establishes the ceiling. Nutan's direct case does not
contain that team relationship, so the subteam lookup cannot simply be copied.

### Findings and core-philosophy check

- `parent_grant_id` locates the parent identity. Actual assignments select
  adopted content, but discovering an assignment is not itself proof that it is
  eligible for this recipient route.
- Both supplied-parent calculations obey permission subset and scope AND, yet
  produce different boundaries. Narrowing does not select supporting context.
- Picking the numerically largest revision, the first database row, the issuer's
  historical membership or every tenant-wide holding would add a selection rule.
  Treating multiple complete parent routes as alternatives would also require
  establishing their eligibility; it is not approved merely because no fields
  are mixed inside an individual route.
- Removing A1 from this fixture leaves AX as the only found holding. That removes
  multiplicity, not the obligation to establish AX's eligibility. This does not
  overturn Q-101's original example, which explicitly assumes valid remaining
  support; it prevents extending its conclusion to unstated support semantics.
- Missing proof must not become an allow, stored disablement, automatic rebinding
  or an assertion of orphanhood. Whether support is absent or evaluation cannot
  establish it still matters. Exact failure representation is not decided here.

**Result:** the lookup trace is complete for the supplied records, but the direct-
human support-selection contract is not. The remaining question is what makes a
holding eligible as continuing support for a direct-human child—not whether the
child must stay within its parent. More lookups of the same facts cannot settle
that meaning. No schema change is proved necessary by this diagnostic.

Verification: eight in-memory diagnostic assertions passed using the three
excerpts above and the stated holding facts. Checks covered versions/no extra
selector, FIN and ENG calculations, distinct outcomes, child versus parent
revision numbers, unique Team1 lookup, the remaining holding after A1 removal,
and equal results in the original same-content comparison. These test supplied-
parent calculations and the separately established subteam lookup, not an
implemented direct-human eligibility rule. C01-D1's other supported behavior
remains usable.

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
