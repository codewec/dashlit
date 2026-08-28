<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import { push } from 'svelte-spa-router';
  import { api, setToken } from '../lib/api';
  import { user, editMode, theme, setTheme } from '../lib/stores';
  import Icon from './Icon.svelte';
  import logoUrl from '../assets/vite.svg';
  import type { DashListItem } from '../lib/dashboard-helpers';
  import { toastError } from '../lib/toasts';
  import { themeOptions } from '../lib/themes';

  let {
    dashboards = [],
    currentSlug = '',
    showEdit = false,
  }: {
    dashboards?: DashListItem[];
    currentSlug?: string;
    showEdit?: boolean;
  } = $props();

  const ownDashboards = $derived(dashboards.filter((d) => d.ownerId === $user?.id));
  const otherDashboards = $derived(dashboards.filter((d) => d.ownerId !== $user?.id));
  let menuOpen = $state(false);
  let themeExpanded = $state(false);

  $effect(() => {
    if (!menuOpen) themeExpanded = false;
  });

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

<DropdownMenu.Root bind:open={menuOpen}>
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
        aria-expanded={themeExpanded}
        onSelect={(event) => {
          event.preventDefault();
          themeExpanded = !themeExpanded;
        }}
      >
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 3a9 9 0 0 0 0 18c-2.2-2.5-2.2-15.5 0-18Z" /></svg>
          <span class="flex-1">Theme</span>
          <span class="transition-transform {themeExpanded ? 'rotate-90' : ''}" aria-hidden="true">›</span>
      </DropdownMenu.Item>
      {#if themeExpanded}
        <div class="ml-3 border-l border-border-soft pl-1">
          {#each themeOptions as option, index}
            {#if index === 1 || index === 3}
              <div class="px-2.5 py-1 text-[10px] font-medium uppercase tracking-wide text-text-subtle">
                {index === 1 ? 'Light' : 'Dark'}
              </div>
            {/if}
            <DropdownMenu.Item
              class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2"
              onSelect={() => setTheme(option.value)}
            >
              <span class="h-3.5 w-3.5 shrink-0 rounded-full border border-border" style:background={option.swatch}></span>
              <span class="flex-1">{option.label}</span>
              {#if $theme === option.value}<span class="text-primary">✓</span>{/if}
            </DropdownMenu.Item>
          {/each}
        </div>
      {/if}

      {#if $user}
        {#if showEdit && !$editMode}
          <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => editMode.set(true)}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9" /><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" /></svg>
            <span>Edit</span>
          </DropdownMenu.Item>
        {/if}
        {#if $user.role === 'admin'}
          <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => push('/admin')}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 3 4 7v5c0 5 3.4 8.3 8 9 4.6-.7 8-4 8-9V7l-8-4Z" /><path d="M9 12l2 2 4-4" /></svg>
            <span>Administration</span>
          </DropdownMenu.Item>
        {/if}
        <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => push('/profile')}>
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></svg>
          <span>Profile</span>
        </DropdownMenu.Item>
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
