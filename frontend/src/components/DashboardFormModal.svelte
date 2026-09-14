<script lang="ts">
  import { Switch } from 'bits-ui'
  import type { DashboardForm } from '../lib/dashboard-helpers'
  import Modal from './Modal.svelte'
  import IconField from './IconField.svelte'
  import { formatHotkey, formatModifierPreview, hotkeyFromEvent, hotkeysEquivalent } from '../lib/hotkeys'

  let {
    open = $bindable(false),
    form = $bindable(),
    canSetDefault = false,
    conflictingHotkeys = [],
    onSave,
  }: {
    open?: boolean
    form: DashboardForm
    canSetDefault?: boolean
    conflictingHotkeys?: string[]
    onSave: () => void | Promise<void>
  } = $props()

  let hotkeyError = $state('')
  let hotkeyPreview = $state('')
  const hotkeyConflict = $derived(!!form.hotkey && conflictingHotkeys.some((hotkey) => hotkeysEquivalent(form.hotkey, hotkey)))

  $effect(() => {
    if (!open) {
      hotkeyError = ''
      hotkeyPreview = ''
    }
  })

  function captureHotkey(event: KeyboardEvent) {
    event.preventDefault()
    event.stopPropagation()
    if (event.key === 'Escape') {
      ;(event.currentTarget as HTMLInputElement).blur()
      hotkeyError = ''
      hotkeyPreview = ''
      return
    }
    if (event.key === 'Backspace' || event.key === 'Delete') {
      form.hotkey = ''
      form.hotkeyLabel = ''
      hotkeyError = ''
      return
    }
    const hotkey = hotkeyFromEvent(event)
    if (!hotkey) {
      hotkeyPreview = formatModifierPreview(event)
      if (!['Control', 'Alt', 'Shift', 'Meta'].includes(event.key)) hotkeyError = 'Use at least one modifier key'
      return
    }
    form.hotkey = hotkey
    form.hotkeyLabel = event.key
    hotkeyPreview = ''
    hotkeyError = ''
  }

  function releaseHotkeyModifier(event: KeyboardEvent) {
    if (['Control', 'Alt', 'Shift', 'Meta'].includes(event.key)) hotkeyPreview = formatModifierPreview(event)
  }

  async function submit(e: Event) {
    e.preventDefault()
    if (!form.name.trim() || !form.slug.trim() || hotkeyConflict) return
    await onSave()
  }
</script>

<Modal bind:open title={form.creating ? 'New dashboard' : 'Dashboard settings'}>
  <form class="space-y-3" onsubmit={submit}>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Name</span>
      <input class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm" bind:value={form.name} required />
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Slug</span>
      <input class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm" bind:value={form.slug} required />
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Description</span>
      <input class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm" bind:value={form.description} />
    </label>
    <div>
      <span class="mb-1 block text-xs text-text-muted">Icon</span>
      <IconField bind:value={form.icon} bind:valueDark={form.iconDark} />
    </div>

    <div class="hidden sm:block">
      <span class="mb-1 block text-xs text-text-muted">Hotkey</span>
      <div class="flex gap-2">
        <input
          readonly
          value={hotkeyPreview || formatHotkey(form.hotkey, form.hotkeyLabel)}
          placeholder="Focus and press a combination"
          aria-label="Dashboard hotkey"
          class="min-w-0 flex-1 cursor-default rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none placeholder:text-text-subtle focus:border-primary"
          onkeydown={captureHotkey}
          onkeyup={releaseHotkeyModifier}
          onblur={() => (hotkeyPreview = '')}
        />
        {#if form.hotkey}
          <button
            type="button"
            class="rounded-btn border border-border px-3 text-sm text-text-muted hover:bg-surface-2 hover:text-text"
            onclick={() => {
              form.hotkey = ''
              form.hotkeyLabel = ''
              hotkeyError = ''
            }}>Clear</button
          >
        {/if}
      </div>
      {#if hotkeyConflict}
        <span class="mt-1 block text-xs text-danger">This hotkey is already assigned to an item or dashboard. Choose another combination.</span>
      {/if}
      <span class="mt-1 block text-xs {hotkeyError ? 'text-danger' : 'text-text-subtle'}">
        {hotkeyError || 'Press Ctrl, Alt, Shift or Meta together with another key. Backspace clears it.'}
      </span>
      <span class="mt-1 block text-xs text-text-subtle">
        Some shortcuts may be intercepted by your operating system or browser, especially combinations using Meta.
      </span>
    </div>

    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <label class="block">
        <span class="mb-1 block text-xs text-text-muted">Privacy</span>
        <select class="h-10 w-full rounded-btn border border-border bg-bg-elevated px-3 text-sm" bind:value={form.privacy}>
          <option value="public">Public</option>
          <option value="users">Authenticated users</option>
          <option value="private">Private</option>
        </select>
      </label>
      <label class="block">
        <span class="mb-1 block text-xs text-text-muted">Layout</span>
        <select class="h-10 w-full rounded-btn border border-border bg-bg-elevated px-3 text-sm" bind:value={form.layout}>
          <option value="rows">Rows</option>
          <option value="columns">Columns</option>
          <option value="masonry">Masonry</option>
        </select>
      </label>
      <div class="flex items-center justify-between gap-3 sm:block">
        <span class="text-sm text-text sm:mb-1 sm:block sm:text-xs sm:text-text-muted">Clean mode</span>
        <div class="sm:flex sm:h-10 sm:items-center sm:rounded-btn sm:border sm:border-border sm:bg-bg-elevated sm:px-3">
          <Switch.Root
            checked={form.cleanMode}
            onCheckedChange={(v) => (form.cleanMode = !!v)}
            class="peer inline-flex h-7 w-12 shrink-0 cursor-pointer items-center rounded-full border border-transparent bg-border transition data-[state=checked]:bg-primary sm:h-5 sm:w-9"
          >
            <Switch.Thumb
              class="pointer-events-none block h-6 w-6 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-[1.35rem] sm:h-4 sm:w-4 sm:data-[state=checked]:translate-x-[1.1rem]"
            />
          </Switch.Root>
        </div>
      </div>
      <div class="flex items-center justify-between gap-3 sm:block">
        <span class="text-sm text-text sm:mb-1 sm:block sm:text-xs sm:text-text-muted">Wide mode</span>
        <div class="sm:flex sm:h-10 sm:items-center sm:rounded-btn sm:border sm:border-border sm:bg-bg-elevated sm:px-3">
          <Switch.Root
            checked={form.width === 'wide'}
            onCheckedChange={(v) => (form.width = v ? 'wide' : 'default')}
            class="peer inline-flex h-7 w-12 shrink-0 cursor-pointer items-center rounded-full border border-transparent bg-border transition data-[state=checked]:bg-primary sm:h-5 sm:w-9"
          >
            <Switch.Thumb
              class="pointer-events-none block h-6 w-6 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-[1.35rem] sm:h-4 sm:w-4 sm:data-[state=checked]:translate-x-[1.1rem]"
            />
          </Switch.Root>
        </div>
      </div>
      {#if canSetDefault}
        <div class="flex items-center justify-between gap-3 sm:block">
          <span class="text-sm text-text sm:mb-1 sm:block sm:text-xs sm:text-text-muted">Default dashboard</span>
          <div class="sm:flex sm:h-10 sm:items-center sm:rounded-btn sm:border sm:border-border sm:bg-bg-elevated sm:px-3">
            <Switch.Root
              checked={form.isDefault}
              onCheckedChange={(v) => (form.isDefault = !!v)}
              class="peer inline-flex h-7 w-12 shrink-0 cursor-pointer items-center rounded-full border border-transparent bg-border transition data-[state=checked]:bg-primary sm:h-5 sm:w-9"
            >
              <Switch.Thumb
                class="pointer-events-none block h-6 w-6 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-[1.35rem] sm:h-4 sm:w-4 sm:data-[state=checked]:translate-x-[1.1rem]"
              />
            </Switch.Root>
          </div>
        </div>
      {/if}
    </div>

    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={() => (open = false)}>Cancel</button>
      <button
        type="submit"
        disabled={hotkeyConflict}
        class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:opacity-50">Save</button
      >
    </div>
  </form>
</Modal>
