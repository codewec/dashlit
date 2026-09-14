<script lang="ts">
  import { Switch } from 'bits-ui'
  import type { ItemForm } from '../lib/dashboard-helpers'
  import Modal from './Modal.svelte'
  import IconField from './IconField.svelte'
  import { formatHotkey, formatModifierPreview, hotkeyFromEvent, hotkeysEquivalent } from '../lib/hotkeys'

  let {
    open = $bindable(false),
    form = $bindable(),
    editing = false,
    itemHotkeys = [],
    dashboardHotkeys = [],
    onSave,
  }: {
    open?: boolean
    form: ItemForm
    editing?: boolean
    itemHotkeys?: string[]
    dashboardHotkeys?: string[]
    onSave: () => void | Promise<void>
  } = $props()

  let titleErr = $state(false)
  let urlErr = $state(false)
  let availabilityExpanded = $state(false)
  let hotkeyError = $state('')
  let hotkeyPreview = $state('')
  const itemHotkeyConflict = $derived(!!form.hotkey && itemHotkeys.some((hotkey) => hotkeysEquivalent(form.hotkey, hotkey)))
  const dashboardHotkeyConflict = $derived(!!form.hotkey && dashboardHotkeys.some((hotkey) => hotkeysEquivalent(form.hotkey, hotkey)))

  $effect(() => {
    if (!open || !form.pingEnabled) {
      availabilityExpanded = false
      if (!open) {
        hotkeyError = ''
        hotkeyPreview = ''
      }
    } else {
      availabilityExpanded = true
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

  function setPingEnabled(value: boolean) {
    form.pingEnabled = value
    availabilityExpanded = value
  }

  async function submit(e: Event) {
    e.preventDefault()
    titleErr = !form.title.trim()
    urlErr = !form.url.trim()
    if (titleErr || urlErr || dashboardHotkeyConflict) return
    await onSave()
  }
</script>

<Modal bind:open title={editing ? 'Edit item' : 'New item'}>
  <form class="space-y-3" onsubmit={submit}>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Title</span>
      <input
        class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary {titleErr
          ? 'field-error'
          : ''}"
        bind:value={form.title}
      />
      {#if titleErr}<span class="mt-1 text-xs text-danger">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">URL</span>
      <input
        class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary {urlErr
          ? 'field-error'
          : ''}"
        bind:value={form.url}
      />
      {#if urlErr}<span class="mt-1 text-xs text-danger">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Description</span>
      <input
        class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary"
        bind:value={form.description}
      />
    </label>
    <div>
      <span class="mb-1 block text-xs text-text-muted">Icon</span>
      <IconField bind:value={form.icon} bind:valueDark={form.iconDark} defaultIcon="mdi:link" />
    </div>
    <div class="hidden sm:block">
      <span class="mb-1 block text-xs text-text-muted">Hotkey</span>
      <div class="flex gap-2">
        <input
          readonly
          value={hotkeyPreview || formatHotkey(form.hotkey, form.hotkeyLabel)}
          placeholder="Focus and press a combination"
          aria-label="Item hotkey"
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
      {#if dashboardHotkeyConflict}
        <span class="mt-1 block text-xs text-danger">
          This hotkey is already assigned to a dashboard. Choose another combination before saving.
        </span>
      {:else if itemHotkeyConflict}
        <span class="mt-1 block text-xs text-amber-600 dark:text-amber-400">
          This hotkey is already used. All items assigned to it will open simultaneously. Your browser may ask for permission to open multiple tabs or
          block some of them.
        </span>
      {/if}
      <span class="mt-1 block text-xs {hotkeyError ? 'text-danger' : 'text-text-subtle'}">
        {hotkeyError || 'Press Ctrl, Alt, Shift or Meta together with another key. Backspace clears it.'}
      </span>
      <span class="mt-1 block text-xs text-text-subtle">
        Some shortcuts may be intercepted by your operating system or browser, especially combinations using Meta.
      </span>
    </div>
    <div class="rounded-btn border border-border bg-bg-elevated p-3">
      <div class="flex items-center justify-between gap-3">
        <button
          type="button"
          class="flex min-w-0 flex-1 items-center gap-2 text-left"
          aria-expanded={form.pingEnabled && availabilityExpanded}
          onclick={() => form.pingEnabled && (availabilityExpanded = !availabilityExpanded)}
        >
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            class="shrink-0 text-text-subtle transition-transform {form.pingEnabled && availabilityExpanded ? 'rotate-90' : ''}"
            aria-hidden="true"><path d="m9 18 6-6-6-6" /></svg
          >
          <span
            ><span class="block text-sm text-text">Availability check</span><span class="block text-xs text-text-subtle"
              >Check the service every 30 seconds</span
            ></span
          >
        </button>
        <Switch.Root
          checked={form.pingEnabled}
          onCheckedChange={(value) => setPingEnabled(!!value)}
          class="peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full bg-border transition data-[state=checked]:bg-primary"
        >
          <Switch.Thumb
            class="pointer-events-none block h-5 w-5 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-5"
          />
        </Switch.Root>
      </div>
      {#if form.pingEnabled && availabilityExpanded}
        <div class="mt-3 space-y-3 border-t border-border-soft pt-3">
          <label class="block">
            <span class="mb-1 block text-xs text-text-muted">Check URL <span class="text-text-subtle">(optional)</span></span>
            <input
              type="url"
              class="w-full rounded-btn border border-border bg-surface px-3 py-2 text-sm outline-none placeholder:text-text-subtle focus:border-primary"
              placeholder="Use the main URL"
              bind:value={form.pingUrl}
            />
            <span class="mt-1 block text-xs text-text-subtle"
              >Use a different internal or health-check address without changing the link destination.</span
            >
          </label>
          <label class="flex items-center justify-between gap-3">
            <span
              ><span class="block text-sm text-text">Skip TLS verification</span><span class="block text-xs text-text-subtle"
                >Allow self-signed or mismatched certificates</span
              ></span
            >
            <Switch.Root
              checked={form.pingSkipTls}
              onCheckedChange={(value) => (form.pingSkipTls = !!value)}
              class="peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full bg-border transition data-[state=checked]:bg-primary"
            >
              <Switch.Thumb
                class="pointer-events-none block h-5 w-5 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-5"
              />
            </Switch.Root>
          </label>
          <label class="flex items-center justify-between gap-3">
            <span
              ><span class="block text-sm text-text">Show failures only</span><span class="block text-xs text-text-subtle"
                >Hide the green chip while the URL is available</span
              ></span
            >
            <Switch.Root
              checked={form.pingOnlyDown}
              onCheckedChange={(value) => (form.pingOnlyDown = !!value)}
              class="peer inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full bg-border transition data-[state=checked]:bg-primary"
            >
              <Switch.Thumb
                class="pointer-events-none block h-5 w-5 translate-x-0.5 rounded-full bg-white shadow transition-transform data-[state=checked]:translate-x-5"
              />
            </Switch.Root>
          </label>
        </div>
      {/if}
    </div>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={() => (open = false)}>Cancel</button>
      <button
        type="submit"
        disabled={dashboardHotkeyConflict}
        class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-white disabled:cursor-not-allowed disabled:opacity-50">Save</button
      >
    </div>
  </form>
</Modal>
