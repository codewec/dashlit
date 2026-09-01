<script lang="ts">
  import type { Group } from '../lib/api'
  import type { DashListItem } from '../lib/dashboard-helpers'
  import Modal from './Modal.svelte'
  import Icon from './Icon.svelte'

  let {
    open = $bindable(false),
    group,
    dashboards = [],
    copyingDashboardId = '',
    onSelect,
  }: {
    open?: boolean
    group: Group | null
    dashboards?: DashListItem[]
    copyingDashboardId?: string
    onSelect: (dashboard: DashListItem) => void | Promise<void>
  } = $props()
</script>

<Modal bind:open title="Copy group" description={group ? `Choose a dashboard for “${group.title}”.` : ''}>
  {#if dashboards.length > 0}
    <div class="space-y-1">
      {#each dashboards as dashboard (dashboard.id)}
        <button
          type="button"
          disabled={!!copyingDashboardId}
          class="flex w-full items-center gap-3 rounded-btn border border-transparent px-3 py-2.5 text-left hover:border-border hover:bg-surface-2 disabled:cursor-wait disabled:opacity-60"
          onclick={() => onSelect(dashboard)}
        >
          {#if dashboard.icon}
            <Icon icon={dashboard.icon} iconDark={dashboard.iconDark} size={20} />
          {:else}
            <span class="flex h-5 w-5 items-center justify-center rounded bg-primary-soft text-[10px] font-semibold text-primary">D</span>
          {/if}
          <span class="min-w-0 flex-1">
            <span class="block truncate text-sm font-medium text-text">{dashboard.name}</span>
            {#if dashboard.ownerUsername}<span class="block truncate text-xs text-text-subtle">{dashboard.ownerUsername}</span>{/if}
          </span>
          {#if copyingDashboardId === dashboard.id}
            <span class="text-xs text-text-muted">Copying…</span>
          {:else}
            <span class="text-text-subtle" aria-hidden="true">›</span>
          {/if}
        </button>
      {/each}
    </div>
  {:else}
    <div class="rounded-btn bg-surface-2 px-4 py-6 text-center">
      <p class="text-sm text-text-muted">No other editable dashboards are available.</p>
    </div>
  {/if}

  <div class="mt-4 flex justify-end">
    <button
      type="button"
      disabled={!!copyingDashboardId}
      class="rounded-btn px-3 py-2 text-sm text-text-muted hover:bg-surface-2 disabled:opacity-50"
      onclick={() => (open = false)}>Cancel</button
    >
  </div>
</Modal>
