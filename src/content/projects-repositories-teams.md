# Projects and repositories: the same boundary model

Post-`0.0.1`: [Q-090](../../docs/grant-assignments.md) separates recipient-free
grant definitions from assignments. Recipient-bearing examples below are
deprecated layouts, preserved for their boundary/enforcement explanations.
[Subgroups](../../docs/subgroups.md) now support explicit dependent authority,
not inherited membership or application project hierarchy.

This example reuses the agreed authorization model in a fictional Git-hosting
application. It does not introduce built-in project types, automatic hierarchy,
or a second grant format. The
[original projects guide](../../docs/history/reconciliation-2026-09-05/src/content/projects-repositories-teams.md.txt)
remains available as deprecated history.

## Register meanings, not universal hierarchy assumptions

Assume this application registers `git:repository::read` and `project`.
For that operation it defines the project boundary using the repository's
containing project. Auth validates registration and grant acceptance; the
application establishes or enforces the actual relationship at execution.

```json
{
  "version": "1",
  "recipient": { "type": "group", "id": "platform-engineers" },
  "permissions": ["git:repository::read"],
  "scope": { "project": "P-1" }
}
```

This abbreviated working grant permits human members to read repositories
within P-1 under this explicitly defined meaning. It does not authorize writes,
other projects, or another tenant. Lifecycle and dependency restrictions still
apply even though their detailed fields are omitted here.

**Rationale:** project-based reach is meaningful because this application defines
it, not because Auth guesses a folder tree. Exact/subtree permission and scope
mechanics are not silently standardized by this example.

## One endpoint policy, one permission

```json
{
  "version": "1",
  "method": "GET",
  "path": "/api/v1/{tenant}/projects/{project}/repositories/{repo}",
  "permission": "git:repository::read",
  "inputs": {
    "tenant": { "source": "path", "name": "tenant" },
    "project": { "source": "path", "name": "project" },
    "repo": { "source": "path", "name": "repo" }
  }
}
```

This uses the approved partial policy structure, not a completed production
schema. The selected inputs are required at their exact sources. Their types and
domain validity belong to the application's request contract. Path names do not
automatically become trusted boundary facts.

For a request identifying P-1 and R-7, the one endpoint-owned gate must bind
actual execution to the trusted tenant, authorized project, and requested
repository. R-7 in P-2 must not be returned merely because the caller wrote P-1
in the path. A constrained query can enforce the containment relationship
without a canonical relationship block or mandatory resolver call.

## Teams are human authorization groups

Team and group are synonyms. Auth owns authorization membership; the application's
collaboration team may be synchronized to that group or kept separate. Neither
being a repository contributor nor creating a project automatically creates
authorization membership.

An agent acting for Vinay remains dependent on Vinay and the delegated subset
of his authority. It cannot become a first-class team member. Removing a required
membership removes that route of derived access, not necessarily every other
valid grant.

## Keep complete grants intact

| Authority | What it does not imply |
|---|---|
| Read within P-1 | Write within P-1. |
| Read tenant-wide via `{}`, plus write within P-1 | Write tenant-wide. Permission and scope cannot be mixed across grants. |
| Two read grants for P-1 and P-2 | One new independent grant; they remain alternative complete bindings. |
| Permission to administer grants | Repository access without an explicit authorized grant. |

A multi-key scope combines with AND only when each key's meaning is supported.
A repo-only scope could be defined by an application, but it is not made a
mandatory platform key by this guide. No wildcard or arbitrary selector
expression is adopted.

## Boundaries of this example

Repository moves, source/destination authorization, list/count/export behavior,
bulk partial success, relationship freshness, and concurrent changes still need
their dedicated contracts. The read example does not settle them.

See [the complete use-case chapter](../../docs/use-case-examples.md),
[registration](../../docs/application-registration.md),
[scope](../../docs/scope-model.md), and
[endpoint policy rationale](../../docs/endpoint-policy-format.md).
The deprecated projects explorer has been removed from the active site. Its
[archived source](../../docs/history/retired-pages-2026-09-05/README.md) remains
available for historical comparison, not as a conformance test.
