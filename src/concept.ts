import { marked } from 'marked'
import conceptGuide from './content/authorization-concept.md?raw'
import hrmsGuide from './content/hrms-tenant-setup.md?raw'
import projectsGuide from './content/projects-repositories-teams.md?raw'
import './styles.css'

const app = document.querySelector<HTMLDivElement>('#concept-app')!
const requestedDocument = new URLSearchParams(window.location.search).get('doc') ?? 'concept'
const documents = {
  concept: {
    label: 'Concept model',
    title: 'The complete authorization model',
    description: 'Permission, scope, target context, resource depth, assignment, enforcement, and audit.',
    filename: 'authorization-concept.md',
    markdown: conceptGuide,
  },
  hrms: {
    label: 'HRMS example',
    title: 'Tenant configures Payroll Administration',
    description: 'Departments, employees, roles, permissions, assignment scopes, schema records, and request consumption.',
    filename: 'hrms-tenant-setup.md',
    markdown: hrmsGuide,
  },
  projects: {
    label: 'Projects & repos',
    title: 'Projects, repositories, teams, and members',
    description: 'A second domain that tests groups, exact resources, subtree scopes, membership, and containment.',
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
    <nav class="page-tabs"><a class="${documentKey === 'concept' ? 'active' : ''}" href="${documentLink('concept')}">Concept</a><a class="${documentKey === 'hrms' ? 'active' : ''}" href="${documentLink('hrms')}">HRMS</a><a class="${documentKey === 'projects' ? 'active' : ''}" href="${documentLink('projects')}">Projects & repos</a><a href="/">Request explorer</a></nav>
    <div class="canonical"><span>Source</span><code>${currentDocument.filename}</code></div>
  </header>
  <main class="concept-main">
    <aside class="concept-aside">
      <small>${currentDocument.label.toUpperCase()}</small>
      <h2>${currentDocument.title}</h2>
      <p>${currentDocument.description}</p>
      <p class="source-note">Rendered directly from <code>${currentDocument.filename}</code>.</p>
      <a href="/">Open request explorer →</a>
    </aside>
    <article class="markdown-body">${marked.parse(currentDocument.markdown)}</article>
  </main>
  <footer>Authorization Explanation Bench · Documents rendered directly from Markdown.</footer>
`
