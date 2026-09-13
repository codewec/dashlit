<script lang="ts">
  import Modal from './Modal.svelte'
  import ImageSourceField from './ImageSourceField.svelte'
  import { customTheme, customThemeEditorOpen, setCustomTheme, setTheme } from '../lib/stores'
  import { defaultCustomTheme, type CustomThemeConfig } from '../lib/themes'

  let open = $state(false)
  let draft = $state<CustomThemeConfig>({ ...defaultCustomTheme })
  let snapshot = $state<CustomThemeConfig>({ ...defaultCustomTheme })
  let applied = $state(false)
  let backgroundImage = $state('')
  let initializedForOpen = $state(false)

  const colorFields: {
    key: keyof Pick<CustomThemeConfig, 'bg' | 'surface' | 'text' | 'primary' | 'accent' | 'danger' | 'success'>
    label: string
  }[] = [
    { key: 'bg', label: 'Background' },
    { key: 'surface', label: 'Tiles' },
    { key: 'text', label: 'Text' },
    { key: 'primary', label: 'Primary' },
    { key: 'accent', label: 'Accent' },
    { key: 'danger', label: 'Danger' },
    { key: 'success', label: 'Success' },
  ]

  $effect(() => {
    open = $customThemeEditorOpen
  })

  $effect(() => {
    if (!open) {
      initializedForOpen = false
      return
    }
    if (initializedForOpen) return
    initializedForOpen = true
    applied = false
    const current = { ...$customTheme }
    snapshot = current
    draft = { ...current }
    backgroundImage = current.backgroundImage
    setTheme('custom', { persist: false })
    setCustomTheme(current, { persist: false, apply: true })
  })

  $effect(() => {
    if (!open) return
    if (draft.backgroundImage === backgroundImage) return
    draft = { ...draft, backgroundImage }
    setCustomTheme(draft, { persist: false, apply: true })
  })

  function preview(next: CustomThemeConfig) {
    draft = next
    setCustomTheme(next, { persist: false, apply: true })
  }

  function onColorInput(key: (typeof colorFields)[number]['key'], value: string) {
    preview({ ...draft, [key]: value })
  }

  function onSchemeChange(scheme: 'light' | 'dark') {
    preview({ ...draft, scheme })
  }

  function revertIfNeeded() {
    if (!applied) setCustomTheme(snapshot, { persist: false, apply: true })
  }

  function cancel() {
    revertIfNeeded()
    applied = false
    open = false
    customThemeEditorOpen.set(false)
  }

  function apply() {
    applied = true
    const next = { ...draft, backgroundImage }
    setCustomTheme(next, { persist: true, apply: true })
    setTheme('custom', { persist: true })
    open = false
    customThemeEditorOpen.set(false)
  }

  function handleOpenChange(next: boolean) {
    if (!next) revertIfNeeded()
    open = next
    customThemeEditorOpen.set(next)
  }
</script>

<Modal
  bind:open
  title="Custom theme"
  description="Pick colors and an optional background image. Changes preview live and save to your account when applied."
  class="w-[min(32rem,calc(100vw-1.5rem))]"
  onOpenChange={handleOpenChange}
>
  <div class="space-y-4">
    <div class="space-y-2">
      <span class="text-sm font-medium text-text">Appearance</span>
      <div class="flex gap-1 rounded-lg bg-surface-2 p-1">
        {#each ['dark', 'light'] as const as scheme}
          <button
            type="button"
            class="flex-1 rounded-md px-2 py-1.5 text-xs font-medium transition
              {draft.scheme === scheme ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'}"
            onclick={() => onSchemeChange(scheme)}
          >
            {scheme === 'dark' ? 'Dark' : 'Light'}
          </button>
        {/each}
      </div>
      <p class="text-xs text-text-subtle">Controls icon variants and browser form controls.</p>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      {#each colorFields as field}
        <label class="flex items-center gap-3 rounded-xl border border-border-soft bg-bg-elevated/60 px-3 py-2.5">
          <input
            type="color"
            class="h-9 w-9 shrink-0 cursor-pointer rounded-lg border border-border bg-transparent p-0.5"
            value={draft[field.key]}
            oninput={(e) => onColorInput(field.key, (e.currentTarget as HTMLInputElement).value)}
          />
          <span class="min-w-0 flex-1">
            <span class="block text-sm font-medium text-text">{field.label}</span>
            <span class="block truncate font-mono text-[11px] text-text-subtle">{draft[field.key]}</span>
          </span>
        </label>
      {/each}
    </div>

    <ImageSourceField bind:value={backgroundImage} />
  </div>

  {#snippet footer()}
    <button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={cancel}>Cancel</button>
    <button type="button" class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-primary-fg hover:bg-primary-hover" onclick={apply}
      >Apply</button
    >
  {/snippet}
</Modal>
