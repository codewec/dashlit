import { get, writable } from 'svelte/store'
import { api, type User, type Dashboard, type SystemInfo } from './api'
import { iconSrc } from './icon-helpers'
import {
  customThemeColorVars,
  isLightResolvedTheme,
  loadGuestCustomTheme,
  loadGuestTheme,
  normalizeCustomTheme,
  normalizeTheme,
  parseCustomTheme,
  saveGuestTheme,
  type CustomThemeConfig,
  type ResolvedTheme,
  type Theme,
} from './themes'

export const user = writable<User | null>(null)
export const editMode = writable(false)
export const theme = writable<Theme>('system')
export const resolvedTheme = writable<ResolvedTheme>('frappe')
export const customTheme = writable<CustomThemeConfig>(loadGuestCustomTheme())
export const customThemeEditorOpen = writable(false)
export const currentDashboard = writable<Dashboard | null>(null)
export const searchQuery = writable('')
export const systemInfo = writable<SystemInfo | null>(null)

let selectedTheme: Theme = 'system'
let mediaListenerAttached = false
let activeCustom: CustomThemeConfig = loadGuestCustomTheme()
let persistTimer: ReturnType<typeof setTimeout> | undefined

function resolveTheme(mode: Theme): ResolvedTheme {
  if (mode !== 'system') return mode
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'frappe' : 'crema'
}

function clearCustomThemeStyles(root: HTMLElement) {
  for (const property of customThemeColorVars) {
    root.style.removeProperty(property)
  }
  root.style.removeProperty('--page-bg-image')
}

export function applyCustomThemeStyles(config: CustomThemeConfig) {
  const root = document.documentElement
  root.style.setProperty('--color-bg', config.bg)
  root.style.setProperty('--color-surface', config.surface)
  root.style.setProperty('--color-text', config.text)
  root.style.setProperty('--color-primary', config.primary)
  root.style.setProperty('--color-accent', config.accent)
  root.style.setProperty('--color-danger', config.danger)
  root.style.setProperty('--color-success', config.success)

  const imageSrc = config.backgroundImage ? iconSrc(config.backgroundImage) : ''
  if (imageSrc) {
    root.style.setProperty('--page-bg-image', `url(${JSON.stringify(imageSrc)})`)
  } else {
    root.style.removeProperty('--page-bg-image')
  }
}

export function applyTheme(mode: Theme) {
  selectedTheme = mode
  const root = document.documentElement
  const resolved = resolveTheme(mode)
  root.setAttribute('data-theme', resolved)

  if (resolved === 'custom') {
    applyCustomThemeStyles(activeCustom)
    root.style.colorScheme = activeCustom.scheme
  } else {
    clearCustomThemeStyles(root)
    root.style.colorScheme = isLightResolvedTheme(resolved) ? 'light' : 'dark'
  }

  resolvedTheme.set(resolved)
  saveGuestTheme(mode, activeCustom)

  if (!mediaListenerAttached) {
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
      if (selectedTheme === 'system') applyTheme('system')
    })
    mediaListenerAttached = true
  }
}

async function persistThemeToServer(mode: Theme, config: CustomThemeConfig) {
  if (!get(user)) return
  try {
    const updated = await api.updateTheme({
      theme: mode,
      customTheme: mode === 'custom' ? config : null,
    })
    user.set(updated)
  } catch {
    // Keep the local selection even if the network request fails.
  }
}

function schedulePersist(mode: Theme, config: CustomThemeConfig) {
  clearTimeout(persistTimer)
  persistTimer = setTimeout(() => {
    void persistThemeToServer(mode, config)
  }, 250)
}

export function setTheme(mode: Theme, options?: { persist?: boolean }) {
  theme.set(mode)
  applyTheme(mode)
  if (options?.persist !== false) schedulePersist(mode, activeCustom)
}

export function setCustomTheme(config: CustomThemeConfig, options?: { persist?: boolean; apply?: boolean }) {
  const next = normalizeCustomTheme(config)
  activeCustom = next
  customTheme.set(next)
  saveGuestTheme(selectedTheme, next)
  if (options?.apply !== false && selectedTheme === 'custom') applyTheme('custom')
  if (options?.persist !== false) schedulePersist(selectedTheme, next)
}

export function openCustomThemeEditor() {
  customThemeEditorOpen.set(true)
}

export function selectCustomTheme() {
  setTheme('custom')
  openCustomThemeEditor()
}

export function hydrateThemeFromUser(nextUser: User | null) {
  if (!nextUser) {
    const guestTheme = loadGuestTheme()
    activeCustom = loadGuestCustomTheme()
    customTheme.set(activeCustom)
    theme.set(guestTheme)
    applyTheme(guestTheme)
    return
  }

  const nextTheme = normalizeTheme(nextUser.theme)
  const nextCustom = parseCustomTheme(nextUser.customTheme)
  activeCustom = nextCustom
  customTheme.set(nextCustom)
  theme.set(nextTheme)
  applyTheme(nextTheme)
}
