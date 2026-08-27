<script lang="ts">
  import type { Snippet } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { api, setToken } from '../lib/api';
  import { user, editMode, theme, searchQuery, applyTheme } from '../lib/stores';

  let {
    children,
    dashboards = [],
  }: {
    children?: Snippet;
    dashboards?: { id: string; name: string; slug: string }[];
  } = $props();

  function toggleTheme() {
    const order: Array<'light' | 'dark' | 'system'> = ['dark', 'light', 'system'];
    const next = order[(order.indexOf($theme) + 1) % order.length];
    theme.set(next);
    applyTheme(next);
  }

  async function logout() {
    try { await api.logout(); } catch {}
    setToken(null);
    user.set(null);
    editMode.set(false);
    push('/login');
  }
</script>

<header class="sticky top-0 z-40 border-b border-[var(--color-border-soft)] bg-[var(--color-bg-elevated)]/80 backdrop-blur-md">
  <div class="mx-auto flex max-w-6xl flex-wrap items-center gap-3 px-4 py-3">
    <a href="#/" class="text-sm font-semibold tracking-tight text-[var(--color-text)]">Bookmarks</a>

    {#if $user && dashboards.length}
      <select
        class="max-w-[10rem] rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-surface)] px-2 py-1.5 text-xs text-[var(--color-text)] outline-none"
        onchange={(e) => {
          const slug = (e.currentTarget as HTMLSelectElement).value;
          if (slug) push('/' + slug);
        }}
      >
        <option value="">Dashboards</option>
        {#each dashboards as d}
          <option value={d.slug}>{d.name}</option>
        {/each}
      </select>
    {/if}

    <div class="min-w-[12rem] flex-1">
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
        <button
          type="button"
          class="rounded-[var(--radius-btn)] px-3 py-1.5 text-xs font-medium transition
            {$editMode
              ? 'bg-[var(--color-primary)] text-white'
              : 'border border-[var(--color-border)] text-[var(--color-text)] hover:bg-[var(--color-surface-2)]'}"
          onclick={() => editMode.update((v) => !v)}
        >
          {$editMode ? 'Done' : 'Edit'}
        </button>
        <span class="hidden text-xs text-[var(--color-text-muted)] sm:inline">{$user.username}</span>
        <button type="button" class="rounded-[var(--radius-btn)] px-2 py-1.5 text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)]" onclick={logout}>
          Logout
        </button>
      {:else}
        <a href="#/login" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-1.5 text-xs font-medium text-white hover:bg-[var(--color-primary-hover)]">
          Sign in
        </a>
      {/if}
    </div>
  </div>
</header>

<main class="mx-auto max-w-6xl px-4 py-6">
  {@render children?.()}
</main>
