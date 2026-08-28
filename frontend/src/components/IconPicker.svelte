<script lang="ts">
  import { api, iconSrc } from '../lib/api';
  import Icon from './Icon.svelte';
  import { toastError } from '../lib/toasts';

  let { value = $bindable('') }: { value?: string } = $props();

  type SourceTab = 'search' | 'url' | 'upload';

  function detectSource(v: string): SourceTab {
    if (!v) return 'search';
    if (v.startsWith('local:')) return 'upload';
    if (v.startsWith('http://') || v.startsWith('https://') || v.startsWith('/')) return 'url';
    return 'search';
  }

  let sourceTab = $state<SourceTab>('search');
  let query = $state('');
  let results = $state<string[]>([]);
  let searching = $state(false);
  let urlInput = $state('');
  let openDropdown = $state(false);
  let initialized = $state(false);
  let debounce: ReturnType<typeof setTimeout> | undefined;

  $effect(() => {
    if (!initialized) {
      sourceTab = detectSource(value);
      if (sourceTab === 'url') urlInput = value;
      initialized = true;
    }
  });

  $effect(() => {
    const q = query;
    clearTimeout(debounce);
    if (sourceTab !== 'search') {
      openDropdown = false;
      return;
    }
    if (!q.trim()) {
      results = [];
      searching = false;
      openDropdown = false;
      return;
    }
    searching = true;
    openDropdown = true;
    debounce = setTimeout(async () => {
      try {
        results = await api.searchIcons(q);
      } catch (e: unknown) {
        results = [];
        toastError(e, 'Could not search icons');
      } finally {
        searching = false;
      }
    }, 280);
  });

  function selectIcon(name: string) {
    value = name;
    openDropdown = false;
    query = '';
  }

  function applyUrl() {
    if (urlInput.trim()) value = urlInput.trim();
  }

  async function onFile(e: Event) {
    const file = (e.target as HTMLInputElement).files?.[0];
    if (!file) return;
    try {
      const res = await api.uploadIcon(file);
      value = res.icon;
    } catch (error: unknown) {
      toastError(error, 'Could not upload icon');
    }
  }
</script>

<div class="space-y-2">
  <div class="flex gap-1 rounded-lg bg-surface-2 p-1">
    {#each ['search', 'url', 'upload'] as SourceTab[] as t}
      <button
        type="button"
        class="flex-1 rounded-md px-2 py-1.5 text-xs font-medium transition
          {sourceTab === t ? 'bg-surface text-text shadow-sm' : 'text-text-muted hover:text-text'}"
        onclick={() => {
          sourceTab = t;
          openDropdown = false;
        }}
      >
        {t === 'search' ? 'Search' : t === 'url' ? 'URL' : 'Upload'}
      </button>
    {/each}
  </div>

  <div class="relative min-h-[2.5rem]">
    {#if sourceTab === 'search'}
      <div class="flex items-center gap-2">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft bg-bg-elevated">
          {#if value}
            <Icon icon={value} size={22} />
          {:else}
            <span class="text-[10px] text-text-subtle">—</span>
          {/if}
        </div>
        <input
          class="min-w-0 flex-1 rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary"
          placeholder="Search icons…"
          bind:value={query}
          onfocus={() => {
            if (query.trim()) openDropdown = true;
          }}
        />
      </div>
      {#if openDropdown && (searching || results.length || query)}
        <div class="absolute left-0 right-0 top-full z-[60] mt-1 max-h-48 overflow-y-auto rounded-xl border border-border bg-surface p-2 shadow-xl">
          {#if searching}
            <p class="py-3 text-center text-xs text-text-subtle">Searching…</p>
          {:else if results.length === 0}
            <p class="py-3 text-center text-xs text-text-subtle">No results</p>
          {:else}
            <div class="grid grid-cols-6 gap-1 sm:grid-cols-8">
              {#each results as name}
                <button
                  type="button"
                  title={name}
                  class="flex aspect-square items-center justify-center rounded-lg border border-transparent p-1 hover:border-primary hover:bg-surface-2
                    {value === name ? 'border-primary bg-surface-2' : ''}"
                  onclick={() => selectIcon(name)}
                >
                  <img src={iconSrc(name)} alt="" class="h-5 w-5 object-contain" loading="lazy" />
                </button>
              {/each}
            </div>
          {/if}
        </div>
      {/if}
    {:else if sourceTab === 'url'}
      <div class="flex h-9 items-center gap-2">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft bg-bg-elevated">
          {#if value}
            <Icon icon={value} size={22} />
          {:else}
            <span class="text-[10px] text-text-subtle">—</span>
          {/if}
        </div>
        <input
          class="min-w-0 flex-1 rounded-btn border border-border bg-bg-elevated px-3 text-sm outline-none focus:border-primary"
          placeholder="https://…"
          bind:value={urlInput}
          onkeydown={(e) => e.key === 'Enter' && (e.preventDefault(), applyUrl())}
        />
        <button type="button" class="h-9 shrink-0 rounded-btn bg-primary px-3 text-xs font-medium text-white hover:bg-primary-hover" onclick={applyUrl}> Apply </button>
      </div>
    {:else}
      <div class="flex items-center gap-2">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft bg-bg-elevated">
          {#if value}
            <Icon icon={value} size={22} />
          {:else}
            <span class="text-[10px] text-text-subtle">—</span>
          {/if}
        </div>
        <label
          class="flex h-9 min-w-0 flex-1 cursor-pointer items-center justify-center gap-2 rounded-btn border border-dashed border-border bg-bg-elevated px-3 text-xs text-text-muted hover:border-primary"
        >
          <span>Upload PNG / SVG / WebP</span>
          <input type="file" accept="image/png,image/svg+xml,image/jpeg,image/webp,image/x-icon" class="hidden" onchange={onFile} />
        </label>
      </div>
    {/if}
  </div>
</div>
