import { marked } from 'marked'
import guide from './content/authorization-concept.md?raw'
import './styles.css'

const app = document.querySelector<HTMLDivElement>('#concept-app')!

app.innerHTML = `
  <header class="topbar concept-topbar">
    <a class="brand" href="/"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10Z"/><path d="m9 12 2 2 4-4"/></svg><span>Permission Quest</span><small>Authorization design lab</small></a>
    <nav class="page-tabs"><a href="/">Playground</a><a class="active" href="/concept.html">Concept guide</a></nav>
    <div class="canonical"><span>Source</span><code>authorization-concept.md</code></div>
  </header>
  <main class="concept-main">
    <aside class="concept-aside">
      <small>READING GUIDE</small>
      <h2>The complete model</h2>
      <p>This page renders the repository’s Markdown directly. The document—not duplicated HTML—is the content source.</p>
      <a href="/">← Return to the game</a>
    </aside>
    <article class="markdown-body">${marked.parse(guide)}</article>
  </main>
  <footer>Permission Quest · Concept guide rendered from Markdown.</footer>
`
