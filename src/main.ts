import './styles.css'
import { payrollScenario, type Grant, type Permission, type Principal, type PrincipalId, type Resource, type Scope, type ScopeType } from './scenario-pack'

type QuestionStatus = 'open' | 'proposed' | 'agreed'

interface DesignQuestion {
  id: number
  question: string
  status: QuestionStatus
}

const { principals, permissions, resources, challenges } = payrollScenario

let designQuestions: DesignQuestion[] = [
  { id: 1, question: 'Should assignment scope live in Auth or an HRMS governance binding?', status: 'open' },
  { id: 2, question: 'How are dynamic scopes such as employee_self represented canonically?', status: 'open' },
  { id: 3, question: 'Can roles carry a default scope, or only assignments?', status: 'open' },
  { id: 4, question: 'How does device-login delegation constrain the runner to the user’s authority?', status: 'open' },
  { id: 5, question: 'What enters the access token and what stays server-side?', status: 'open' },
  { id: 6, question: 'Do explicit deny assignments override every allow?', status: 'open' },
]

let selectedPrincipal: PrincipalId = 'vinay'
let selectedPermission: Permission = 'hrms:payroll:ledger::read'
let selectedResource = 'PAY-000005'
let grants: Grant[] = structuredClone(payrollScenario.initialGrants)
let lastDecision: null | { allowed: boolean; title: string; reason: string; grant?: Grant; predicate: string; checks: { label: string; pass: boolean; detail: string }[] } = null
let questionCounter = 7
let activeChallenge: number | null = 1
let guideStep = 0
let completedChallenges = new Set<number>()

const app = document.querySelector<HTMLDivElement>('#app')!

function icon(name: 'user' | 'key' | 'target' | 'shield' | 'spark' | 'book') {
  const paths = {
    user: '<circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/>',
    key: '<circle cx="8" cy="15" r="4"/><path d="m11 12 9-9m-3 3 3 3m-6 0 3 3"/>',
    target: '<circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="5"/><circle cx="12" cy="12" r="1"/>',
    shield: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/>',
    spark: '<path d="m12 3-1.7 4.3L6 9l4.3 1.7L12 15l1.7-4.3L18 9l-4.3-1.7L12 3Z"/><path d="m5 16-.8 2.2L2 19l2.2.8L5 22l.8-2.2L8 19l-2.2-.8L5 16Z"/>',
    book: '<path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20V3H6.5A2.5 2.5 0 0 0 4 5.5v14Z"/><path d="M4 19.5V6"/>',
  }
  return `<svg viewBox="0 0 24 24" aria-hidden="true">${paths[name]}</svg>`
}

function scopeLabel(scope: Scope) {
  const names: Record<ScopeType, string> = { employee_self: 'Employee · self', employee: `Employee · ${scope.targetId}`, department: `Department · ${scope.targetId}`, tenant: `Tenant · ${scope.targetId}` }
  return names[scope.type]
}

function escapeHtml(value: string) {
  return value.replace(/[&<>'"]/g, character => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' })[character]!)
}

function scopeContains(scope: Scope, principal: Principal, resource: Resource) {
  if (scope.type === 'employee_self') return principal.employeeId !== null && resource.employeeId === principal.employeeId
  if (scope.type === 'employee') return resource.employeeId === scope.targetId
  if (scope.type === 'department') return resource.departmentId === scope.targetId
  return resource.tenantId === scope.targetId
}

function predicateFor(scope: Scope, principal: Principal) {
  if (scope.type === 'employee_self') return `tenant_id = token.tenant_id\nAND ledger_owner_id = ${principal.employeeId ?? '∅'}`
  if (scope.type === 'employee') return `tenant_id = token.tenant_id\nAND ledger_owner_id = ${scope.targetId}`
  if (scope.type === 'department') return `tenant_id = token.tenant_id\nAND department_id = ${scope.targetId}`
  return `tenant_id = ${scope.targetId}`
}

function wireJsonBlock(title: string, payload: unknown) {
  return `<details class="wire-json"><summary>View wire JSON · ${title}</summary><pre>${escapeHtml(JSON.stringify(payload, null, 2))}</pre></details>`
}

// What the Payroll API receives from the caller. Deliberately thin: the caller
// may name a target, but cannot assert scope, tenant, or ownership.
function buildClientRequest(principal: Principal, permission: Permission, target: Resource) {
  const [resource, verb] = permission.split('::')
  const path = resource.split(':').slice(2).join('/')
  return {
    actor_token: `session:user:${principal.id}`,
    operation: `${verb.toUpperCase()} /payroll/${path}/${target.id}`,
    requested_permission: permission,
    requested_target_id: target.id,
  }
}

// What Auth returns for this principal: the grant set an assignment produced.
function buildGrantSet(principal: Principal, principalGrants: Grant[]) {
  return {
    principal_id: `user:${principal.id}`,
    grants: principalGrants.map(g => ({ id: g.id, permission: g.permission, scope: g.scope, valid: g.valid, source: g.source })),
  }
}

// What the HRMS Authorization Agent hands to the decision step, after crossing
// the trust boundary: every field here is HRMS-resolved, never caller-supplied.
function buildResolvedRequest(principal: Principal, permission: Permission, target: Resource) {
  return {
    principal: { id: `user:${principal.id}`, tenant_id: payrollScenario.tenantId },
    permission,
    resolved_context: principal.employeeId ? { employee_self: principal.employeeId } : {},
    target: {
      type: permission.split('::')[0],
      id: target.id,
      tenant_id: target.tenantId,
      owner_employee_id: target.employeeId,
      department_id: target.departmentId,
    },
  }
}

function buildDecisionResult(decision: NonNullable<typeof lastDecision>) {
  return {
    decision: decision.allowed ? 'allow' : 'deny',
    matched_grant_id: decision.allowed && decision.grant ? decision.grant.id : null,
    checks: Object.fromEntries(decision.checks.map(c => [c.label.toLowerCase().replace(/ /g, '_'), c.pass ? 'pass' : 'fail'])),
    reason: decision.reason,
    enforced_predicate: decision.predicate,
    audit: { principal: `user:${selectedPrincipal}`, permission: selectedPermission, target: selectedResource, decision: decision.allowed ? 'allow' : 'deny' },
  }
}

function evaluate(focusGuide = false) {
  const principal = principals.find(p => p.id === selectedPrincipal)!
  const resource = resources.find(r => r.id === selectedResource)!
  const matching = grants.filter(g => g.principalId === selectedPrincipal && g.permission === selectedPermission && g.valid)
  const tenantPass = resource.tenantId === 'TENANT-001'
  const scopedGrant = matching.find(g => tenantPass && scopeContains(g.scope, principal, resource))
  const capabilityPass = matching.length > 0
  const scopePass = Boolean(scopedGrant)
  const allowed = capabilityPass && scopePass && tenantPass
  const effectiveGrant = scopedGrant ?? matching[0]
  const checks = [
    { label: 'Identity', pass: true, detail: `${principal.name} authenticated in TENANT-001` },
    { label: 'Capability', pass: capabilityPass, detail: capabilityPass ? 'An active assignment contains the required permission' : 'No active assignment contains this permission' },
    { label: 'Tenant boundary', pass: tenantPass, detail: tenantPass ? 'Resource belongs to authenticated tenant' : `${resource.tenantId} is outside TENANT-001` },
    { label: 'Scope reach', pass: scopePass, detail: scopePass ? `${scopeLabel(scopedGrant!.scope)} contains ${resource.employeeId}` : 'No matching grant reaches the target owner' },
  ]
  lastDecision = {
    allowed,
    title: allowed ? 'Access granted' : 'Access denied',
    reason: allowed ? 'Capability, assigned reach, and request target intersect.' : !capabilityPass ? 'The principal lacks the required capability.' : !tenantPass ? 'Tenant isolation rejects the target before row access.' : 'The capability exists, but its assigned scope does not reach this ledger.',
    grant: effectiveGrant,
    predicate: effectiveGrant ? predicateFor(effectiveGrant.scope, principal) : 'FALSE  /* fail closed */',
    checks,
  }
  const challenge = challenges.find(item => item.id === activeChallenge)
  if (challenge && challenge.principalId === selectedPrincipal && challenge.permission === selectedPermission && challenge.resourceId === selectedResource && challenge.expected === allowed && !completedChallenges.has(challenge.id)) {
    completedChallenges.add(challenge.id)
  }
  render()
  document.querySelector(focusGuide ? '.guided-walkthrough' : '.decision-panel')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}

function loadChallenge(id: number) {
  const challenge = challenges.find(item => item.id === id)!
  activeChallenge = id
  selectedPrincipal = challenge.principalId
  selectedPermission = challenge.permission
  selectedResource = challenge.resourceId
  guideStep = 0
  lastDecision = null
  render()
  document.querySelector('.guided-walkthrough')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function moveGuide(direction: number) {
  guideStep = Math.max(0, Math.min(4, guideStep + direction))
  render()
  document.querySelector('.guided-walkthrough')?.scrollIntoView({ behavior: 'smooth', block: 'center' })
}

function cycleQuestion(id: number) {
  const question = designQuestions.find(item => item.id === id)
  if (!question) return
  question.status = question.status === 'open' ? 'proposed' : question.status === 'proposed' ? 'agreed' : 'open'
  render()
}

function exportSession() {
  const payload = {
    exported_at: new Date().toISOString(),
    model: 'principal + permission + scope + trusted target context',
    grammar: '<namespaced-noun>::<verb>',
    explored_scenarios: [...completedChallenges],
    grants,
    design_questions: designQuestions,
  }
  const link = document.createElement('a')
  link.href = URL.createObjectURL(new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' }))
  link.download = 'permission-quest-session.json'
  link.click()
  URL.revokeObjectURL(link.href)
}

function render() {
  const principal = principals.find(p => p.id === selectedPrincipal)!
  const principalGrants = grants.filter(g => g.principalId === selectedPrincipal && g.valid)
  const target = resources.find(resource => resource.id === selectedResource)!
  const activeScenario = challenges.find(scenario => scenario.id === activeChallenge) ?? challenges[0]
  const comparisonScenarioId = ({ 1: 2, 2: 1, 3: 4, 4: 3, 5: 1, 6: 3 } as Record<number, number>)[activeScenario.id]
  const comparisonScenario = challenges.find(scenario => scenario.id === comparisonScenarioId)
  const matchingGrant = principalGrants.find(grant => grant.permission === selectedPermission)
  const permissionParts = selectedPermission.split('::')
  const resourceParts = permissionParts[0].split(':')
  const scopeMatches = matchingGrant ? scopeContains(matchingGrant.scope, principal, target) : false
  const tenantMatches = target.tenantId === payrollScenario.tenantId
  const guideSteps = [
    {
      eyebrow: 'REQUEST STORY',
      title: 'Start with a human question',
      body: `<p>${activeScenario.brief}</p><div class="guide-facts"><span><small>PRINCIPAL</small><b>${principal.name}</b><code>user:${principal.id}</code></span><span><small>OPERATION</small><b>${permissions.find(item => item.value === selectedPermission)?.label}</b><code>${selectedPermission}</code></span><span><small>TARGET</small><b>${target.employeeName}’s ledger</b><code>${target.id}</code></span></div><aside><b>Concept:</b> A request combines an authenticated principal, an operation, and a target. None of these alone determines authority.</aside>${wireJsonBlock('client request → Payroll API', buildClientRequest(principal, selectedPermission, target))}`,
    },
    {
      eyebrow: 'CAPABILITY',
      title: 'Read the permission—without inventing scope',
      body: `<div class="permission-anatomy"><span><small>APPLICATION</small>${resourceParts[0]}</span><i>:</i><span><small>DOMAIN</small>${resourceParts[1]}</span><i>:</i><span><small>RESOURCE</small>${resourceParts.slice(2).join(':')}</span><strong>::</strong><span class="verb"><small>VERB</small>${permissionParts[1]}</span></div><p><code>${selectedPermission}</code> names the required capability. It does not identify ${principal.name}, ${target.employeeId}, ${target.id}, or any tenant.</p><aside><b>Concept:</b> Permission strings remain stable. User, tenant, resource ID, and reach belong to assignments and request context.</aside>`,
    },
    {
      eyebrow: 'ASSIGNMENT & SCOPE',
      title: matchingGrant ? 'Find how the principal received reach' : 'Notice the missing authority',
      body: (matchingGrant ? `<div class="assignment-record"><div><small>PRINCIPAL</small><code>user:${principal.id}</code></div><b>＋</b><div><small>PERMISSION</small><code>${matchingGrant.permission}</code></div><b>＋</b><div><small>ASSIGNED SCOPE</small><code>${scopeLabel(matchingGrant.scope)}</code></div></div><dl class="guide-dl"><dt>Permission received through</dt><dd>${matchingGrant.source.type === 'role' ? `Role · ${matchingGrant.source.name}` : matchingGrant.source.name}</dd><dt>Scope type defined by</dt><dd>${matchingGrant.scope.type === 'tenant' ? 'Auth' : 'HRMS'}</dd><dt>Assignment created by</dt><dd>${matchingGrant.source.assignedBy}</dd><dt>Assignment stored by</dt><dd>Auth</dd><dt>Scope meaning resolved by</dt><dd>${matchingGrant.scope.type === 'tenant' ? 'Auth' : 'HRMS Authorization Agent'}</dd></dl><aside><b>Concept:</b> A role groups capabilities; its assignment gives those capabilities reach. Applications define scope vocabulary, applications create scopeable resources, and tenant administrators create the binding.</aside>` : `<div class="missing-assignment">No active assignment contains <code>${selectedPermission}</code> for <code>user:${principal.id}</code>.</div><p>A permission definition may exist globally without being granted to this principal. A missing assignment must fail closed.</p><aside><b>Concept:</b> Authentication is not authorization, and a permission catalogue is not a grant.</aside>`) + wireJsonBlock('grant set · Auth → HRMS', buildGrantSet(principal, grants.filter(g => g.principalId === selectedPrincipal))),
    },
    {
      eyebrow: 'TRUSTED TARGET CONTEXT',
      title: 'Compare assigned reach with the real resource',
      body: `<div class="context-compare"><div><small>ASSIGNMENT REACH</small><h3>${matchingGrant ? scopeLabel(matchingGrant.scope) : 'No matching scope'}</h3><p>${matchingGrant?.scope.type === 'employee_self' ? `HRMS resolves user:${principal.id} → ${principal.employeeId}` : matchingGrant ? 'The assignment contains a concrete tenant or resource boundary.' : 'Nothing can contain the target without a matching grant.'}</p></div><div><small>TRUSTED TARGET</small><h3>${target.id}</h3><dl><dt>Tenant</dt><dd>${target.tenantId}</dd><dt>Owner</dt><dd>${target.employeeId}</dd><dt>Department</dt><dd>${target.departmentId}</dd></dl></div></div><div class="containment-checks"><span class="${tenantMatches ? 'yes' : 'no'}">${tenantMatches ? '✓' : '×'} Tenant boundary ${tenantMatches ? 'matches' : 'does not match'}</span><span class="${scopeMatches ? 'yes' : 'no'}">${scopeMatches ? '✓' : '×'} Assignment scope ${scopeMatches ? 'contains' : 'does not contain'} target</span></div><aside><b>Concept:</b> HRMS supplies trusted target attributes. Caller-provided employee IDs or tenant IDs cannot establish authority.</aside>${wireJsonBlock('resolved authorization request → PDP input', buildResolvedRequest(principal, selectedPermission, target))}`,
    },
    {
      eyebrow: 'DECISION, ENFORCEMENT & AUDIT',
      title: lastDecision ? lastDecision.title : 'Evaluate the complete authority',
      body: lastDecision ? `<div class="guided-result ${lastDecision.allowed ? 'allowed' : 'denied'}"><strong>${lastDecision.allowed ? 'ALLOW' : 'DENY'}</strong><p>${lastDecision.reason}</p></div><div class="guided-output"><div><small>ENFORCED PREDICATE</small><pre>${lastDecision.predicate}</pre></div><div><small>AUDIT RECEIPT</small><code>principal=user:${selectedPrincipal}\npermission=${selectedPermission}\ntarget=${selectedResource}\ndecision=${lastDecision.allowed ? 'allow' : 'deny'}</code></div></div><aside><b>Concept:</b> The API must enforce the decision as a data restriction. The audit receipt records which principal, assignment context, target, and policy produced it.</aside>${wireJsonBlock('decision result → API + audit sink', buildDecisionResult(lastDecision))}${comparisonScenario ? `<button class="compare-request" data-compare-scenario="${comparisonScenario.id}"><span><small>CHANGE ONE THING</small><b>${comparisonScenario.brief}</b></span><em>Compare explanation →</em></button>` : ''}` : `<p>The system now has the complete story: principal, capability, assignment scope, and trusted target. Evaluate their intersection.</p><button class="guide-evaluate" data-guide-evaluate>Evaluate and explain →</button><aside><b>Expected:</b> This worked scenario should produce <strong>${activeScenario.expected ? 'ALLOW' : 'DENY'}</strong>. The explanation matters more than the label.</aside>`,
    },
  ]
  const currentGuide = guideSteps[guideStep]
  app.innerHTML = `
    <header class="topbar">
      <a class="brand" href="/concept.html">${icon('shield')}<span>Authorization Bench</span><small>Explain · inspect · decide</small></a>
      <nav class="page-tabs"><a href="/concept.html">Concept</a><a href="/concept.html?doc=hrms">HRMS</a><a href="/concept.html?doc=projects">Projects & repos</a><a class="active" href="/">Guided explorer</a></nav>
      <div class="canonical"><span>Canonical grammar</span><code>&lt;namespaced-noun&gt;::<b>verb</b></code></div>
    </header>

    <main>
      <section class="hero" id="play">
        <div class="eyebrow">INTERACTIVE EXPLANATION · PAYROLL</div>
        <h1>A permission is only half<br/>of an <em>authority.</em></h1>
        <p>Construct a grant, apply it to a payroll request, and inspect exactly where access is allowed or denied.</p>
        <nav class="domain-switch"><a class="active" href="/">HRMS · Payroll</a><a href="/projects-explorer.html">Projects · Repositories</a></nav>
        <div class="equation">
          <span>${icon('key')}<small>CAPABILITY</small>Permission</span><b>∩</b>
          <span>${icon('target')}<small>REACH</small>Scope</span><b>∩</b>
          <span>${icon('spark')}<small>REQUEST</small>Target</span><b>=</b>
          <span class="result"><small>DECISION</small>Authority</span>
        </div>
      </section>

      <section class="learning-path">
        <div class="section-heading"><div><small>WHAT DO YOU WANT TO UNDERSTAND?</small><h2>Choose a guided explanation</h2></div><span>Each path starts with a familiar story</span></div>
        <div class="learning-path-grid">
          <a href="/concept.html?doc=hrms"><span>01</span><div><small>I WANT TO CONFIGURE ACCESS</small><b>How does a tenant create Payroll Admin?</b><p>Walk through permissions, role creation, scope selection, assignment records, and tenant protection.</p></div><em>Start explanation →</em></a>
          <button data-path-scenario="1"><span>02</span><div><small>I AM AN EMPLOYEE</small><b>Why can I see my salary but not another employee’s?</b><p>Follow identity, employee-self scope, target ownership, and the enforced row restriction.</p></div><em>Start explanation ↓</em></button>
          <button data-path-scenario="3"><span>03</span><div><small>I AM A PAYROLL ADMIN</small><b>Why can I read payroll only inside my tenant?</b><p>See how a tenant-wide assignment reaches employees without crossing the tenant boundary.</p></div><em>Start explanation ↓</em></button>
          <a href="/projects-explorer.html"><span>04</span><div><small>I WORK WITH TEAMS & REPOS</small><b>How do team membership and project scope combine?</b><p>Separate group principals from exact-resource and subtree assignment scopes.</p></div><em>Start guided explorer →</em></a>
        </div>
      </section>

      <section class="challenge-strip" id="scenarios">
        <div class="section-heading"><div><small>MORE WORKED REQUESTS</small><h2>Choose the story you want explained</h2></div><span>${completedChallenges.size} / ${challenges.length} viewed</span></div>
        <div class="challenge-list">
          ${challenges.map(challenge => `<button class="challenge ${activeChallenge === challenge.id ? 'active' : ''} ${completedChallenges.has(challenge.id) ? 'complete' : ''}" data-challenge="${challenge.id}"><span>${completedChallenges.has(challenge.id) ? '✓' : String(challenge.id).padStart(2, '0')}</span><div><b>${challenge.title}</b><small>${challenge.brief}</small></div><em>EXPECTED · ${challenge.expected ? 'ALLOW' : 'DENY'}</em></button>`).join('')}
        </div>
      </section>

      <section class="guided-walkthrough">
        <div class="guide-progress">
          ${guideSteps.map((step, index) => `<button class="${index === guideStep ? 'active' : ''} ${index < guideStep ? 'visited' : ''}" data-guide-step="${index}"><span>${index < guideStep ? '✓' : index + 1}</span><small>${step.eyebrow}</small></button>`).join('')}
        </div>
        <article class="guide-stage">
          <header><div><small>STEP ${guideStep + 1} OF ${guideSteps.length} · ${currentGuide.eyebrow}</small><h2>${currentGuide.title}</h2></div><span class="guide-scenario">${activeScenario.title}</span></header>
          <div class="guide-content">${currentGuide.body}</div>
          <footer class="guide-navigation">
            <button data-guide-previous ${guideStep === 0 ? 'disabled' : ''}>← Previous concept</button>
            <span>Use the steps above to revisit any concept.</span>
            ${guideStep < guideSteps.length - 1 ? '<button class="next" data-guide-next>Next concept →</button>' : '<a href="#scenarios">Choose another story ↑</a>'}
          </footer>
        </article>
      </section>

      <section class="rules">
        <div class="section-heading"><div><small>THE RULEBOOK</small><h2>What the string does—and does not—say</h2></div></div>
        <div class="rule-grid">
          <div><span>01</span><h3>Permission is capability</h3><code>hrms:payroll:ledger::read</code><p>Stable across users and scopes. It says what operation exists.</p></div>
          <div><span>02</span><h3>Assignment carries reach</h3><code>principal + permission + scope</code><p>The same capability can be narrow for an employee and broad for an administrator.</p></div>
          <div><span>03</span><h3>Context must be trusted</h3><code>user → employee mapping</code><p>Self is resolved by HRMS. It never comes from a caller-controlled parameter.</p></div>
          <div><span>04</span><h3>Missing scope fails closed</h3><code>permission − scope = no authority</code><p>A naked permission assignment is incomplete and cannot authorize row access.</p></div>
        </div>
      </section>

      <section class="boundary" id="boundary">
        <div class="section-heading"><div><small>RESPONSIBILITY MAP</small><h2>Auth proves the grant. HRMS understands the resource.</h2></div></div>
        <div class="boundary-map">
          <article class="system-card auth-card"><header><span>01</span><div><small>GENERIC PLATFORM</small><h3>Auth</h3></div></header><ul><li>Principal identity</li><li>Roles and direct grants</li><li>Permission strings</li><li>Assignment scope descriptor</li><li>Validity and assignment version</li></ul><div class="output"><small>PRODUCES</small><code>authenticated grant set</code></div></article>
          <div class="boundary-arrow"><span>principal</span><span>permission</span><span>scope</span><b>→</b></div>
          <article class="system-card agent-card"><header><span>02</span><div><small>DOMAIN POLICY</small><h3>HRMS Authorization Agent</h3></div></header><ul><li>API operation mapping</li><li>User → employee resolution</li><li>Meaning of HRMS scope types</li><li>Target resource attributes</li><li>Allow/deny and row restriction</li></ul><div class="output"><small>PRODUCES</small><code>decision + enforced predicate</code></div></article>
          <div class="boundary-arrow"><span>decision</span><span>predicate</span><b>→</b></div>
          <article class="system-card api-card"><header><span>03</span><div><small>ENFORCEMENT POINT</small><h3>Payroll API</h3></div></header><ul><li>Never trusts caller scope</li><li>Applies tenant restriction</li><li>Applies owner restriction</li><li>Returns only authorized rows</li><li>Emits business audit event</li></ul><div class="output"><small>RULE</small><code>no decision → no data</code></div></article>
        </div>
        <div class="contract-line"><code>permission</code><b>=</b><span>capability</span><i>+</i><code>assignment scope</code><b>=</b><span>reach</span><i>+</i><code>trusted context</code><b>=</b><span>target</span></div>
      </section>

      <section class="questions" id="questions">
        <div class="section-heading"><div><small>DESIGN WORKBENCH</small><h2>Questions that must become decisions</h2></div><button id="export-session">Export session JSON ↓</button></div>
        <p class="questions-intro">Click a status to advance it. The board keeps uncertainty visible; an explanation should not accidentally turn an assumption into architecture.</p>
        <div class="question-board">
          ${designQuestions.map(item => `<article><span class="question-number">Q${String(item.id).padStart(2, '0')}</span><p>${escapeHtml(item.question)}</p><button class="status-${item.status}" data-question="${item.id}">${item.status.toUpperCase()} <b>→</b></button></article>`).join('')}
        </div>
        <form id="question-form"><label for="new-question">A question discovered while exploring</label><div><input id="new-question" required maxlength="180" placeholder="What authority rule is still ambiguous?"/><button type="submit">Add to board ＋</button></div></form>
        <div class="legend"><span><i class="open-dot"></i>OPEN · unanswered</span><span><i class="proposed-dot"></i>PROPOSED · candidate shape</span><span><i class="agreed-dot"></i>AGREED · accepted for the model</span></div>
      </section>
    </main>
    <footer>Authorization Explanation Bench · An in-memory design instrument, not a production policy engine.</footer>
  `

  document.querySelectorAll<HTMLButtonElement>('[data-challenge]').forEach(el => el.onclick = () => loadChallenge(Number(el.dataset.challenge)))
  document.querySelectorAll<HTMLButtonElement>('[data-path-scenario]').forEach(el => el.onclick = () => loadChallenge(Number(el.dataset.pathScenario)))
  document.querySelectorAll<HTMLButtonElement>('[data-guide-step]').forEach(el => el.onclick = () => { guideStep = Number(el.dataset.guideStep); render() })
  document.querySelector<HTMLButtonElement>('[data-guide-previous]')?.addEventListener('click', () => moveGuide(-1))
  document.querySelector<HTMLButtonElement>('[data-guide-next]')?.addEventListener('click', () => moveGuide(1))
  document.querySelector<HTMLButtonElement>('[data-guide-evaluate]')?.addEventListener('click', () => evaluate(true))
  document.querySelector<HTMLButtonElement>('[data-compare-scenario]')?.addEventListener('click', element => loadChallenge(Number((element.currentTarget as HTMLButtonElement).dataset.compareScenario)))
  document.querySelectorAll<HTMLButtonElement>('[data-question]').forEach(el => el.onclick = () => cycleQuestion(Number(el.dataset.question)))
  document.querySelector<HTMLButtonElement>('#export-session')!.onclick = exportSession
  document.querySelector<HTMLFormElement>('#question-form')!.onsubmit = event => {
    event.preventDefault()
    const input = document.querySelector<HTMLInputElement>('#new-question')!
    const question = input.value.trim()
    if (question) designQuestions.push({ id: questionCounter++, question, status: 'open' })
    render()
    document.querySelector('#questions')?.scrollIntoView({ block: 'start' })
  }
}

render()
