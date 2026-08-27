<script lang="ts">
  import type { Snippet } from 'svelte';
  import { push } from 'svelte-spa-router';
  import { Select } from 'bits-ui';
  import { api, setToken } from '../lib/api';
  import { user, editMode, theme, searchQuery, applyTheme } from '../lib/stores';
  import { cn } from '../lib/cn';
  import Icon from '../components/Icon.svelte';
  import NavMenu from '../components/NavMenu.svelte';
  import logoUrl from '../assets/vite.svg';

  let {
    children,
    dashboards = [],
    currentSlug = '',
    wide = false,
  }: {
    children?: Snippet;
    dashboards?: { id: string; name: string; slug: string; icon?: string; iconDark?: string }[];
    currentSlug?: string;
    wide?: boolean;
  } = $props();

  const selected = $derived(dashboards.find((d) => d.slug === currentSlug) ?? null);
  const year = new Date().getFullYear();

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

<div class="flex min-h-dvh flex-col">
  <header class="sticky top-0 z-40 border-b border-border-soft bg-bg-elevated/80 backdrop-blur-md">
    <div class={cn('mx-auto flex items-center gap-2 px-4 py-2.5 sm:gap-3', wide ? 'max-w-none' : 'max-w-6xl')}>
      <a href="#/" class="flex shrink-0 items-center gap-2 text-sm font-semibold tracking-tight text-text">
        <img src={logoUrl} alt="" class="h-6 w-6" width="24" height="24" />
        <span>DashLit</span>
      </a>

      {#if $user && dashboards.length > 0 && selected}
        <div class="hidden sm:block">
          <Select.Root
            type="single"
            value={selected.slug}
            onValueChange={(v) => {
              if (v) push('/' + v);
            }}
          >
            <Select.Trigger
              class="inline-flex h-8 min-w-[9rem] max-w-[14rem] items-center justify-between gap-2 rounded-btn border border-border bg-surface px-2.5 text-xs text-text outline-none hover:bg-surface-2"
            >
              <span class="flex min-w-0 items-center gap-1.5">
                {#if selected.icon}
                  <Icon icon={selected.icon} iconDark={selected.iconDark} size={16} />
                {:else}
                  <img src={logoUrl} alt="" class="h-4 w-4" width="16" height="16" />
                {/if}
                <span class="truncate">{selected.name}</span>
              </span>
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="shrink-0 opacity-60"><path d="m6 9 6 6 6-6" /></svg>
            </Select.Trigger>
            <Select.Portal>
              <Select.Content class="z-50 min-w-[var(--bits-select-anchor-width)] overflow-hidden rounded-xl border border-border bg-surface p-1 shadow-xl outline-none" sideOffset={6}>
                {#each dashboards as d}
                  <Select.Item
                    value={d.slug}
                    label={d.name}
                    class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-1.5 text-xs outline-none data-[highlighted]:bg-surface-2 data-[selected]:text-primary"
                  >
                    {#snippet children({ selected: isSelected })}
                      <span class="w-3 shrink-0 text-primary">{isSelected ? '✓' : ''}</span>
                      {#if d.icon}
                        <Icon icon={d.icon} iconDark={d.iconDark} size={16} />
                      {:else}
                        <img src={logoUrl} alt="" class="h-4 w-4" width="16" height="16" />
                      {/if}
                      <span class="truncate">{d.name}</span>
                    {/snippet}
                  </Select.Item>
                {/each}
              </Select.Content>
            </Select.Portal>
          </Select.Root>
        </div>
      {/if}

      <div class="min-w-0 flex-1">
        <input
          type="search"
          placeholder="Filter…"
          bind:value={$searchQuery}
          class="w-full max-w-xs rounded-btn border border-border bg-surface px-3 py-1.5 text-sm outline-none placeholder:text-text-subtle focus:border-primary"
        />
      </div>

      <div class="ml-auto hidden items-center gap-1 sm:flex">
        <button type="button" class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text" onclick={toggleTheme} title="Theme: {$theme}">
          {$theme === 'dark' ? '☾' : $theme === 'light' ? '☀' : '◐'}
        </button>
        {#if $user}
          <button type="button" class="flex h-8 w-8 items-center justify-center rounded-btn text-text-muted hover:bg-surface-2 hover:text-text" onclick={logout} title="Logout" aria-label="Logout">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              ><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="M16 17l5-5-5-5" /><path d="M21 12H9" /></svg
            >
          </button>
        {:else}
          <a href="#/login" class="rounded-btn bg-primary px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-hover"> Sign in </a>
        {/if}
      </div>

      <div class="ml-auto sm:hidden">
        <NavMenu {dashboards} {currentSlug} />
      </div>
    </div>
  </header>

  <main class={cn('mx-auto w-full flex-1 px-4 py-6', wide ? 'max-w-none' : 'max-w-6xl')}>
    {@render children?.()}
  </main>

  <footer class="border-t border-border-soft bg-bg-elevated/60">
    <div class={cn('mx-auto flex items-center justify-center gap-2 px-4 py-3 text-xs text-text-subtle', wide ? 'max-w-none' : 'max-w-6xl')}>
      <img src={logoUrl} alt="" class="h-3.5 w-3.5 opacity-70" width="14" height="14" />
      <span>DashLit · {year}</span>
    </div>
  </footer>
</div>
