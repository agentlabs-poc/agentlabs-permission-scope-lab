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
let selectedScope: ScopeType = 'employee_self'
let selectedScopeTarget = 'EMP-005'
let selectedResource = 'PAY-000005'
let grants: Grant[] = structuredClone(payrollScenario.initialGrants)
let lastDecision: null | { allowed: boolean; title: string; reason: string; grant?: Grant; predicate: string; checks: { label: string; pass: boolean; detail: string }[] } = null
let grantCounter = 3
let questionCounter = 7
let activeChallenge: number | null = null
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

function resolveScope(type: ScopeType): Scope {
  if (type === 'employee_self') return { type }
  if (type === 'employee') return { type, targetId: selectedScopeTarget.startsWith('EMP') ? selectedScopeTarget : 'EMP-005' }
  if (type === 'department') return { type, targetId: selectedScopeTarget === 'FIN' ? 'FIN' : 'ENG' }
  return { type, targetId: selectedScopeTarget === 'TENANT-002' ? 'TENANT-002' : 'TENANT-001' }
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

function evaluate() {
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
  document.querySelector('.decision-panel')?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
}

function loadChallenge(id: number) {
  const challenge = challenges.find(item => item.id === id)!
  activeChallenge = id
  selectedPrincipal = challenge.principalId
  selectedPermission = challenge.permission
  selectedResource = challenge.resourceId
  selectedScopeTarget = principals.find(p => p.id === selectedPrincipal)?.employeeId ?? 'EMP-005'
  lastDecision = null
  render()
  document.querySelector('.game-grid')?.scrollIntoView({ behavior: 'smooth' })
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

function addGrant() {
  const scope = resolveScope(selectedScope)
  const exists = grants.some(g => g.principalId === selectedPrincipal && g.permission === selectedPermission && JSON.stringify(g.scope) === JSON.stringify(scope) && g.valid)
  if (!exists) grants.push({ id: grantCounter++, principalId: selectedPrincipal, permission: selectedPermission, scope, valid: true })
  lastDecision = null
  render()
}

function render() {
  const principal = principals.find(p => p.id === selectedPrincipal)!
  const principalGrants = grants.filter(g => g.principalId === selectedPrincipal && g.valid)
  app.innerHTML = `
    <header class="topbar">
      <a class="brand" href="/concept.html">${icon('shield')}<span>Authorization Bench</span><small>Explain · inspect · decide</small></a>
      <nav class="page-tabs"><a href="/concept.html">Concept</a><a href="/concept.html?doc=hrms">HRMS</a><a href="/concept.html?doc=projects">Projects & repos</a><a class="active" href="/">Request explorer</a></nav>
      <div class="canonical"><span>Canonical grammar</span><code>&lt;namespaced-noun&gt;::<b>verb</b></code></div>
    </header>

    <main>
      <section class="hero" id="play">
        <div class="eyebrow">INTERACTIVE EXPLANATION · PAYROLL</div>
        <h1>A permission is only half<br/>of an <em>authority.</em></h1>
        <p>Construct a grant, apply it to a payroll request, and inspect exactly where access is allowed or denied.</p>
        <div class="equation">
          <span>${icon('key')}<small>CAPABILITY</small>Permission</span><b>∩</b>
          <span>${icon('target')}<small>REACH</small>Scope</span><b>∩</b>
          <span>${icon('spark')}<small>REQUEST</small>Target</span><b>=</b>
          <span class="result"><small>DECISION</small>Authority</span>
        </div>
      </section>

      <section class="challenge-strip">
        <div class="section-heading"><div><small>WORKED REQUESTS</small><h2>Choose a scenario to explain</h2></div><span>${completedChallenges.size} / ${challenges.length} explored</span></div>
        <div class="challenge-list">
          ${challenges.map(challenge => `<button class="challenge ${activeChallenge === challenge.id ? 'active' : ''} ${completedChallenges.has(challenge.id) ? 'complete' : ''}" data-challenge="${challenge.id}"><span>${completedChallenges.has(challenge.id) ? '✓' : String(challenge.id).padStart(2, '0')}</span><div><b>${challenge.title}</b><small>${challenge.brief}</small></div><em>EXPECTED · ${challenge.expected ? 'ALLOW' : 'DENY'}</em></button>`).join('')}
        </div>
      </section>

      <section class="game-grid">
        <article class="panel step-panel">
          <div class="step-title"><span>1</span><div><small>SELECT PRINCIPAL</small><h2>Who is asking?</h2></div>${icon('user')}</div>
          <div class="principal-list">
            ${principals.map(p => `<button class="principal ${p.id === selectedPrincipal ? 'active' : ''}" data-principal="${p.id}"><span class="avatar" style="--avatar:${p.color}">${p.initials}</span><span><b>${p.name}</b><small>${p.title}</small></span><i></i></button>`).join('')}
          </div>
          <div class="identity-card"><span>Trusted identity context</span><dl><dt>Principal</dt><dd>user:${principal.id}</dd><dt>Tenant</dt><dd>TENANT-001</dd><dt>Employee</dt><dd>${principal.employeeId ?? 'not linked'}</dd></dl></div>
        </article>

        <article class="panel step-panel builder-panel">
          <div class="step-title"><span>2</span><div><small>CONSTRUCT ASSIGNMENT</small><h2>Capability + reach</h2></div>${icon('key')}</div>
          <label>Permission <select id="permission">${permissions.map(p => `<option value="${p.value}" ${p.value === selectedPermission ? 'selected' : ''}>${p.label}</option>`).join('')}</select></label>
          <div class="permission-preview"><span>NOUN</span><code>${selectedPermission.split('::')[0]}</code><strong>::</strong><span>VERB</span><code>${selectedPermission.split('::')[1]}</code></div>
          <label>Assignment scope <select id="scope">
            <option value="employee_self" ${selectedScope === 'employee_self' ? 'selected' : ''}>Employee · self (dynamic)</option>
            <option value="employee" ${selectedScope === 'employee' ? 'selected' : ''}>One employee</option>
            <option value="department" ${selectedScope === 'department' ? 'selected' : ''}>One department</option>
            <option value="tenant" ${selectedScope === 'tenant' ? 'selected' : ''}>Entire tenant</option>
          </select></label>
          ${selectedScope === 'employee' ? `<label>Employee <select id="scope-target">${resources.filter(r => r.tenantId === 'TENANT-001').map(r => `<option value="${r.employeeId}">${r.employeeName} · ${r.employeeId}</option>`).join('')}</select></label>` : ''}
          ${selectedScope === 'department' ? '<label>Department <select id="scope-target"><option value="ENG">Engineering</option><option value="FIN">Finance</option></select></label>' : ''}
          ${selectedScope === 'tenant' ? '<label>Tenant <select id="scope-target"><option value="TENANT-001">TENANT-001</option><option value="TENANT-002">TENANT-002</option></select></label>' : ''}
          <button class="primary" id="add-grant">Add assignment for ${principal.name} <span>＋</span></button>
        </article>

        <article class="panel step-panel request-panel">
          <div class="step-title"><span>3</span><div><small>EVALUATE REQUEST</small><h2>Select a ledger target</h2></div>${icon('target')}</div>
          <label>Operation <select id="request-permission">${permissions.map(p => `<option value="${p.value}" ${p.value === selectedPermission ? 'selected' : ''}>${p.verb.toUpperCase()} · ${p.label}</option>`).join('')}</select></label>
          <label>Target resource <select id="resource">${resources.map(r => `<option value="${r.id}" ${r.id === selectedResource ? 'selected' : ''}>${r.employeeName} · ${r.id}</option>`).join('')}</select></label>
          ${(() => { const r = resources.find(x => x.id === selectedResource)!; return `<div class="resource-card"><span class="resource-type">PAYROLL LEDGER</span><b>${r.id}</b><strong>${r.amount}</strong><dl><dt>Owner</dt><dd>${r.employeeName} · ${r.employeeId}</dd><dt>Department</dt><dd>${r.departmentId}</dd><dt>Tenant</dt><dd>${r.tenantId}</dd></dl></div>` })()}
          <button class="launch" id="evaluate">Evaluate request <span>→</span></button>
        </article>
      </section>

      <section class="grant-deck">
        <div class="section-heading"><div><small>EFFECTIVE ASSIGNMENTS</small><h2>${principal.name}’s current authority</h2></div><span>${principalGrants.length} active grant${principalGrants.length === 1 ? '' : 's'}</span></div>
        <div class="grant-list">
          ${principalGrants.length ? principalGrants.map(g => `<div class="grant-card"><div class="grant-glyph">${icon('key')}</div><div><code>${g.permission}</code><p>${scopeLabel(g.scope)}</p></div><button data-revoke="${g.id}" title="Revoke grant">×</button></div>`).join('') : '<div class="empty">No active authority. Add an assignment above.</div>'}
        </div>
      </section>

      ${lastDecision ? `<section class="decision-panel ${lastDecision.allowed ? 'allow' : 'deny'}">
        <div class="decision-head"><div class="decision-seal">${lastDecision.allowed ? '✓' : '×'}</div><div><small>AUTHORIZATION DECISION</small><h2>${lastDecision.title}</h2><p>${lastDecision.reason}</p></div><strong>${lastDecision.allowed ? 'ALLOW' : 'DENY'}</strong></div>
        <div class="decision-body">
          <ol>${lastDecision.checks.map((c, i) => `<li class="${c.pass ? 'pass' : 'fail'}"><span>${c.pass ? '✓' : '×'}</span><div><small>0${i + 1}</small><b>${c.label}</b><p>${c.detail}</p></div></li>`).join('')}</ol>
          <div class="predicate"><span>ENFORCED DATA PREDICATE</span><pre>${lastDecision.predicate}</pre><p>The request cannot broaden this predicate.</p></div>
        </div>
        <div class="audit"><span>${icon('book')} AUDIT RECEIPT</span><code>principal=user:${selectedPrincipal} · permission=${selectedPermission} · target=${selectedResource} · decision=${lastDecision.allowed ? 'allow' : 'deny'}</code></div>
      </section>` : ''}

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
        <form id="question-form"><label for="new-question">A question discovered while playing</label><div><input id="new-question" required maxlength="180" placeholder="What authority rule is still ambiguous?"/><button type="submit">Add to board ＋</button></div></form>
        <div class="legend"><span><i class="open-dot"></i>OPEN · unanswered</span><span><i class="proposed-dot"></i>PROPOSED · candidate shape</span><span><i class="agreed-dot"></i>AGREED · accepted for the model</span></div>
      </section>
    </main>
    <footer>Authorization Explanation Bench · An in-memory design instrument, not a production policy engine.</footer>
  `

  document.querySelectorAll<HTMLElement>('[data-principal]').forEach(el => el.onclick = () => { selectedPrincipal = el.dataset.principal as PrincipalId; selectedScopeTarget = principals.find(p => p.id === selectedPrincipal)?.employeeId ?? 'EMP-005'; lastDecision = null; render() })
  document.querySelectorAll<HTMLButtonElement>('[data-challenge]').forEach(el => el.onclick = () => loadChallenge(Number(el.dataset.challenge)))
  document.querySelector<HTMLSelectElement>('#permission')!.onchange = e => { selectedPermission = (e.target as HTMLSelectElement).value as Permission; lastDecision = null; render() }
  document.querySelector<HTMLSelectElement>('#request-permission')!.onchange = e => { selectedPermission = (e.target as HTMLSelectElement).value as Permission; lastDecision = null; render() }
  document.querySelector<HTMLSelectElement>('#scope')!.onchange = e => { selectedScope = (e.target as HTMLSelectElement).value as ScopeType; lastDecision = null; render() }
  document.querySelector<HTMLSelectElement>('#scope-target')?.addEventListener('change', e => { selectedScopeTarget = (e.target as HTMLSelectElement).value })
  document.querySelector<HTMLSelectElement>('#resource')!.onchange = e => { selectedResource = (e.target as HTMLSelectElement).value; lastDecision = null; render() }
  document.querySelector<HTMLButtonElement>('#add-grant')!.onclick = addGrant
  document.querySelector<HTMLButtonElement>('#evaluate')!.onclick = evaluate
  document.querySelectorAll<HTMLButtonElement>('[data-revoke]').forEach(el => el.onclick = () => { const grant = grants.find(g => g.id === Number(el.dataset.revoke)); if (grant) grant.valid = false; lastDecision = null; render() })
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
