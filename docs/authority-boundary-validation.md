# Authority-boundary validation — C01 working contract

**C01-D1: behavior consolidation from approved decisions, not a new policy or
wire schema.** This chapter completes the combined validation flow and worked
subteam case that were previously scattered across Q-093–Q-130. Source decisions
remain authoritative. The final section identifies the precise limits of this
draft; it does not silently approve missing representation choices.

## 1. Where this runs

This runs **inside Auth Service**, in the handler of an Auth administrative API.
It does not ask a customer application's endpoint to approve Auth's writes.

![Auth's administrative evaluator and authority-boundary validator before persistence](assets/auth-service-authority-gate.svg)

The Auth endpoint declares one administrative permission and its input sources.
The administrative evaluator checks that operation within its administrative
boundary. The authority-boundary validator checks the authority being distributed.
Only then may the handler persist the checked change, preserving its dependencies
through the write. This is one endpoint-owned gate with two responsibilities,
not two independent approvals that a caller can reuse separately.

| Check | Concrete question | Why it cannot replace the other check |
|---|---|---|
| Administration | May Maya create an assignment to Team2? | Team2 administration does not supply FIN read or ENG write. |
| Proposed authority | Does the complete proposed grant/assignment remain within its permitted parent support and team ceiling? | Having FIN read does not authorize Maya to administer Team2. |

Sources: [Q-093](assignment-authority.md), [Q-100](auth-service-authority-gate.md),
[Q-110](auth-write-consistency.md). Auth can share resolution primitives between
these responsibilities; no deployment boundary or mandatory network call follows.

## 2. Required material and who establishes it

These are semantic input requirements, **not names for new JSON fields**.

| Material | Established from | Required use |
|---|---|---|
| Acting human and any direct proxy | Verified identity and delegation context | Evaluate current administrative authority and applicable human ceiling; never trust a body-supplied identity as authentication. |
| Enclosing tenant/application context | Trusted endpoint/context binding | Keep ordinary authority within its enclosing boundary. Platform publication uses its separately authorized management context. |
| Exact operation, recipient and proposed result | Endpoint declaration and validated request | Bind the operation to the actual record written; validate the whole result, not only a patch's added fields. |
| Grant controls and immutable adopted content | Auth's authoritative records | Preserve global enablement, assignment enablement, revision selection, role adoption and time validity. |
| Parent grant and actual supporting assignments | `parent_grant_id` plus Auth's real relationship records | Establish usable support, not just existence of a grant definition. |
| Team parentage and human membership | Auth's relationship records | Establish the parent-team ceiling and Maya's source access separately; ownership is not membership. |
| Registered permissions and scope definitions | Applicable accepted application catalog | Validate structure and supported meanings without inventing application facts. |
| Affected bindings and shared uses | Reverse lookup of assignments and relevant parent relationships | Check all affected branches when a structural change or grant-wide operation can affect them. |
| Relevant state through persistence | Auth's consistent read/write mechanism | Ensure the persisted result is still supported by the facts that justified it. |

An input path or SQL table is not prescribed. Caller-supplied IDs locate records;
they do not certify their current state. A failed required lookup is not proof
that a record is absent. Sufficient already-available valid evidence can be used;
this chapter does not mandate a remote fetch per check.

## 3. Support discovery without another parent field

Keep `parent_grant_id`. Do not add `parent_assignment_id` or
`parent_grant_revision` to solve discovery by silently changing the model.

For **G2 assigned to Team2, a child of Team1**:

1. Read G2's selected immutable revision and its parent G1.
2. Establish Team2's actual parent Team1; do not infer it from a name or scope.
3. Find the current G1 assignment to Team1 within the trusted tenant. Q-104
   permits at most one current assignment for that grant/recipient, including
   disabled records. An assignment at TeamX is not this supporting assignment.
4. Read the revision actually adopted by Team1's assignment. Validate that
   assignment, G1's control, its content, and its required upstream support.
5. Repeat toward the legitimate root, then resolve effective authority top-down.
   Select relevant adopted revisions, not an artificial union of every historical
   parent edge. A cycle cannot supply its own authority.
6. Check G2 against that complete effective parent route. Preserve every inherited
   restriction and applicable Team1 ceiling.

If the required Team1 holding is conclusively absent or unusable, this route
cannot supply authority. An incomplete or inconsistent lookup cannot be treated
as proof of support. Do not switch to TeamX merely because its revision is newer.

For **Maya's source access at issuance**, independently establish a valid direct
assignment or valid human membership in a group holding the required source.
This is checked when she acts. It does not make her issuance-time membership a
new permanent dependency of Team2's team-held route.

For **direct-human recipients**, Q-112A still requires actual valid adopted
lineage. Neither the newest published revision, a union of unrelated holdings,
nor the issuer's remembered assignment may substitute for that lineage. The
unresolved discovery case is stated precisely in section 9; this draft does
not invent its answer or prohibit direct assignments.

Sources: [Q-099](ownership-lineage.md), [Q-101](parent-grant-bindings.md),
[Q-103/104](grant-revisions.md), [Q-111](lineage-cycles.md),
[Q-112A](direct-human-parent-context.md).

## 4. Containment is constructed, not guessed

For a fixed valid source context and the same operation's data:

```text
Child selected permissions ⊆ effective parent permissions
Effective child boundary = effective parent boundary AND additional child scope
Effective child lifetime is also limited by every required upstream restriction
```

Equality is allowed. Additional `{}` retains all inherited constraints. These
rules apply recursively, so a narrow grandchild cannot repair an unsupported
parent. Do not accept an oversized submitted permission set by silently trimming
it, and do not assemble one child grant from unrelated permission/scope fragments.

Scope accumulation is a conjunction of predicates, **not a JSON object merge**.
If the parent requires FIN and the child adds ENG under the same exact-department
meaning, both restrictions remain. The child cannot replace FIN with ENG. Whether
every unsatisfiable definition must be rejected at publication is not settled;
such a route cannot authorize either department by dropping a constraint.

Auth validates registered syntax and canonical lineage construction. It does
not query a customer HRMS database to prove that C17 belongs to FIN. The HRMS
endpoint must constrain the actual read/write to FIN AND C17. Names alone do
not establish the relationship.

`$self` requires care across recipients. An Employees grant with `user: "$self"`
means the requesting human, not Employees as a group. That fact does not prove
that Maya's personal self authority can be reassigned as Nutan's self authority.
Identical text with a different human binding is not a containment proof. The
validator must not accept a proposed cross-context change by string equality
when the required binding cannot be established.

Sources: [scope](scope-model.md), [Q-095](authority-lineage.md),
[Q-100](auth-service-authority-gate.md), [endpoint enforcement](endpoint-policy-format.md).

## 5. Worked records: Maya assigns G2 to Team2

These use the **approved core record layouts**, not a new assignment-operation
request envelope or complete bootstrap example. All shown records have version
`"1"`. Tenant is implied and identical throughout; it is not a scope entry.

Established premises, which an implementation must verify rather than assume:

- Permissions and the illustrated `group`, `dept`, and `cert` scope meanings
  are registered by their respective applications and supported where used.
- Team2 is a child of Team1. Maya is a member of Team1 and AssignmentAdmins.
  Nutan is a member of Team2. These are memberships, not ownership records.
- G0 has valid upstream/root support for Team1's HRMS authority. G-AUTH-ROOT
  has valid support for the administrative grant. Their trusted establishment
  is a prerequisite, not an ordinary caller-controlled parent omission.
- The shown grant controls are enabled, required upstream support is current,
  and the shown revisions are latest at the relevant creation/adoption times.
  There are no additional local time windows in these example revisions.

### Administrative authority held through AssignmentAdmins

```json
{
  "version": "1",
  "grant_id": "G-ASSIGN-TEAM2",
  "revision": 1,
  "parent_grant_id": "G-AUTH-ROOT",
  "permissions": ["auth:assignment::create"],
  "scope": {"group": "Team2"}
}
```

```json
{
  "version": "1",
  "id": "A-ADMIN",
  "grant_id": "G-ASSIGN-TEAM2",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "AssignmentAdmins"},
  "status": "enabled"
}
```

`group: Team2` bounds assignment administration to that recipient. It is not
an HRMS department, membership declaration or ownership grant.

### Parent business authority held by Team1

```json
{
  "version": "1",
  "grant_id": "G1",
  "revision": 1,
  "parent_grant_id": "G0",
  "permissions": ["hrms:payroll:payslip::read", "hrms:payroll:payslip::write"],
  "scope": {"dept": "FIN"}
}
```

```json
{
  "version": "1",
  "id": "A1",
  "grant_id": "G1",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "Team1"},
  "status": "enabled"
}
```

### Child content and proposed assignment

```json
{
  "version": "1",
  "id": "G2",
  "status": "enabled"
}
```

```json
{
  "version": "1",
  "grant_id": "G2",
  "revision": 1,
  "parent_grant_id": "G1",
  "permissions": ["hrms:payroll:payslip::read"],
  "scope": {"cert": "C17"}
}
```

```json
{
  "version": "1",
  "id": "A2",
  "grant_id": "G2",
  "grant_revision": 1,
  "recipient": {"type": "group", "id": "Team2"},
  "status": "enabled"
}
```

The last block is the proposed resulting assignment, not evidence that A2 has
already been written. Publishing G2 alone gives Team2 and Nutan no access.

```text
Maya → AssignmentAdmins → A-ADMIN → may create assignment to Team2
Maya → Team1 → A1 → may supply G1's current FIN read/write authority
                                     │
Team1 ───────── parent team ───────── Team2
  A1                                  proposed A2
  G1 ───────── parent grant ───────── G2
  FIN read/write                      FIN AND C17 read
                                       │
                                Nutan's membership
                                       │
                     HRMS endpoint evaluates AND enforces actual data
```

The administration check passes for Team2. The source check establishes G1 via
Maya's Team1 membership, and the continuing child-team check establishes A1 at
Team1. Read is a subset of read/write; C17 is additional narrowing. If all other
checks hold through persistence, A2 may be created. Later Nutan can use this
route for reads constrained to FIN AND C17—not write, all FIN, or C17 in ENG.

Changing the proposed recipient to Team3 does not fit the shown administrative
grant. Changing the proposed permission to delete does not fit G1. These fail
different checks. No owner grant is needed to explain either result.

## 6. Operation coverage

All protected Auth operations retain their endpoint-declared administrative
check. The second column specifies additional obligations already implied by
the operation; it is not a new permission-name or API catalogue.

| Operation | Applicable authority/binding checks | Must not happen implicitly |
|---|---|---|
| Publish ordinary child content | Validate the proposed authority against permitted source bounds, registration and relevant lineage; content is immutable once published. | Treat publication as assignment or adoption. |
| Create assignment | Recipient administration, current source access, complete recipient route and team ceiling, latest published content, uniqueness and consistency through write. | Fall back to an older revision, or replace a disabled current assignment with a duplicate. |
| Explicitly upgrade assignment | Validate the whole resulting authority against current support; select latest and preserve checked state through write. | Select an intermediate revision or auto-upgrade parent support. |
| Re-enable assignment | Validate current bindings, adopted revision, support, validity and administrator authority. | Upgrade its adopted revision or restore other explicitly disabled records. |
| Re-enable shared grant | Apply the grant-wide enablement checks to required affected bindings and current authority. | Enable the same grant only for recipients whose checks happen to pass. |
| Disable grant | Authorized grant-wide withdrawal; descendants may become ineffective without changing their stored flags. | Treat derived ineffectiveness as explicit structural unbinding. |
| Disable/remove an assignment supporting descendants | Apply Q-101's bottom-up guards to all affected bindings. | Break an enabled dependent binding first and rely on a later silent cascade. |
| Change a bound team/grant parent | Handle affected bindings by explicit disablement/removal, bottom-up; validate resulting parentage, complete authority and acyclicity. | Bypass the guard because the old grant is temporarily ineffective, or re-enable retained assignments automatically. |
| Delete grant | Authorized permanent withdrawal plus applicable structural guards; required unsupported descendants cannot authorize. | Re-enable a deleted grant or automatically choose replacement support. |
| Add a human member | Q-092 team-write authority within its administrative boundary; distribute the team's existing valid authority through membership. | Require a new business-permission-possession check not approved for membership writes, or let membership alone administer the team. |
| Change ownership | Explicit authorized administration; preserve unchanged team-held support and actual dependencies under Q-099. | Import the new owner's personal permissions or infer the owner-management permission from this table. |
| Establish root / publish app catalog | Apply the separate trusted-bootstrap/platform-publication rules. | Turn ordinary parent omission into root authority or catalog administration into tenant business access. |

Sources: [Q-092](team-administration.md), [Q-099](ownership-lineage.md),
[Q-101](parent-grant-bindings.md), [Q-102–Q-106](grant-revisions.md),
[Q-110](auth-write-consistency.md), [bootstrap](bootstrap-authority.md),
[Q-121–Q-123](root-permission-evolution.md).

## 7. Completion, failure and evidence

Passing administration alone is not final authorization to persist. Successful
boundary validation is tied to the exact proposed change and the relevant state
used to establish it. A conflicting change before persistence stops that attempt
without a partial assignment update or silent substitution. An explicit fresh
attempt starts with current authorization; no reusable validation token is
introduced. Q-129's prior-allow application completion does not relax Q-110 for
Auth's authority-changing writes.

| Established result | Required consequence |
|---|---|
| Both gates pass and relevant state is preserved through persistence | The exact validated write may complete. This is not permission for a later unrelated write. |
| Complete evidence proves administrative authority or required source support insufficient | Reject the protected change with a reason; no protected write. |
| A required lookup or freshness proof cannot be completed | Evaluation error supplies no authority; do not label a timeout as proof of orphanhood. |
| Malformed input, unsupported structure, duplicate current assignment or forbidden structural edit | Reject under the relevant validation rule; never repair by dropping restrictions. Exact transport/error codes remain separate. |
| Relevant state changes before write | Stop the attempt under Q-110; no stale approval or automatic revision substitution. |

Evidence needed to establish this result consists of the actual administrative
route, affected assignment IDs and adopted content, required support chain,
team/membership/delegation context where applicable, preserved scope bindings,
applicable controls/validity/catalog, and the checked proposal. These are logical
proof dependencies, not an approved new public envelope or an audit-log design.
Keep each route's permission and restrictions together. Assignment IDs can occur
in evidence without becoming a second parent-reference field on grant content.

Sources: [decision results](decision-results.md), [Q-110](auth-write-consistency.md),
[Q-128](authority-freshness.md), [Q-129](concurrent-enforcement.md).

## 8. Review cases and rationale

These are documentary expected outcomes derived from existing approval, **not
executed Auth-engine conformance tests**. Each variation starts from the worked
example unless stated otherwise. Rejected routes do not invalidate unrelated
complete valid routes.

| Case | Variation | Expected consequence and rationale | Source |
|---|---|---|---|
| C01-T01 | Baseline FIN/C17 read assignment | May persist when all premises and write consistency hold; both gates are satisfied. | Q-093/100/110 |
| C01-T02 | Maya has G1 but no assignment administration | No write; possession is not administration. | Q-093 |
| C01-T03 | Maya has administration but no valid access to G1 | No write; administration does not supply the source. | Q-093 |
| C01-T04 | Proposed recipient becomes Team3 | Shown admin route cannot authorize it; recipient boundary is enforced. | Q-093/100 |
| C01-T05 | G2 selects delete absent from G1 | Reject unsupported authority; do not trim the submitted grant to make it pass. | Q-095/100 |
| C01-T06 | G2 additional scope becomes `{}` | Retains FIN; no tenant-wide escape. | Q-091/101 |
| C01-T07 | G2 adds ENG under exact-department meaning | FIN AND ENG, never ENG by overwrite; cannot authorize by dropping FIN. Publication rejection policy is not decided here. | Q-100/101 |
| C01-T08 | G1 exists only as a stored definition | Not usable support; definitions alone grant no access. | Q-090/101 |
| C01-T09 | Required G1 support absent at Team1 but present at TeamX | TeamX cannot substitute for Team2's parent-team holding. | Q-101/103 |
| C01-T10 | Team1 holds revision 1, TeamX holds broader revision 2 | Team2 uses actual Team1-adopted support, not the broader unrelated holding. | Q-103 |
| C01-T11 | New assignment selects older revision although newer is published | Reject creation; no old-revision fallback. | Q-104A/105 |
| C01-T12 | Another A2 already exists disabled for G2/Team2 | Reject duplicate; disabled assignments still count as current. | Q-104 |
| C01-T13 | G2 disabled while A2 enabled | No authority from G2; assignment does not override global control. | Q-101A/B |
| C01-T14 | A2 disabled while G2 enabled | No authority through A2; other assignments are not globally disabled. | Q-101A/B |
| C01-T15 | G2 disabled, enabled G3 depends on it; G2 later validly enabled | G3 may regain effectiveness if all requirements hold, without a G3 state rewrite. | Q-101C |
| C01-T16 | G3 itself explicitly disabled before G2 restoration | G3 remains disabled; returning support is not enablement. | Q-101C |
| C01-T17 | Required parent change attempted beneath an enabled affected binding | Reject structural change; first handle relevant bindings bottom-up. | Q-101E |
| C01-T18 | All relevant bindings disabled/removed before valid reparenting | Authorized change may proceed; retained disabled assignments need explicit current-state enablement. | Q-101E-3 |
| C01-T19 | Another affected shared branch remains enabled | Its guard still applies; inspecting only one branch is insufficient. | Q-101E |
| C01-T20 | Proposed parent introduces a cycle while assignments disabled | Reject cycle; disablement is not permission for invalid lineage. | Q-111 |
| C01-T21 | Required parent is in another tenant | Cannot establish ordinary tenant-bounded support. | Q-093/101 |
| C01-T22 | Parent support or required adopted validity expires | Route cannot authorize; enabled flags do not extend validity. | Q-083/109 |
| C01-T23 | Required parent lookup times out | No write from incomplete proof; error, not established orphanhood. | Q-051/101 |
| C01-T24 | New latest revision or withdrawal invalidates checked state before persistence | Stop original write; fresh attempt must validate current selected content. | Q-110/128 |
| C01-T25 | Maya ceases being owner/member after valid team-held issuance, support otherwise unchanged | Do not invent a permanent issuer dependency; her future administration/source access is checked anew. Explicit dependencies still matter. | Q-099/101 |
| C01-T26 | Last required support disappears despite stored G1/G2/A2 | Affected orphan lineage cannot authorize; no silent replacement or cleanup. | Q-094/101D |
| C01-T27 | Nutan requests C17 but actual data belongs to ENG | HRMS must not return/mutate it under FIN authority; Auth issuance does not establish domain facts. | Q-050-C/D |
| C01-T28 | Maya's `$self` is copied to a different recipient and string equality is offered as proof | Not a containment proof; changing the human binding cannot be ignored. | SELF-001/Q-100 |

## 9. What this finishes, and what it deliberately does not invent

**Written and reviewable now:** the two-gate behavioral contract, input ownership,
deterministic parent-team support lookup, narrowing semantics, complete worked
core records, operation obligations, no-write failure rules, logical evidence
requirements and 28 sourced review cases. These are consequences of existing
approval; they do not require 28 new questions.

**Remaining within C01:**

The [C01-D2 record-level trace](direct-human-parent-context.md) now distinguishes
finding candidate parent holdings from proving their eligibility for Nutan's
direct route. The original same-content Q-101 case and later differing-revision
case are kept separate. This completes the diagnostic, not the missing selection
policy or interface representation.

1. **Direct-human support discovery:** in Q-112A's example G1 is held as FIN at
   Team1 and ENG at TeamX, while G2 is directly assigned to Nutan. The known
   records do not establish which is Nutan's actual lineage. Complete the
   discovery contract without a union, implicit issuer dependency or withdrawn
   parent field. The approved selection principle is not reopened.
2. **Recipient-relative source binding:** establish how a source's human-relative
   constraints are preserved when assigning to another human/group. Group-self
   evaluation and direct-human proxy ceilings are agreed; copying `$self` text
   is not enough to complete this issuance contract.
3. **Exact interface/evidence representation:** assemble versioned operation and
   validation-result contracts with C02–C05. No public fields, error codes, owner
   permission, hierarchy payload, database isolation level or root encoding are
   invented in this behavioral consolidation.

These gaps already existed in Q-100/Q-112A and the C01 inventory; they are not
new scope or a reason to withhold the supported subteam deliverable. HC-05-08
remains open until its full containment and representation requirements are met.
Auth runtime implementation and external audit design are not handbook closure
requirements. The existing root/catalog contract work remains in C02/C04.
