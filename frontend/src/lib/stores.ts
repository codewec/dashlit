import { writable } from 'svelte/store'
import type { User, Dashboard, SystemInfo } from './api'
import type { ResolvedTheme, Theme } from './themes'

export const user = writable<User | null>(null)
export const editMode = writable(false)
export const theme = writable<Theme>('system')
export const resolvedTheme = writable<ResolvedTheme>('frappe')
export const currentDashboard = writable<Dashboard | null>(null)
export const searchQuery = writable('')
export const systemInfo = writable<SystemInfo | null>(null)

let selectedTheme: Theme = 'system'
let mediaListenerAttached = false

function resolveTheme(mode: Theme): ResolvedTheme {
  if (mode !== 'system') return mode
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'frappe' : 'crema'
}

export function applyTheme(mode: Theme) {
  selectedTheme = mode
  const root = document.documentElement
  const resolved = resolveTheme(mode)
  root.setAttribute('data-theme', resolved)
  root.style.colorScheme = resolved === 'crema' || resolved === 'latte' ? 'light' : 'dark'
  resolvedTheme.set(resolved)
  localStorage.setItem('bd_theme', mode)

  if (!mediaListenerAttached) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (selectedTheme === 'system') applyTheme('system')
    })
    mediaListenerAttached = true
  }
}

export function setTheme(mode: Theme) {
  theme.set(mode)
  applyTheme(mode)
}
