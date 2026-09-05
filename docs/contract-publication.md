# Contract publication and versioning — CONTRACT-009

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
