import { writable } from 'svelte/store'

function normalizePath(pathname: string): string {
  if (!pathname || pathname === '/') return '/'
  return pathname.replace(/\/+$/, '') || '/'
}

function migrateLegacyHash(): string {
  if (!window.location.hash.startsWith('#/') || window.location.hash.startsWith('#//')) return normalizePath(window.location.pathname)

  const legacyURL = window.location.hash.slice(1)
  history.replaceState(history.state, '', legacyURL)
  return normalizePath(window.location.pathname)
}

export const routePath = writable(migrateLegacyHash())

function navigate(to: string, replaceState: boolean) {
  const url = new URL(to, window.location.origin)
  if (url.origin !== window.location.origin) {
    window.location.assign(url)
    return
  }

  const nextURL = `${url.pathname}${url.search}${url.hash}`
  if (replaceState) history.replaceState(history.state, '', nextURL)
  else history.pushState(history.state, '', nextURL)
  routePath.set(normalizePath(url.pathname))
  window.scrollTo({ top: 0, left: 0 })
}

export function push(to: string) {
  navigate(to, false)
}

export function replace(to: string) {
  navigate(to, true)
}

export function startRouter() {
  const handlePopState = () => routePath.set(normalizePath(window.location.pathname))
  const handleClick = (event: MouseEvent) => {
    if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return

    const target = event.target
    const anchor = target instanceof Element ? target.closest<HTMLAnchorElement>('a[href]') : null
    if (!anchor || anchor.target || anchor.download || anchor.hasAttribute('data-native-navigation')) return

    const url = new URL(anchor.href, window.location.href)
    if (url.origin !== window.location.origin || url.pathname.startsWith('/api/')) return

    event.preventDefault()
    push(`${url.pathname}${url.search}${url.hash}`)
  }

  window.addEventListener('popstate', handlePopState)
  document.addEventListener('click', handleClick)
  return () => {
    window.removeEventListener('popstate', handlePopState)
    document.removeEventListener('click', handleClick)
  }
}
