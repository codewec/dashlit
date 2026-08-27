<script lang="ts">
  import type { Snippet } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Select } from 'bits-ui';
  import { api, setToken } from '../lib/api';
  import { user, editMode, theme, searchQuery, applyTheme } from '../lib/stores';
  import { cn } from '../lib/cn';

  let {
    children,
    dashboards = [],
    currentSlug = '',
    wide = false,
    onCreateDashboard,
    onDeleteDashboard,
  }: {
    children?: Snippet;
    dashboards?: { id: string; name: string; slug: string }[];
    currentSlug?: string;
    wide?: boolean;
    onCreateDashboard?: () => void;
    onDeleteDashboard?: () => void;
  } = $props();

  const selected = $derived(dashboards.find((d) => d.slug === currentSlug) ?? null);

  function toggleTheme() {
    const order: Array<'light' | 'dark' | 'system'> = ['dark', 'light', 'system'];
    const next = order[(order.indexOf($theme) + 1) % order.length];
    theme.set(next);
    applyTheme(next);
  }

  async function logout() {
    try {
      await api.logout();
    } catch {}
    setToken(null);
    user.set(null);
    editMode.set(false);
    push('/login');
  }
</script>

<header class="sticky top-0 z-40 border-b border-[var(--color-border-soft)] bg-[var(--color-bg-elevated)]/80 backdrop-blur-md">
  <div class={cn('mx-auto flex flex-wrap items-center gap-3 px-4 py-3', wide ? 'max-w-none' : 'max-w-6xl')}>
    <a href="#/" class="text-sm font-semibold tracking-tight text-[var(--color-text)]">DashLit</a>

    {#if $user && dashboards.length > 0 && selected}
      <div class="flex items-center gap-1">
        <Select.Root
          type="single"
          value={selected.slug}
          onValueChange={(v) => {
            if (v) push('/' + v);
          }}
        >
          <Select.Trigger
            class="inline-flex h-8 min-w-[9rem] max-w-[14rem] items-center justify-between gap-2 rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-surface)] px-2.5 text-xs text-[var(--color-text)] outline-none hover:bg-[var(--color-surface-2)]"
          >
            <span class="truncate">{selected.name}</span>
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="shrink-0 opacity-60"><path d="m6 9 6 6 6-6" /></svg>
          </Select.Trigger>
          <Select.Portal>
            <Select.Content
              class="z-50 min-w-[var(--bits-select-anchor-width)] overflow-hidden rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)] p-1 shadow-xl outline-none"
              sideOffset={6}
            >
              {#each dashboards as d}
                <Select.Item
                  value={d.slug}
                  label={d.name}
                  class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs outline-none data-[highlighted]:bg-[var(--color-surface-2)] data-[selected]:text-[var(--color-primary)]"
                >
                  {#snippet children({ selected: isSelected })}
                    <span class="w-3 shrink-0 text-[var(--color-primary)]">{isSelected ? '✓' : ''}</span>
                    <span class="truncate">{d.name}</span>
                  {/snippet}
                </Select.Item>
              {/each}
            </Select.Content>
          </Select.Portal>
        </Select.Root>
        <button
          type="button"
          class="flex h-8 w-8 items-center justify-center rounded-[var(--radius-btn)] border border-[var(--color-border)] text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-text)]"
          title="New dashboard"
          onclick={() => onCreateDashboard?.()}
        >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
        </button>
      </div>
    {:else if $user}
      <button
        type="button"
        class="flex h-8 items-center gap-1 rounded-[var(--radius-btn)] border border-[var(--color-border)] px-2.5 text-xs text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-text)]"
        onclick={() => onCreateDashboard?.()}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
        New dashboard
      </button>
    {/if}

    <div class="min-w-[10rem] flex-1">
      <input
        type="search"
        placeholder="Filter items…"
        bind:value={$searchQuery}
        class="w-full max-w-xs rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-1.5 text-sm outline-none placeholder:text-[var(--color-text-subtle)] focus:border-[var(--color-primary)]"
      />
    </div>

    <div class="ml-auto flex items-center gap-1.5">
      <button type="button" class="rounded-[var(--radius-btn)] p-2 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)]" onclick={toggleTheme} title="Theme">
        {$theme === 'dark' ? '☾' : $theme === 'light' ? '☀' : '◐'}
      </button>
      {#if $user}
        {#if $editMode && onDeleteDashboard}
          <button
            type="button"
            class="rounded-[var(--radius-btn)] border border-[var(--color-danger)]/40 px-3 py-1.5 text-xs font-medium text-[var(--color-danger)] hover:bg-[var(--color-danger-soft)]"
            onclick={() => onDeleteDashboard?.()}
          >
            Delete
          </button>
        {/if}
        <button
          type="button"
          class="rounded-[var(--radius-btn)] px-3 py-1.5 text-xs font-medium transition
            {$editMode ? 'bg-[var(--color-primary)] text-white' : 'border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-surface-2)]'}"
          onclick={() => editMode.update((v) => !v)}
        >
          {$editMode ? 'Done' : 'Edit'}
        </button>
        <span class="hidden text-xs text-[var(--color-text-muted)] sm:inline">{$user.username}</span>
        <button type="button" class="rounded-[var(--radius-btn)] px-2 py-1.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)]" onclick={logout}> Logout </button>
      {:else}
        <a href="#/login" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-1.5 text-xs font-medium text-white hover:bg-[var(--color-primary-hover)]"> Sign in </a>
      {/if}
    </div>
  </div>
</header>

<main class={cn('mx-auto px-4 py-6', wide ? 'max-w-none' : 'max-w-6xl')}>
  {@render children?.()}
</main>
