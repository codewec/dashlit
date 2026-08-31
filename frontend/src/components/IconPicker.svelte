<script lang="ts">
  import { api, type IconSearchResult } from '../lib/api';
  import {
    applyIconSelection,
    combineIconResults,
    detectIconSourceTab,
    iconForVariant,
    iconPreviewClass,
    iconSearchTitle,
    type IconSourceTab,
    type IconVariant,
  } from '../lib/icon-helpers';
  import Icon from './Icon.svelte';
  import { toastError } from '../lib/toasts';

  let {
    value = $bindable(''),
    pairedValue = $bindable(''),
    variant = 'light',
  }: {
    value?: string;
    pairedValue?: string;
    variant?: IconVariant;
  } = $props();

  let sourceTab = $state<IconSourceTab>('search');
  let query = $state('');
  let iconifyResults = $state<IconSearchResult[]>([]);
  let selfhstResults = $state<IconSearchResult[]>([]);
  let searchingIconify = $state(false);
  let searchingSelfhst = $state(false);
  let urlInput = $state('');
  let openDropdown = $state(false);
  let initialized = $state(false);
  let debounce: ReturnType<typeof setTimeout> | undefined;
  let searchSequence = 0;

  const results = $derived(combineIconResults(selfhstResults, iconifyResults));
  const searching = $derived(searchingIconify || searchingSelfhst);

  $effect(() => {
    if (!initialized) {
      sourceTab = detectIconSourceTab(value);
      if (sourceTab === 'url') urlInput = value;
      initialized = true;
    }
  });

  $effect(() => {
    const q = query;
    const sequence = ++searchSequence;
    clearTimeout(debounce);
    if (sourceTab !== 'search') {
      openDropdown = false;
      return;
    }
    if (!q.trim()) {
      iconifyResults = [];
      selfhstResults = [];
      searchingIconify = false;
      searchingSelfhst = false;
      openDropdown = false;
      return;
    }
    // Never show results produced for a shorter, previous query while the
    // debounced requests for the current query are pending.
    iconifyResults = [];
    selfhstResults = [];
    searchingIconify = true;
    searchingSelfhst = true;
    openDropdown = true;
    debounce = setTimeout(() => {
      void api.searchIconifyIcons(q).then(
        (found) => { if (sequence === searchSequence) iconifyResults = found; },
        () => { if (sequence === searchSequence) iconifyResults = []; },
      ).finally(() => { if (sequence === searchSequence) searchingIconify = false; });
      void api.searchSelfhstIcons(q).then(
        (found) => { if (sequence === searchSequence) selfhstResults = found; },
        () => { if (sequence === searchSequence) selfhstResults = []; },
      ).finally(() => { if (sequence === searchSequence) searchingSelfhst = false; });
    }, 280);
  });

  function selectIcon(result: IconSearchResult) {
    const selection = applyIconSelection(result, variant, pairedValue);
    value = selection.value;
    pairedValue = selection.pairedValue;
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
    {#each ['search', 'url', 'upload'] as IconSourceTab[] as t}
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
        <div class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft {iconPreviewClass(variant)}">
          {#if value}
            <Icon icon={value} theme={variant} size={22} />
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
          {#if results.length > 0}
            <div class="grid grid-cols-6 gap-1 sm:grid-cols-8">
              {#each results as result}
                <button
                  type="button"
                  title={iconSearchTitle(result)}
                  class="flex aspect-square items-center justify-center rounded-lg border border-transparent p-1 hover:border-primary
                    {variant === 'dark' ? 'bg-[#181825] hover:bg-[#24243a]' : 'bg-white hover:bg-[#f2f2f5]'}
                    {value === iconForVariant(result, variant) ? 'border-primary ring-1 ring-primary' : ''}"
                  onclick={() => selectIcon(result)}
                >
                  <Icon icon={iconForVariant(result, variant)} theme={variant} size={20} />
                  <span class="sr-only">{iconSearchTitle(result)}</span>
                </button>
              {/each}
            </div>
            <p class="mt-2 text-center text-[10px] text-text-subtle">
              Icons by <a class="underline hover:text-text-muted" href="https://selfh.st/icons/" target="_blank" rel="noreferrer">selfh.st/icons</a> (CC BY 4.0) and Iconify
            </p>
            {#if searching}
              <p class="mt-1 text-center text-[10px] text-text-subtle">Loading more results…</p>
            {/if}
          {:else if searching}
            <p class="py-3 text-center text-xs text-text-subtle">Searching…</p>
          {:else}
            <p class="py-3 text-center text-xs text-text-subtle">No results</p>
          {/if}
        </div>
      {/if}
    {:else if sourceTab === 'url'}
      <div class="flex h-10 items-center gap-2">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft {iconPreviewClass(variant)}">
          {#if value}
            <Icon icon={value} theme={variant} size={22} />
          {:else}
            <span class="text-[10px] text-text-subtle">—</span>
          {/if}
        </div>
        <input
          class="h-10 min-w-0 flex-1 rounded-btn border border-border bg-bg-elevated px-3 text-sm outline-none focus:border-primary"
          placeholder="https://…"
          bind:value={urlInput}
          onkeydown={(e) => e.key === 'Enter' && (e.preventDefault(), applyUrl())}
        />
        <button type="button" class="h-10 shrink-0 rounded-btn bg-primary px-4 text-sm font-medium text-white hover:bg-primary-hover" onclick={applyUrl}>Load</button>
      </div>
    {:else}
      <div class="flex items-center gap-2">
        <div class="flex h-10 w-10 shrink-0 items-center justify-center overflow-hidden rounded-lg border border-border-soft {iconPreviewClass(variant)}">
          {#if value}
            <Icon icon={value} theme={variant} size={22} />
          {:else}
            <span class="text-[10px] text-text-subtle">—</span>
          {/if}
        </div>
        <label
          class="flex h-10 min-w-0 flex-1 cursor-pointer items-center justify-center gap-2 rounded-btn border border-dashed border-border bg-bg-elevated px-3 text-xs text-text-muted hover:border-primary"
        >
          <span>Upload PNG / SVG / WebP</span>
          <input type="file" accept="image/png,image/svg+xml,image/jpeg,image/webp,image/x-icon" class="hidden" onchange={onFile} />
        </label>
      </div>
    {/if}
  </div>
</div>
