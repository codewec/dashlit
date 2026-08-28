<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import { push } from 'svelte-spa-router';
  import { api, setToken } from '../lib/api';
  import { user, editMode, theme, applyTheme } from '../lib/stores';
  import Icon from './Icon.svelte';
  import logoUrl from '../assets/vite.svg';
  import type { DashListItem } from '../lib/dashboard-helpers';
  import { toastError } from '../lib/toasts';

  let {
    dashboards = [],
    currentSlug = '',
  }: {
    dashboards?: DashListItem[];
    currentSlug?: string;
  } = $props();

  const ownDashboards = $derived(dashboards.filter((d) => d.ownerId === $user?.id));
  const otherDashboards = $derived(dashboards.filter((d) => d.ownerId !== $user?.id));

  function toggleTheme() {
    const order: Array<'light' | 'dark' | 'system'> = ['dark', 'light', 'system'];
    const next = order[(order.indexOf($theme) + 1) % order.length];
    theme.set(next);
    applyTheme(next);
  }

  async function logout() {
    try {
      await api.logout();
    } catch (error: unknown) {
      toastError(error, 'Could not sign out on the server');
    }
    setToken(null);
    user.set(null);
    editMode.set(false);
    push('/login');
  }
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger class="inline-flex h-9 w-9 items-center justify-center rounded-btn border border-border text-text hover:bg-surface-2" aria-label="Menu">
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M4 6h16M4 12h16M4 18h16" />
    </svg>
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content class="z-50 min-w-[12rem] overflow-hidden rounded-xl border border-border bg-surface p-1 shadow-xl outline-none" sideOffset={6} align="end">
      {#if $user && dashboards.length > 0}
        {#if ownDashboards.length > 0}
          <div class="px-2 py-1.5 text-[10px] font-medium uppercase tracking-wide text-text-subtle">My dashboards</div>
        {/if}
        {#each ownDashboards as d}
          <DropdownMenu.Item
            class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm outline-none data-[highlighted]:bg-surface-2 {d.slug === currentSlug ? 'text-primary' : 'text-text'}"
            onSelect={() => push('/' + d.slug)}
          >
            {#if d.icon}
              <Icon icon={d.icon} iconDark={d.iconDark} size={16} />
            {:else}
              <img src={logoUrl} alt="" class="h-4 w-4" width="16" height="16" />
            {/if}
            <span class="truncate">{d.name}</span>
          </DropdownMenu.Item>
        {/each}
        {#if ownDashboards.length > 0 && otherDashboards.length > 0}
          <DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
        {/if}
        {#if otherDashboards.length > 0}
          <div class="px-2 py-1.5 text-[10px] font-medium uppercase tracking-wide text-text-subtle">Other dashboards</div>
        {/if}
        {#each otherDashboards as d}
          <DropdownMenu.Item
            class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm outline-none data-[highlighted]:bg-surface-2 {d.slug === currentSlug ? 'text-primary' : 'text-text'}"
            onSelect={() => push('/' + d.slug)}
          >
            {#if d.icon}
              <Icon icon={d.icon} iconDark={d.iconDark} size={16} />
            {:else}
              <img src={logoUrl} alt="" class="h-4 w-4" width="16" height="16" />
            {/if}
            <span class="min-w-0">
              <span class="block truncate">{d.name}</span>
              <span class="block truncate text-[10px] text-text-subtle">{d.ownerUsername}</span>
            </span>
          </DropdownMenu.Item>
        {/each}
        <DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
      {/if}

      <DropdownMenu.Item
        class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2"
        onSelect={(e) => {
          e.preventDefault();
          toggleTheme();
        }}
      >
        <span class="w-4 text-center">{$theme === 'dark' ? '☾' : $theme === 'light' ? '☀' : '◐'}</span>
        <span>Theme: {$theme}</span>
      </DropdownMenu.Item>

      {#if $user}
        <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={logout}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            ><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" /><path d="M16 17l5-5-5-5" /><path d="M21 12H9" /></svg
          >
          <span>Logout</span>
        </DropdownMenu.Item>
      {:else}
        <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => push('/login')}>
          <span>Sign in</span>
        </DropdownMenu.Item>
      {/if}
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>
