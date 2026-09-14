export type Theme = 'system' | 'crema' | 'latte' | 'frappe' | 'macchiato' | 'mocha' | 'black' | 'custom'
export type ResolvedTheme = Exclude<Theme, 'system'>

export type ThemeOption = {
  value: Theme
  label: string
  mode: 'system' | 'light' | 'dark' | 'custom'
  swatch: string
}

export type CustomThemeConfig = {
  scheme: 'light' | 'dark'
  bg: string
  surface: string
  text: string
  primary: string
  accent: string
  danger: string
  success: string
  backgroundImage: string
}

export const THEME_STORAGE_KEY = 'bd_theme'
export const CUSTOM_THEME_STORAGE_KEY = 'bd_custom_theme'
export const CUSTOM_THEME_EXISTS_STORAGE_KEY = 'bd_custom_theme_exists'

export const defaultCustomTheme: CustomThemeConfig = {
  scheme: 'dark',
  bg: '#000000',
  surface: '#1a1a1a',
  text: '#e8e8e8',
  primary: '#8aadf4',
  accent: '#c6a0f6',
  danger: '#ed8796',
  success: '#a6da95',
  backgroundImage: '',
}

/** CSS custom properties set from the color pickers (derived tokens live in CSS). */
export const customThemeColorVars = [
  '--color-bg',
  '--color-surface',
  '--color-text',
  '--color-primary',
  '--color-accent',
  '--color-danger',
  '--color-success',
] as const

export const themeOptions: ThemeOption[] = [
  { value: 'system', label: 'System', mode: 'system', swatch: 'linear-gradient(135deg, #f3f1ea 50%, #303446 50%)' },
  { value: 'crema', label: 'Crema', mode: 'light', swatch: '#f3f1ea' },
  { value: 'latte', label: 'Latte', mode: 'light', swatch: '#eff1f5' },
  { value: 'frappe', label: 'Frappé', mode: 'dark', swatch: '#303446' },
  { value: 'macchiato', label: 'Macchiato', mode: 'dark', swatch: '#24273a' },
  { value: 'mocha', label: 'Mocha', mode: 'dark', swatch: '#1e1e2e' },
  { value: 'custom', label: 'Custom', mode: 'custom', swatch: 'conic-gradient(from 90deg, #8aadf4, #c6a0f6, #ed8796, #a6da95, #8aadf4)' },
  { value: 'black', label: 'Black', mode: 'dark', swatch: '#000000' },
]

export function normalizeTheme(value: string | null | undefined): Theme {
  if (themeOptions.some((option) => option.value === value)) return value as Theme
  // Migrate values saved by the old light/dark toggle.
  if (value === 'light') return 'crema'
  if (value === 'dark') return 'frappe'
  return 'system'
}

export function themeLabel(value: Theme): string {
  return themeOptions.find((option) => option.value === value)?.label ?? value
}

export function isLightResolvedTheme(resolved: ResolvedTheme, customScheme: 'light' | 'dark' = 'dark'): boolean {
  if (resolved === 'custom') return customScheme === 'light'
  return resolved === 'crema' || resolved === 'latte'
}

function isHexColor(value: unknown): value is string {
  return typeof value === 'string' && /^#[0-9a-fA-F]{6}$/.test(value)
}

export function normalizeCustomTheme(raw: unknown): CustomThemeConfig {
  if (!raw || typeof raw !== 'object') return { ...defaultCustomTheme }
  const value = raw as Partial<CustomThemeConfig>
  return {
    scheme: value.scheme === 'light' ? 'light' : 'dark',
    bg: isHexColor(value.bg) ? value.bg : defaultCustomTheme.bg,
    surface: isHexColor(value.surface) ? value.surface : defaultCustomTheme.surface,
    text: isHexColor(value.text) ? value.text : defaultCustomTheme.text,
    primary: isHexColor(value.primary) ? value.primary : defaultCustomTheme.primary,
    accent: isHexColor(value.accent) ? value.accent : defaultCustomTheme.accent,
    danger: isHexColor(value.danger) ? value.danger : defaultCustomTheme.danger,
    success: isHexColor(value.success) ? value.success : defaultCustomTheme.success,
    backgroundImage: typeof value.backgroundImage === 'string' ? value.backgroundImage : '',
  }
}

export function parseCustomTheme(raw: string | null | undefined): CustomThemeConfig {
  if (!raw?.trim()) return { ...defaultCustomTheme }
  try {
    return normalizeCustomTheme(JSON.parse(raw))
  } catch {
    return { ...defaultCustomTheme }
  }
}

export function loadGuestTheme(): Theme {
  return normalizeTheme(localStorage.getItem(THEME_STORAGE_KEY))
}

export function loadGuestCustomTheme(): CustomThemeConfig {
  try {
    const raw = localStorage.getItem(CUSTOM_THEME_STORAGE_KEY)
    if (!raw) return { ...defaultCustomTheme }
    return normalizeCustomTheme(JSON.parse(raw))
  } catch {
    return { ...defaultCustomTheme }
  }
}

export function guestHasCustomTheme(): boolean {
  if (localStorage.getItem(CUSTOM_THEME_EXISTS_STORAGE_KEY) === 'true') return true
  return localStorage.getItem(THEME_STORAGE_KEY) === 'custom' && !!localStorage.getItem(CUSTOM_THEME_STORAGE_KEY)
}

export function saveGuestTheme(theme: Theme, custom: CustomThemeConfig, hasCustomTheme: boolean) {
  localStorage.setItem(THEME_STORAGE_KEY, theme)
  localStorage.setItem(CUSTOM_THEME_STORAGE_KEY, JSON.stringify(normalizeCustomTheme(custom)))
  if (hasCustomTheme) localStorage.setItem(CUSTOM_THEME_EXISTS_STORAGE_KEY, 'true')
  else localStorage.removeItem(CUSTOM_THEME_EXISTS_STORAGE_KEY)
}

export function customThemeSwatch(config: CustomThemeConfig): string {
  return `linear-gradient(135deg, ${config.bg} 40%, ${config.surface} 40%, ${config.surface} 70%, ${config.primary} 70%)`
}
