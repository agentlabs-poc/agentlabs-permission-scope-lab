# Contract publication and versioning — CONTRACT-009

## Shared version convention — CONTRACT-010 / Q-050-A, agreed

Every published JSON/YAML contract has a required top-level `version` field.
Its value is a string, initially `"1"`; quote the value in YAML as well.
This is approved metadata, not a complete grant or endpoint policy:

```json
{
  "version": "1"
}
```

The value identifies the contract format and meaning, not an individual
document's edit revision. Updating a grant's scope within the same contract
does not by itself change its contract version. The version is interpreted
within its contract type: grant `"1"` and endpoint policy `"1"` refer to their
respective definitions and need not evolve together. Version alone does not
identify a contract type; establishing that type remains part of the surrounding
contract/integration design.

### Rationale and alternative

A top-level field identifies the applicable definition before interpreting its
contents. A string keeps JSON/YAML representation consistent and avoids treating
the identifier as an arithmetic value. `schema_version` was considered as a more
explicit name; the simpler `version` was selected with its meaning defined once.
Separating contract version from document revision avoids a new contract version
for each ordinary grant edit. Independent contract-type versions avoid forcing
unrelated grant-format changes when endpoint policy formats evolve.

### Rejection rules and counterexamples

Consumers reject missing, malformed, or unsupported versions. They must not
guess a default, silently downgrade, or interpret an unknown version as current.

| Version material | Consequence |
|---|---|
| Top-level `"version": "1"`, supported for the expected contract type | Version check passes; all other contract and authorization checks remain. |
| Version omitted | Reject; no implicit `"1"`. |
| Numeric `"version": 1` or `"version": null` | Reject; the required string representation is absent. |
| A version the consumer does not support | Reject; do not reinterpret it as `"1"`. |
| Version supplied only inside `scope` | Does not meet the top-level requirement; version metadata is not a scope boundary key. |

Scope `{}` retains its tenant-wide meaning. This convention does not finalize
any complete contract. Compatibility, migration rules, future version numbering,
full endpoint policy/grant schemas, and document-revision representation remain
open. Older statements below that version field/value/placement or unsupported
version behavior are undecided are preserved history, superseded specifically
by CONTRACT-010.

## Agreed publication requirement

Every published JSON/YAML contract must include a version. This applies to
grant contracts, endpoint policy contracts, and other contracts when published.
The user explicitly required this while approving the single-permission rule.
Publication must not present an unversioned working illustration as a finalized
contract.

## Rationale

A consumer needs to know which contract definition a document follows. Without
an explicit version, later changes to fields or meanings can leave producers
and consumers interpreting the same document differently. This is particularly
important when interpretation affects authorization boundaries or enforcement.
An explicit version makes that distinction available; versioning alone does not
establish compatibility, validate a payload, or authorize the operation.

## Current examples versus published contracts

The handbook's existing grant JSON is a working layout, not a complete published
contract. Scope fragments such as `{}` and `{"dept":"FIN"}` illustrate agreed
boundary semantics; they are not complete versioned grant or endpoint-policy
documents. Existing examples are retained, not silently rewritten with invented
version syntax. Each eventual full published contract must include its version.

For example, a published grant document must identify its contract version as
well as express the grant binding. A published endpoint policy document must
identify its contract version and encode the agreed endpoint requirements.
These are requirements, not adopted JSON/YAML examples. No field name, value,
placement, wrapper, or numbering convention is selected here, and no version
field is added to the scope-boundary key catalog by this requirement.

## Endpoint policy contract remains open

CONTRACT-007/008 settle what the endpoint predeclares: one required permission,
inputs, sources, and how to establish any required relationship. They do not
settle the endpoint policy's JSON/YAML contract. The user explicitly reminded
us that this contract has not yet been discussed.

Q-050 opens that discussion. Contract structure, version representation, binding
syntax, validation timing, and registration/integration details remain open.
Compatibility, unsupported-version handling, migration rules, and the distinction
between contract version and individual document revision also need decisions.
None is implicitly approved by requiring an explicit version.
