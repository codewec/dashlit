<script lang="ts">
  import type { Dashboard } from '../lib/api';
  import type { DashListItem } from '../lib/dashboard-helpers';
  import Icon from './Icon.svelte';
  import NavMenu from './NavMenu.svelte';

  let {
    dashboard,
    dashboards = [],
  }: {
    dashboard: Dashboard;
    dashboards?: DashListItem[];
  } = $props();
</script>

<div class={`mb-6 flex gap-3 ${dashboard.description ? 'items-start' : 'items-center'}`}>
  {#if dashboard.icon}
    <Icon icon={dashboard.icon} iconDark={dashboard.iconDark} size={40} class={`shrink-0 rounded-xl ${dashboard.description ? 'mt-0.5' : ''}`} />
  {/if}
  <div class="min-w-0 flex-1">
    <h1 class="text-xl font-semibold tracking-tight text-text">{dashboard.name}</h1>
    {#if dashboard.description}
      <p class="mt-1 text-sm text-text-muted">{dashboard.description}</p>
    {/if}
  </div>
  <div class="shrink-0">
    <NavMenu {dashboards} currentSlug={dashboard.slug} showEdit />
  </div>
</div>
