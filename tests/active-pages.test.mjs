import assert from 'node:assert/strict'
import test, { after } from 'node:test'
import { build, createServer } from 'vite'

// Catch retired HTML entry points accidentally shipping as active pages again.
test('production ships only the homepage and handbook entry points', async () => {
  const bundle = await build({ logLevel: 'silent', build: { write: false } })
  const outputs = (Array.isArray(bundle) ? bundle : [bundle]).flatMap(result => result.output)
  const pages = outputs.map(output => output.fileName).filter(name => name.endsWith('.html')).sort()
  assert.deepEqual(pages, ['concept.html', 'index.html'])
})

test('homepage loads the same handbook application as the concept page', async () => {
  const bundle = await build({ logLevel: 'silent', build: { write: false } })
  const outputs = (Array.isArray(bundle) ? bundle : [bundle]).flatMap(result => result.output)
  const scriptFor = name => {
    const page = outputs.find(output => output.fileName === name)
    const script = String(page.source).match(/<script[^>]+src="([^"]+)"/)
    assert.ok(script, `Expected application script in ${name}`)
    return script[1]
  }
  assert.equal(scriptFor('index.html'), scriptFor('concept.html'))
  const modules = outputs.filter(output => output.type === 'chunk').flatMap(output => Object.keys(output.modules))
  assert.equal(modules.some(id => /\/src\/(main|scenario-pack)\.ts$/.test(id)), false)
})

// Catch chapter diagrams linked by the reader being absent from production.
test('production includes the linked ownership-lineage SVG', async () => {
  const bundle = await build({ logLevel: 'silent', build: { write: false } })
  const outputs = (Array.isArray(bundle) ? bundle : [bundle]).flatMap(result => result.output)
  const diagram = outputs.find(output => /^assets\/ownership-lineage-[^/]+\.svg$/.test(output.fileName))
  assert.ok(diagram, 'The reader-linked ownership diagram must be emitted')
  assert.ok(outputs.some(output => output.type === 'chunk' && output.code.includes(diagram.fileName)),
    'The application must reference the emitted diagram URL')
})

const server = await createServer({
  logLevel: 'silent',
  server: { host: '127.0.0.1', port: 0 },
})
after(() => server.close())
await server.listen()
const origin = server.resolvedUrls.local[0]

// An SPA fallback must not silently turn a retired URL into an active page.
for (const route of ['projects-explorer.html', 'enforcement-trace.html']) {
  test(`retired ${route} returns 404`, async () => {
    const response = await fetch(new URL(route, origin), { headers: { Accept: 'text/html' } })
    assert.equal(response.status, 404)
  })
}

for (const route of ['', 'concept.html?doc=projects']) {
  test(`retained page /${route} stays available`, async () => {
    const response = await fetch(new URL(route, origin), { headers: { Accept: 'text/html' } })
    assert.equal(response.status, 200)
  })
}
