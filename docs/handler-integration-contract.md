# Handler integration contract — Q-146 / CONTRACT-015

Status: **AGREED.** This closes HC-07-10, which asked for the handler and
embedded-agent integration contract and the failures at its boundaries. It is
adopted from a reference implementation that an application links and runs, with
the split it claims enforced by a test rather than asserted.

An application integrates authorization through **one gate in front of one
endpoint**. The gate owns the order of operations; the application supplies four
things and receives one.

## What the application supplies

| | What it is | What it must not do |
|---|---|---|
| **Policy** | The endpoint's static declaration — version, method, path, one permission, selected inputs, trusted correlations. | Change after mounting. A mounted copy decides; the caller's struct is no longer consulted. |
| **Identity source** | Establishes the trusted request context from the request. | Read the business body. It is handed a request whose body is empty. |
| **Binder** | Validates application schema, selects the material authorization will use, and returns the effect as a closure over those same validated values. | Publish output, or perform the effect. It runs before any decision exists. |
| **Failure handler** | Renders a denial or an evaluation failure. | Turn one into the other, or proceed. |

And it receives the **Result** — the decision and, on an allow, the contributing
chain of [Q-134](decision-results.md).

## The order, and why each step is where it is

1. **Method and route** must match the policy, or nothing else runs.
2. **Identity** is established, from a request with no business body — so a
   credential can never be taken from the payload it is meant to protect.
3. **Trusted correlations** are verified against the resolved inputs — Q-133.
4. **The binder** produces material and the effect closure.
5. **Authority is loaded and evaluated** — [Q-139](authority-resolve-transport.md).
6. **Only on an allow** does the effect run, and it is handed the Result.

The binder runs before evaluation deliberately: the effect must close over the
*same* validated values the decision was made about. A binder that ran afterwards
could bind to something else, which is the check-to-use gap ENFORCEMENT-008 exists
to prevent.

## Failures at the boundary

Every failure is one of two things, and the distinction is the contract —
[Q-051](decision-results.md):

- **A completed denial.** Evidence established that no complete route authorizes.
- **A failure to establish authority.** Anything else: the identity source
  refused, the binder rejected the request, the answer did not arrive or could not
  be used, the context was cancelled, or an unusable route left a denial
  unestablished ([Q-137](decision-results.md)).

The effect never runs for either. A denial is never reported as a failure, and a
failure is never reported as a denial.

**No HTTP request reaches the effect.** It receives a context, a writer and the
Result, so it cannot reparse a path or a body after the decision.

## What the application must not link

An application linking this gate links **no authority records, no schema and no
authority store**. It needs the gate and a source of answers, and nothing else.

That is a structural claim, so it is held structurally: the reference
implementation fails its build if the authority domain enters the application's
dependency closure, and also fails if the gate and its client are absent — so the
check cannot pass by proving nothing. An earlier version of that check was
defeated by two lines in a module file, which is why it is a compiled dependency
assertion rather than a convention.

## What is deliberately not specified

**An SDK.** The criterion is explicit that handbook completion does not require
one. What is specified is the contract: the four things supplied, the order, and
the two failure kinds.

**HTTP status mapping.** Which status renders a denial, and which a failure,
belongs to the application. The reference implementation answers 403 and 503, and
records that as a local convention rather than a rule.

**Transport for administration.** One endpoint is published
([Q-139](authority-resolve-transport.md)); administrative operations have no wire
contract and that stays open.

## Rationale / conscious tradeoff

The alternative was to leave integration to each deployment and specify only the
wire. It was rejected because the ordering *is* the authorization: establish
identity before reading the body, bind before deciding, and run the effect only
after. A deployment that reorders those has a gate that looks right and is not,
and nothing in the wire contract would catch it.

The cost is that this constrains application structure more than a pure wire
contract would. An application with its own middleware chain must place this gate
where the order above still holds, and the contract cannot check that it did.
