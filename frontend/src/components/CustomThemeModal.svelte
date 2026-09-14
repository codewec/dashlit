<script lang="ts">
  import Modal from './Modal.svelte'
  import ImageSourceField from './ImageSourceField.svelte'
  import {
    customTheme,
    customThemeEditorOpen,
    deleteCustomTheme,
    hasCustomTheme,
    resolvedTheme,
    setCustomTheme,
    setTheme,
    theme,
  } from '../lib/stores'
  import { defaultCustomTheme, isLightResolvedTheme, normalizeCustomTheme, type CustomThemeConfig } from '../lib/themes'
  import { toastError } from '../lib/toasts'

  let open = $state(false)
  let draft = $state<CustomThemeConfig>({ ...defaultCustomTheme })
  let snapshot = $state<CustomThemeConfig>({ ...defaultCustomTheme })
  let applied = $state(false)
  let backgroundImage = $state('')
  let initializedForOpen = $state(false)
  let previousTheme = $state($theme)

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
    previousTheme = $theme
    const styles = getComputedStyle(document.documentElement)
    const current: CustomThemeConfig = $hasCustomTheme
      ? { ...$customTheme }
      : normalizeCustomTheme({
          scheme: isLightResolvedTheme($resolvedTheme) ? 'light' : 'dark',
          bg: styles.getPropertyValue('--color-bg').trim(),
          surface: styles.getPropertyValue('--color-surface').trim(),
          text: styles.getPropertyValue('--color-text').trim(),
          primary: styles.getPropertyValue('--color-primary').trim(),
          accent: styles.getPropertyValue('--color-accent').trim(),
          danger: styles.getPropertyValue('--color-danger').trim(),
          success: styles.getPropertyValue('--color-success').trim(),
          backgroundImage: '',
        })
    snapshot = current
    draft = { ...current }
    backgroundImage = current.backgroundImage
    // Prepare the preview palette before switching to custom, otherwise the
    // default custom palette is briefly applied while the editor opens.
    setCustomTheme(current, { persist: false, apply: false })
    setTheme('custom', { persist: false })
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
    if (!applied) {
      setCustomTheme(snapshot, { persist: false, apply: false })
      setTheme(previousTheme, { persist: false })
    }
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
    setCustomTheme(next, { persist: true, apply: true, exists: true })
    setTheme('custom', { persist: true })
    open = false
    customThemeEditorOpen.set(false)
  }

  async function removeTheme() {
    try {
      await deleteCustomTheme()
      applied = true
      open = false
      customThemeEditorOpen.set(false)
    } catch (error: unknown) {
      toastError(error, 'Could not delete theme')
    }
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
          <span class="relative h-9 w-9 shrink-0 overflow-hidden rounded-lg" style:background-color={draft[field.key]}>
            <input
              type="color"
              class="absolute inset-0 h-full w-full cursor-pointer opacity-0"
              value={draft[field.key]}
              aria-label={field.label}
              oninput={(e) => onColorInput(field.key, (e.currentTarget as HTMLInputElement).value)}
            />
          </span>
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
    {#if $hasCustomTheme}
      <button type="button" class="mr-auto rounded-btn px-3 py-2 text-sm text-danger hover:bg-danger-soft" onclick={removeTheme}>Delete theme</button>
    {/if}
    <button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={cancel}>Cancel</button>
    <button type="button" class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-primary-fg hover:bg-primary-hover" onclick={apply}
      >Apply</button
    >
  {/snippet}
</Modal>
