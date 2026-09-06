import { renderGuide } from './render-guide'
import systemDiagram from '../docs/assets/authorization-system.svg?url'
import conceptGuide from './content/authorization-concept.md?raw'
import hrmsGuide from './content/hrms-tenant-setup.md?raw'
import projectsGuide from './content/projects-repositories-teams.md?raw'
import './styles.css'

// Emit chapter sources, preserved originals, and current diagrams for reader links.
const sourceAssets = import.meta.glob<string>(['../docs/**/*.{md,txt}', '../docs/assets/*.svg'], {
  eager: true, query: '?url', import: 'default',
})
const guideAssets = Object.fromEntries(
  Object.entries(sourceAssets).map(([path, url]) => [`../${path}`, url]),
)

const app = document.querySelector<HTMLDivElement>('#concept-app')!
const requestedDocument = new URLSearchParams(window.location.search).get('doc') ?? 'concept'
const documents = {
  concept: {
    label: 'Concept model',
    title: 'The Authorization Handbook',
    description: 'Reusable grants, assignments, dependent subgroups, and one endpoint-owned authorization gate.',
    filename: 'authorization-concept.md',
    markdown: conceptGuide,
  },
  hrms: {
    label: 'HRMS example',
    title: 'HRMS: bounded access',
    description: 'Human membership, self-scoped payslips, Finance grants, and declared path/body inputs.',
    filename: 'hrms-tenant-setup.md',
    markdown: hrmsGuide,
  },
  projects: {
    label: 'Projects & repos',
    title: 'Projects, repositories, teams, and members',
    description: 'Application-defined project boundaries, human teams, and constrained repository access.',
    filename: 'projects-repositories-teams.md',
    markdown: projectsGuide,
  },
} as const
type DocumentKey = keyof typeof documents
const documentKey: DocumentKey = requestedDocument in documents ? requestedDocument as DocumentKey : 'concept'
const currentDocument = documents[documentKey]

const documentLink = (key: DocumentKey) => `/concept.html?doc=${key}`

app.innerHTML = `
  <header class="topbar concept-topbar">
    <a class="brand" href="${documentLink('concept')}"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/></svg><span>Authorization Bench</span><small>Explain · inspect · decide</small></a>
    <nav class="page-tabs"><a class="${documentKey === 'concept' ? 'active' : ''}" href="${documentLink('concept')}">Concept</a><a class="${documentKey === 'hrms' ? 'active' : ''}" href="${documentLink('hrms')}">HRMS</a><a class="${documentKey === 'projects' ? 'active' : ''}" href="${documentLink('projects')}">Projects & repos</a></nav>
    <div class="canonical"><span>Source</span><code>${currentDocument.filename}</code></div>
  </header>
  <main class="concept-main">
    <aside class="concept-aside">
      <small>${currentDocument.label.toUpperCase()}</small>
      <h2>${currentDocument.title}</h2>
      <p>${currentDocument.description}</p>
      <p class="source-note">Rendered directly from <code>${currentDocument.filename}</code>.</p>
    </aside>
    <article class="markdown-body">
      <p class="reconciliation-notice">Working handbook · Q-090/Q-091: separate assignments and dependent subgroups. Baseline preserved as 0.0.1. Linked chapters open their local Markdown source.</p>
      <figure class="system-diagram">
        <a href="${systemDiagram}" aria-label="Open request-flow diagram at full size"><img src="${systemDiagram}" alt="Client request flows through authentication middleware and the endpoint handler. The embedded Auth Agent loads Auth authority and evaluates grants; the handler enforces a constrained database read and returns the response."></a>
        <figcaption>Request-flow SVG · <a href="${systemDiagram}">Open full size</a></figcaption>
      </figure>
      ${renderGuide(currentDocument.markdown, guideAssets)}
    </article>
  </main>
  <footer>Authorization Explanation Bench · Documents rendered directly from Markdown.</footer>
`
