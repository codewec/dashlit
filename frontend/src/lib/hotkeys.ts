export type ModifierState = {
  ctrl: boolean
  alt: boolean
  shift: boolean
  meta: boolean
}

const modifierKeys = new Set(['Control', 'Alt', 'Shift', 'Meta'])
const physicalCodePattern = /^(?:Key[A-Z]|Digit[0-9]|Numpad\w+|F(?:[1-9]|1[0-9]|2[0-4])|Space|Minus|Equal|Slash|Backslash|BracketLeft|BracketRight|Semicolon|Quote|Comma|Period|Backquote|Enter|Tab|ArrowUp|ArrowDown|ArrowLeft|ArrowRight|Home|End|PageUp|PageDown|Insert|Delete)$/

function parseHotkey(hotkey: string) {
  const parts = hotkey.split('+').map((part) => part.trim()).filter(Boolean)
  return {
    modifiers: new Set(parts.slice(0, -1).map((part) => part.toLowerCase())),
    key: parts.at(-1) || '',
  }
}

function isPhysicalCode(key: string): boolean {
  return physicalCodePattern.test(key)
}

function normalizeLegacyKey(key: string): string {
  if (key === ' ') return 'SPACE'
  if (key.length === 1) return key.toLocaleUpperCase()
  return key.toLocaleUpperCase()
}

function codeLabel(code: string): string {
  if (/^Key[A-Z]$/.test(code)) return code.slice(3)
  if (/^Digit[0-9]$/.test(code)) return code.slice(5)
  if (/^Numpad[0-9]$/.test(code)) return `Num ${code.slice(-1)}`
  const labels: Record<string, string> = {
    Space: 'Space',
    Minus: '-',
    Equal: '+',
    Slash: '/',
    Backslash: '\\',
    BracketLeft: '[',
    BracketRight: ']',
    Semicolon: ';',
    Quote: "'",
    Comma: ',',
    Period: '.',
    Backquote: '`',
  }
  return labels[code] || code
}

function modifiersMatch(event: KeyboardEvent, modifiers: Set<string>): boolean {
  return (
    event.ctrlKey === modifiers.has('ctrl') &&
    event.altKey === modifiers.has('alt') &&
    event.shiftKey === modifiers.has('shift') &&
    event.metaKey === modifiers.has('meta')
  )
}

function modifiersEquivalent(first: Set<string>, second: Set<string>): boolean {
  return ['ctrl', 'alt', 'shift', 'meta'].every((modifier) => first.has(modifier) === second.has(modifier))
}

export function hotkeyFromEvent(event: KeyboardEvent): string | null {
  if (modifierKeys.has(event.key) || event.key === 'Dead' || event.key === 'Unidentified' || !event.code) return null
  const modifiers: string[] = []
  if (event.ctrlKey) modifiers.push('Ctrl')
  if (event.altKey) modifiers.push('Alt')
  if (event.shiftKey) modifiers.push('Shift')
  if (event.metaKey) modifiers.push('Meta')
  if (modifiers.length === 0) return null
  return [...modifiers, event.code].join('+')
}

export function hotkeyMatches(event: KeyboardEvent, hotkey: string): boolean {
  const parsed = parseHotkey(hotkey)
  if (!parsed.key || !modifiersMatch(event, parsed.modifiers)) return false

  // New hotkeys are layout-independent and match the physical keyboard code.
  if (isPhysicalCode(parsed.key)) return event.code === parsed.key

  // Legacy hotkeys did not save event.code. They can only be matched by the
  // character in the active layout without making locale-specific guesses.
  return normalizeLegacyKey(event.key) === normalizeLegacyKey(parsed.key)
}

export function hotkeysEquivalent(first: string, second: string): boolean {
  const a = parseHotkey(first)
  const b = parseHotkey(second)
  if (!modifiersEquivalent(a.modifiers, b.modifiers)) return false
  if (isPhysicalCode(a.key) && isPhysicalCode(b.key)) return a.key === b.key
  if (!isPhysicalCode(a.key) && !isPhysicalCode(b.key)) return normalizeLegacyKey(a.key) === normalizeLegacyKey(b.key)
  return false
}

export function hotkeyKeyLabel(hotkey: string): string {
  const key = parseHotkey(hotkey).key
  return isPhysicalCode(key) ? codeLabel(key) : key
}

export function formatHotkey(hotkey: string): string {
  if (!hotkey) return ''
  const parsed = parseHotkey(hotkey)
  const modifiers = ['Ctrl', 'Alt', 'Shift', 'Meta'].filter((modifier) => parsed.modifiers.has(modifier.toLowerCase()))
  return [...modifiers, hotkeyKeyLabel(hotkey)].join(' + ')
}

export function modifierStateFromEvent(event: KeyboardEvent): ModifierState {
  return { ctrl: event.ctrlKey, alt: event.altKey, shift: event.shiftKey, meta: event.metaKey }
}

export function formatModifierPreview(event: KeyboardEvent): string {
  const modifiers: string[] = []
  if (event.ctrlKey) modifiers.push('Ctrl')
  if (event.altKey) modifiers.push('Alt')
  if (event.shiftKey) modifiers.push('Shift')
  if (event.metaKey) modifiers.push('Meta')
  return modifiers.length ? `${modifiers.join(' + ')} + …` : ''
}

export function shouldShowHotkey(hotkey: string, held: ModifierState): boolean {
  if (!hotkey) return false
  const { modifiers } = parseHotkey(hotkey)
  const hasModifier = held.ctrl || held.alt || held.shift || held.meta
  if (!hasModifier) return false
  return (
    held.ctrl === modifiers.has('ctrl') &&
    held.alt === modifiers.has('alt') &&
    held.shift === modifiers.has('shift') &&
    held.meta === modifiers.has('meta')
  )
}
