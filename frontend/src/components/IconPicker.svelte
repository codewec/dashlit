<script lang="ts">
  import { api, iconSrc } from '../lib/api';
  import Icon from './Icon.svelte';

  let { value = $bindable('') }: { value?: string } = $props();

  type Tab = 'search' | 'url' | 'upload';
  let tab = $state<Tab>('search');
  let query = $state('');
  let results = $state<string[]>([]);
  let searching = $state(false);
  let urlInput = $state('');
  let debounce: ReturnType<typeof setTimeout> | undefined;

  $effect(() => {
    const q = query;
    clearTimeout(debounce);
    if (tab !== 'search') return;
    if (!q.trim()) {
      results = [];
      searching = false;
      return;
    }
    searching = true;
    debounce = setTimeout(async () => {
      results = await api.searchIcons(q);
      searching = false;
    }, 280);
  });

  function selectIcon(name: string) {
    value = name;
  }

  function applyUrl() {
    if (urlInput.trim()) value = urlInput.trim();
  }

  async function onFile(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    const res = await api.uploadIcon(file);
    value = res.icon;
  }
</script>

<div class="space-y-3">
  <div class="flex gap-1 rounded-lg bg-[var(--color-surface-2)] p-1">
    {#each ['search', 'url', 'upload'] as Tab[] as t}
      <button
        type="button"
        class="flex-1 rounded-md px-2 py-1.5 text-xs font-medium transition
          {tab === t ? 'bg-[var(--color-surface)] text-[var(--color-text)] shadow-sm' : 'text-[var(--color-text-muted)] hover:text-[var(--color-text)]'}"
        onclick={() => (tab = t)}
      >
        {t === 'search' ? 'Search' : t === 'url' ? 'URL' : 'Upload'}
      </button>
    {/each}
  </div>

  {#if value}
    <div class="flex items-center gap-3 rounded-lg border border-[var(--color-border-soft)] bg-[var(--color-bg-elevated)] px-3 py-2">
      <Icon icon={value} size={28} />
      <span class="min-w-0 flex-1 truncate text-xs text-[var(--color-text-muted)]">{value}</span>
      <button type="button" class="text-xs text-[var(--color-text-subtle)] hover:text-[var(--color-danger)]" onclick={() => (value = '')}> Clear </button>
    </div>
  {/if}

  {#if tab === 'search'}
    <input
      class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]"
      placeholder="Search icons…"
      bind:value={query}
    />
    <div class="grid max-h-48 grid-cols-6 gap-1.5 overflow-y-auto sm:grid-cols-8">
      {#if searching}
        <p class="col-span-full py-4 text-center text-xs text-[var(--color-text-subtle)]">Searching…</p>
      {:else if results.length === 0 && query}
        <p class="col-span-full py-4 text-center text-xs text-[var(--color-text-subtle)]">No results</p>
      {:else}
        {#each results as name}
          <button
            type="button"
            title={name}
            class="flex aspect-square items-center justify-center rounded-lg border border-transparent p-1.5 hover:border-[var(--color-primary)] hover:bg-[var(--color-surface-2)]
              {value === name ? 'border-[var(--color-primary)] bg-[var(--color-surface-2)]' : ''}"
            onclick={() => selectIcon(name)}
          >
            <img src={iconSrc(name)} alt={name} class="h-6 w-6 object-contain" loading="lazy" />
          </button>
        {/each}
      {/if}
    </div>
  {:else if tab === 'url'}
    <div class="flex gap-2">
      <input
        class="min-w-0 flex-1 rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]"
        placeholder="https://…"
        bind:value={urlInput}
        onkeydown={(e) => e.key === 'Enter' && (e.preventDefault(), applyUrl())}
      />
      <button type="button" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-2 text-sm font-medium text-white hover:bg-[var(--color-primary-hover)]" onclick={applyUrl}>
        Apply
      </button>
    </div>
  {:else}
    <label
      class="flex cursor-pointer flex-col items-center justify-center gap-2 rounded-[var(--radius-card)] border border-dashed border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-4 py-8 text-sm text-[var(--color-text-muted)] hover:border-[var(--color-primary)]"
    >
      <span>Click to upload PNG / SVG / WebP</span>
      <input type="file" accept="image/png,image/svg+xml,image/jpeg,image/webp,image/x-icon" class="hidden" onchange={onFile} />
    </label>
  {/if}
</div>
