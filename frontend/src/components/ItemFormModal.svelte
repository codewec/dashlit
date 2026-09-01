<script lang="ts">
  import { Switch } from 'bits-ui'
  import type { ItemForm } from '../lib/dashboard-helpers'
  import Modal from './Modal.svelte'
  import IconField from './IconField.svelte'

  let {
    open = $bindable(false),
    form = $bindable(),
    editing = false,
    onSave,
  }: {
    open?: boolean
    form: ItemForm
    editing?: boolean
    onSave: () => void | Promise<void>
  } = $props()

  let titleErr = $state(false)
  let urlErr = $state(false)
  let availabilityExpanded = $state(false)

  $effect(() => {
    if (!open || !form.pingEnabled) {
      availabilityExpanded = false
    } else {
      availabilityExpanded = true
    }
  })

  function setPingEnabled(value: boolean) {
    form.pingEnabled = value
    availabilityExpanded = value
  }

  async function submit(e: Event) {
    e.preventDefault()
    titleErr = !form.title.trim()
    urlErr = !form.url.trim()
    if (titleErr || urlErr) return
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
      <button type="submit" class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>
