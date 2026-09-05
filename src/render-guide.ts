import { marked, Renderer } from 'marked'

export function renderGuide(markdown: string, assets: Record<string, string>): string {
  const renderer = new Renderer()
  const renderLink = renderer.link
  renderer.link = function (token) {
    const fragmentAt = token.href.indexOf('#')
    const file = fragmentAt < 0 ? token.href : token.href.slice(0, fragmentAt)
    const fragment = fragmentAt < 0 ? '' : token.href.slice(fragmentAt)
    const href = assets[file] ? assets[file] + fragment : token.href
    return renderLink.call(this, { ...token, href })
  }
  return marked.parse(markdown, { renderer, async: false })
}
