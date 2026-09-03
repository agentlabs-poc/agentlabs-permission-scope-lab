import './styles.css'

interface ProjectScenario {
  id: number
  title: string
  brief: string
  principal: { id: string; name: string; tenant: string }
  membership: { team: string; members: string[] }
  role: { name: string; permissions: string[]; assignedBy: string }
  permission: string
  scope: { type: 'resource_exact' | 'resource_subtree'; resourceType: string; resourceId: string }
  target: { id: string; name: string; tenant: string; organization: string; project: string; type: string }
  expected: boolean
}

const scenarios: ProjectScenario[] = [
  {
    id: 1,
    title: 'Backend team reads its API repository',
    brief: 'Vinay reads REPO-API through membership in TEAM-BACKEND.',
    principal: { id: 'user:vinay', name: 'Vinay', tenant: 'TENANT-001' },
    membership: { team: 'TEAM-BACKEND', members: ['user:vinay', 'user:arjun'] },
    role: { name: 'Repository Maintainer', permissions: ['codehost:repository::read', 'codehost:repository::write'], assignedBy: 'user:tenant-admin' },
    permission: 'codehost:repository::read',
    scope: { type: 'resource_exact', resourceType: 'codehost:repository', resourceId: 'REPO-API' },
    target: { id: 'REPO-API', name: 'Commerce API', tenant: 'TENANT-001', organization: 'ORG-ACME', project: 'PROJECT-COMMERCE', type: 'codehost:repository' },
    expected: true,
  },
  {
    id: 2,
    title: 'Exact repository scope stops at API',
    brief: 'Vinay tries to read REPO-UI using an assignment limited to REPO-API.',
    principal: { id: 'user:vinay', name: 'Vinay', tenant: 'TENANT-001' },
    membership: { team: 'TEAM-BACKEND', members: ['user:vinay', 'user:arjun'] },
    role: { name: 'Repository Maintainer', permissions: ['codehost:repository::read', 'codehost:repository::write'], assignedBy: 'user:tenant-admin' },
    permission: 'codehost:repository::read',
    scope: { type: 'resource_exact', resourceType: 'codehost:repository', resourceId: 'REPO-API' },
    target: { id: 'REPO-UI', name: 'Commerce UI', tenant: 'TENANT-001', organization: 'ORG-ACME', project: 'PROJECT-COMMERCE', type: 'codehost:repository' },
    expected: false,
  },
  {
    id: 3,
    title: 'Project subtree reaches contained repositories',
    brief: 'Maya reads REPO-UI through the Auditors team assignment at PROJECT-COMMERCE.',
    principal: { id: 'user:maya', name: 'Maya', tenant: 'TENANT-001' },
    membership: { team: 'TEAM-AUDIT', members: ['user:maya'] },
    role: { name: 'Project Viewer', permissions: ['codehost:project::read', 'codehost:repository::read'], assignedBy: 'user:tenant-admin' },
    permission: 'codehost:repository::read',
    scope: { type: 'resource_subtree', resourceType: 'codehost:project', resourceId: 'PROJECT-COMMERCE' },
    target: { id: 'REPO-UI', name: 'Commerce UI', tenant: 'TENANT-001', organization: 'ORG-ACME', project: 'PROJECT-COMMERCE', type: 'codehost:repository' },
    expected: true,
  },
  {
    id: 4,
    title: 'Project subtree excludes another project',
    brief: 'Maya tries to read REPO-PAYROLL outside PROJECT-COMMERCE.',
    principal: { id: 'user:maya', name: 'Maya', tenant: 'TENANT-001' },
    membership: { team: 'TEAM-AUDIT', members: ['user:maya'] },
    role: { name: 'Project Viewer', permissions: ['codehost:project::read', 'codehost:repository::read'], assignedBy: 'user:tenant-admin' },
    permission: 'codehost:repository::read',
    scope: { type: 'resource_subtree', resourceType: 'codehost:project', resourceId: 'PROJECT-COMMERCE' },
    target: { id: 'REPO-PAYROLL', name: 'Internal Payroll', tenant: 'TENANT-001', organization: 'ORG-ACME', project: 'PROJECT-INTERNAL', type: 'codehost:repository' },
    expected: false,
  },
  {
    id: 5,
    title: 'Tenant boundary remains absolute',
    brief: 'Maya targets a repository in another tenant even though its project name looks familiar.',
    principal: { id: 'user:maya', name: 'Maya', tenant: 'TENANT-001' },
    membership: { team: 'TEAM-AUDIT', members: ['user:maya'] },
    role: { name: 'Project Viewer', permissions: ['codehost:project::read', 'codehost:repository::read'], assignedBy: 'user:tenant-admin' },
    permission: 'codehost:repository::read',
    scope: { type: 'resource_subtree', resourceType: 'codehost:project', resourceId: 'PROJECT-COMMERCE' },
    target: { id: 'REPO-EXTERNAL', name: 'External Commerce API', tenant: 'TENANT-002', organization: 'ORG-OTHER', project: 'PROJECT-COMMERCE', type: 'codehost:repository' },
    expected: false,
  },
]

let scenarioId = 1
let stepIndex = 0
const explored = new Set<number>()
const app = document.querySelector<HTMLDivElement>('#projects-app')!

function shield() {
  return '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/></svg>'
}

function wireJsonBlock(title: string, payload: unknown) {
  return `<details class="wire-json"><summary>View wire JSON · ${title}</summary><pre>${escapeHtml(JSON.stringify(payload, null, 2))}</pre></details>`
}

function escapeHtml(value: string) {
  return value.replace(/[&<>'"]/g, character => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' })[character]!)
}

// What the code-hosting API receives from the caller. The caller can name a
// target repository, but cannot assert team membership, scope, or tenant.
function buildClientRequest(scenario: ProjectScenario) {
  const [resource, verb] = scenario.permission.split('::')
  return {
    actor_token: scenario.principal.id,
    operation: `${verb.toUpperCase()} /${resource.replace(/:/g, '/')}/${scenario.target.id}`,
    requested_permission: scenario.permission,
    requested_target_id: scenario.target.id,
  }
}

// What Auth returns for this principal: team membership and the grant that
// membership carries. No membership, no inherited grant.
function buildGrantSet(scenario: ProjectScenario, membershipMatches: boolean) {
  return {
    principal_id: scenario.principal.id,
    team_memberships: membershipMatches ? [scenario.membership.team] : [],
    grants: membershipMatches ? [{
      via_group: scenario.membership.team,
      role: scenario.role.name,
      permissions: scenario.role.permissions,
      scope: { type: scenario.scope.type, resource_type: scenario.scope.resourceType, resource_id: scenario.scope.resourceId },
      assigned_by: scenario.role.assignedBy,
    }] : [],
  }
}

// What the code-hosting Authorization Agent hands to the decision step, after
// resolving trusted target attributes — never taken from the caller.
function buildResolvedRequest(scenario: ProjectScenario) {
  return {
    principal: { id: scenario.principal.id, tenant_id: scenario.principal.tenant },
    permission: scenario.permission,
    resolved_context: { via_group: scenario.membership.team },
    target: {
      type: scenario.target.type,
      id: scenario.target.id,
      tenant_id: scenario.target.tenant,
      project_id: scenario.target.project,
      organization_id: scenario.target.organization,
    },
  }
}

function buildDecisionResult(scenario: ProjectScenario, allowed: boolean, membershipMatches: boolean, capabilityMatches: boolean, tenantMatches: boolean, containmentMatches: boolean, predicate: string) {
  return {
    decision: allowed ? 'allow' : 'deny',
    checks: {
      membership: membershipMatches ? 'pass' : 'fail',
      capability: capabilityMatches ? 'pass' : 'fail',
      tenant_boundary: tenantMatches ? 'pass' : 'fail',
      scope_reach: containmentMatches ? 'pass' : 'fail',
    },
    enforced_predicate: allowed ? predicate : 'FALSE  /* no repository rows */',
    audit: { principal: scenario.principal.id, via_group: scenario.membership.team, permission: scenario.permission, target: scenario.target.id, decision: allowed ? 'allow' : 'deny' },
  }
}

function scopeContains(scenario: ProjectScenario) {
  if (scenario.scope.type === 'resource_exact') return scenario.target.type === scenario.scope.resourceType && scenario.target.id === scenario.scope.resourceId
  return scenario.scope.resourceType === 'codehost:project' && scenario.target.project === scenario.scope.resourceId
}

function selectScenario(id: number) {
  scenarioId = id
  stepIndex = 0
  render()
  document.querySelector('.guided-walkthrough')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function render() {
  const scenario = scenarios.find(item => item.id === scenarioId)!
  const membershipMatches = scenario.membership.members.includes(scenario.principal.id)
  const capabilityMatches = scenario.role.permissions.includes(scenario.permission)
  const tenantMatches = scenario.principal.tenant === scenario.target.tenant
  const containmentMatches = scopeContains(scenario)
  const allowed = membershipMatches && capabilityMatches && tenantMatches && containmentMatches
  const permissionParts = scenario.permission.split('::')
  const resourceParts = permissionParts[0].split(':')
  const predicate = scenario.scope.type === 'resource_exact'
    ? `tenant_id = ${scenario.principal.tenant}\nAND repository_id = ${scenario.scope.resourceId}`
    : `tenant_id = ${scenario.principal.tenant}\nAND project_id = ${scenario.scope.resourceId}`
  const steps = [
    { eyebrow: 'REQUEST STORY', title: 'Start with the person and their request', body: `<p>${scenario.brief}</p><div class="guide-facts"><span><small>PERSON</small><b>${scenario.principal.name}</b><code>${scenario.principal.id}</code></span><span><small>OPERATION</small><b>Read repository</b><code>${scenario.permission}</code></span><span><small>TARGET</small><b>${scenario.target.name}</b><code>${scenario.target.id}</code></span></div><aside><b>Concept:</b> The user is the authenticated principal. The repository is the target resource. Their relationship is not assumed.</aside>${wireJsonBlock('client request → code-hosting API', buildClientRequest(scenario))}` },
    { eyebrow: 'GROUP PRINCIPAL', title: 'Expand team membership before evaluating the role', body: `<div class="membership-flow"><span><small>AUTHENTICATED USER</small><b>${scenario.principal.id}</b></span><i>member of</i><span><small>GROUP PRINCIPAL</small><b>${scenario.membership.team}</b></span><i>assigned</i><span><small>ROLE</small><b>${scenario.role.name}</b></span></div><dl class="guide-dl"><dt>Team members</dt><dd>${scenario.membership.members.join(', ')}</dd><dt>Membership match</dt><dd>${membershipMatches ? 'Yes' : 'No'}</dd><dt>Assignment created by</dt><dd>${scenario.role.assignedBy}</dd></dl><aside><b>Concept:</b> A Team is a group principal, not a resource scope. Team membership lets a person inherit assignments held by that group.</aside>` },
    { eyebrow: 'CAPABILITY & ASSIGNMENT', title: 'Separate the role capability from its reach', body: `<div class="permission-anatomy"><span><small>APPLICATION</small>${resourceParts[0]}</span><i>:</i><span><small>RESOURCE</small>${resourceParts.slice(1).join(':')}</span><strong>::</strong><span class="verb"><small>VERB</small>${permissionParts[1]}</span></div><div class="assignment-record"><div><small>GROUP</small><code>${scenario.membership.team}</code></div><b>＋</b><div><small>ROLE CAPABILITY</small><code>${scenario.permission}</code></div><b>＋</b><div><small>ASSIGNED SCOPE</small><code>${scenario.scope.type}:${scenario.scope.resourceId}</code></div></div><aside><b>Concept:</b> The role explicitly provides repository-read capability. The assignment scope says where that capability reaches; project-read alone would not invent repository-read.</aside>${wireJsonBlock('grant set · Auth → application', buildGrantSet(scenario, membershipMatches))}` },
    { eyebrow: 'TARGET & CONTAINMENT', title: 'Ask the application whether scope contains the target', body: `<div class="context-compare"><div><small>ASSIGNMENT SCOPE</small><h3>${scenario.scope.type}</h3><dl><dt>Resource type</dt><dd>${scenario.scope.resourceType}</dd><dt>Resource ID</dt><dd>${scenario.scope.resourceId}</dd></dl></div><div><small>TRUSTED TARGET</small><h3>${scenario.target.id}</h3><dl><dt>Tenant</dt><dd>${scenario.target.tenant}</dd><dt>Project</dt><dd>${scenario.target.project}</dd><dt>Organization</dt><dd>${scenario.target.organization}</dd></dl></div></div><div class="containment-checks"><span class="${tenantMatches ? 'yes' : 'no'}">${tenantMatches ? '✓' : '×'} Tenant ${tenantMatches ? 'matches' : 'does not match'}</span><span class="${containmentMatches ? 'yes' : 'no'}">${containmentMatches ? '✓' : '×'} Scope ${containmentMatches ? 'contains' : 'does not contain'} repository</span></div><aside><b>Concept:</b> The code-hosting application owns the Project → Repository containment graph and supplies trusted target attributes. Auth does not infer it from names or URLs.</aside>${wireJsonBlock('resolved authorization request → PDP input', buildResolvedRequest(scenario))}` },
    { eyebrow: 'DECISION & ENFORCEMENT', title: allowed ? 'Access granted' : 'Access denied', body: `<div class="guided-result ${allowed ? 'allowed' : 'denied'}"><strong>${allowed ? 'ALLOW' : 'DENY'}</strong><p>${allowed ? 'Group membership, capability, tenant, and resource containment all match.' : !membershipMatches ? 'The user is not a member of the assigned team.' : !capabilityMatches ? 'The role does not contain the required capability.' : !tenantMatches ? 'The target belongs to another tenant.' : 'The assigned resource scope does not contain this repository.'}</p></div><div class="guided-output"><div><small>${allowed ? 'ENFORCED PREDICATE' : 'FAIL-CLOSED PREDICATE'}</small><pre>${allowed ? predicate : 'FALSE  /* no repository rows */'}</pre></div><div><small>AUDIT RECEIPT</small><code>principal=${scenario.principal.id}\nvia_group=${scenario.membership.team}\npermission=${scenario.permission}\ntarget=${scenario.target.id}\ndecision=${allowed ? 'allow' : 'deny'}</code></div></div><aside><b>Concept:</b> Membership identifies who receives the assignment; scope controls which resources it reaches. Both must be true at the same time.</aside>${wireJsonBlock('decision result → API + audit sink', buildDecisionResult(scenario, allowed, membershipMatches, capabilityMatches, tenantMatches, containmentMatches, predicate))}` },
  ]
  if (stepIndex === steps.length - 1 && allowed === scenario.expected) explored.add(scenario.id)
  const step = steps[stepIndex]

  app.innerHTML = `
    <header class="topbar">
      <a class="brand" href="/concept.html">${shield()}<span>Authorization Bench</span><small>Explain · inspect · decide</small></a>
      <nav class="page-tabs"><a href="/concept.html">Concept</a><a href="/concept.html?doc=hrms">HRMS</a><a href="/concept.html?doc=projects">Projects & repos</a><a class="active" href="/projects-explorer.html">Guided explorer</a></nav>
      <div class="canonical"><span>Canonical grammar</span><code>&lt;namespaced-noun&gt;::<b>verb</b></code></div>
    </header>
    <main>
      <section class="explorer-heading">
        <div><small>GUIDED REQUEST EXPLORER</small><h1>Projects, repositories, teams, and members</h1><p>Understand how a user's team membership receives a role at an exact repository or project subtree.</p></div>
        <nav class="domain-switch"><a href="/">HRMS · Payroll</a><a class="active" href="/projects-explorer.html">Projects · Repositories</a></nav>
      </section>
      <section class="challenge-strip" id="scenarios">
        <div class="section-heading"><div><small>WORKED REQUESTS</small><h2>Choose the story you want explained</h2></div><span>${explored.size} / ${scenarios.length} viewed</span></div>
        <div class="challenge-list project-scenarios">${scenarios.map(item => `<button class="challenge ${item.id === scenarioId ? 'active' : ''} ${explored.has(item.id) ? 'complete' : ''}" data-scenario="${item.id}"><span>${explored.has(item.id) ? '✓' : String(item.id).padStart(2, '0')}</span><div><b>${item.title}</b><small>${item.brief}</small></div><em>OUTCOME · ${item.expected ? 'ALLOW' : 'DENY'}</em></button>`).join('')}</div>
      </section>
      <section class="guided-walkthrough">
        <div class="guide-progress">${steps.map((item, index) => `<button class="${index === stepIndex ? 'active' : ''} ${index < stepIndex ? 'visited' : ''}" data-step="${index}"><span>${index < stepIndex ? '✓' : index + 1}</span><small>${item.eyebrow}</small></button>`).join('')}</div>
        <article class="guide-stage">
          <header><div><small>STEP ${stepIndex + 1} OF ${steps.length} · ${step.eyebrow}</small><h2>${step.title}</h2></div><span class="guide-scenario">${scenario.title}</span></header>
          <div class="guide-content">${step.body}</div>
          <footer class="guide-navigation"><button data-previous ${stepIndex === 0 ? 'disabled' : ''}>← Previous concept</button><span>Use the steps above to revisit any concept.</span>${stepIndex < steps.length - 1 ? '<button class="next" data-next>Next concept →</button>' : '<a href="#scenarios">Choose another story ↑</a>'}</footer>
        </article>
      </section>
    </main>
    <footer>Authorization Explanation Bench · Projects and repositories guided explorer.</footer>
  `
  document.querySelectorAll<HTMLButtonElement>('[data-scenario]').forEach(button => button.onclick = () => selectScenario(Number(button.dataset.scenario)))
  document.querySelectorAll<HTMLButtonElement>('[data-step]').forEach(button => button.onclick = () => { stepIndex = Number(button.dataset.step); render() })
  document.querySelector<HTMLButtonElement>('[data-previous]')?.addEventListener('click', () => { stepIndex = Math.max(0, stepIndex - 1); render() })
  document.querySelector<HTMLButtonElement>('[data-next]')?.addEventListener('click', () => { stepIndex = Math.min(4, stepIndex + 1); render() })
}

render()
