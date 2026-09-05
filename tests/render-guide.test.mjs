import assert from 'node:assert/strict'
import test, { after } from 'node:test'
import { createServer } from 'vite'

const server = await createServer({ configFile: false, server: { middlewareMode: true } })
after(() => server.close())
const { renderGuide } = await server.ssrLoadModule('/src/render-guide.ts')

// Catch a reader that renders repository-relative links unchanged: those links
// work in the checkout but point at absent files in the production build.
test('reader links use emitted local chapter assets and retain fragments', () => {
  const html = renderGuide('[Policy](../../docs/endpoint-policy-format.md#put-example)', {
    '../../docs/endpoint-policy-format.md': '/assets/endpoint-policy-format-123.md',
  })
  assert.match(html, /href="\/assets\/endpoint-policy-format-123\.md#put-example"/)
  assert.match(html, />Policy<\/a>/)
})

test('reader preserves external links and in-page navigation', () => {
  const html = renderGuide('[Reference](https://example.org/read#one) [Here](#section)', {})
  assert.match(html, /href="https:\/\/example.org\/read#one"/)
  assert.match(html, /href="#section"/)
})

test('reader links to preserved original files using emitted archive assets', () => {
  const html = renderGuide('[Original](../../docs/history/original.md.txt)', {
    '../../docs/history/original.md.txt': '/assets/original-456.md.txt',
  })
  assert.match(html, /href="\/assets\/original-456\.md\.txt"/)
})
