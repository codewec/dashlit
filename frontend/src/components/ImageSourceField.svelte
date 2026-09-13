<script lang="ts">
  import { api } from '../lib/api'
  import { detectIconSourceTab, iconSrc } from '../lib/icon-helpers'
  import { toastError } from '../lib/toasts'

  let {
    value = $bindable(''),
    label = 'Background image',
  }: {
    value?: string
    label?: string
  } = $props()

  type SourceTab = 'url' | 'upload'

  let sourceTab = $state<SourceTab>('url')
  let urlInput = $state('')
  let initialized = $state(false)

  const previewSrc = $derived(value ? iconSrc(value) : '')

  $effect(() => {
    if (!initialized) {
      const detected = detectIconSourceTab(value)
      sourceTab = detected === 'upload' ? 'upload' : 'url'
      if (sourceTab === 'url' && value) urlInput = value
      initialized = true
    }
  })

  function applyUrl() {
    if (urlInput.trim()) value = urlInput.trim()
  }

  function clear() {
    value = ''
    urlInput = ''
  }

  async function onFile(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0]
    if (!file) return
    try {
      const res = await api.uploadIcon(file)
      value = res.icon
    } catch (error: unknown) {
      toastError(error, 'Could not upload image')
    }
  }
</script>

<div class="space-y-2">
  <div class="flex items-center justify-between gap-2">
    <span class="text-sm font-medium text-text">{label}</span>
    {#if value}
      <button type="button" class="text-xs text-text-muted hover:text-danger" onclick={clear}>Clear</button>
    {/if}
  </div>

  <div class="flex gap-1 rounded-lg bg-surface-2 p-1">
    {#each ['url', 'upload'] as SourceTab[] as t}
      <button
        type="button"
        class="flex-1 rounded-md px-2 py-1.5 text-xs font-medium transition
          {sourceTab === t ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'}"
        onclick={() => (sourceTab = t)}
      >
        {t === 'url' ? 'URL' : 'Upload'}
      </button>
    {/each}
  </div>

  <div class="flex items-center gap-2">
    <div class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft bg-bg">
      {#if previewSrc}
        <img src={previewSrc} alt="" class="h-full w-full object-cover" />
      {:else}
        <span class="text-[10px] text-text-subtle">—</span>
      {/if}
    </div>

    {#if sourceTab === 'url'}
      <input
        class="h-10 min-w-0 flex-1 rounded-btn border border-border bg-bg-elevated px-3 text-sm outline-none focus:border-primary"
        placeholder="https://…"
        bind:value={urlInput}
        onkeydown={(e) => e.key === 'Enter' && (e.preventDefault(), applyUrl())}
      />
      <button
        type="button"
        class="h-10 shrink-0 rounded-btn bg-primary px-4 text-sm font-medium text-primary-fg hover:bg-primary-hover"
        onclick={applyUrl}>Load</button
      >
    {:else}
      <label
        class="flex h-10 min-w-0 flex-1 cursor-pointer items-center justify-center gap-2 rounded-btn border border-dashed border-border bg-bg-elevated px-3 text-xs text-text-muted hover:border-primary"
      >
        <span>Upload PNG / JPEG / WebP</span>
        <input type="file" accept="image/png,image/jpeg,image/webp" class="hidden" onchange={onFile} />
      </label>
    {/if}
  </div>
</div>
