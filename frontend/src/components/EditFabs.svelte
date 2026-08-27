<script lang="ts">
  import { editMode } from '../lib/stores';

  let {
    dashboardName = '',
    raised = false,
    onCreateDashboard,
    onNewGroup,
    onSettings,
    onDeleteDashboard,
  }: {
    dashboardName?: string;
    raised?: boolean;
    onCreateDashboard: () => void;
    onNewGroup: () => void;
    onSettings: () => void;
    onDeleteDashboard: () => void;
  } = $props();

  const bottom = $derived(raised ? 'bottom-16' : 'bottom-5');
</script>

{#if $editMode}
  <div class="pointer-events-none fixed {bottom} left-4 right-4 z-30 flex items-center justify-between gap-3">
    <button
      type="button"
      class="pointer-events-auto flex h-11 items-center gap-1.5 rounded-full border border-border bg-surface px-4 text-xs font-medium text-text shadow-lg hover:bg-surface-2"
      onclick={onCreateDashboard}
      title="New dashboard"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
      <span class="hidden sm:inline">Dashboard</span>
    </button>

    <div class="pointer-events-auto flex items-center gap-1 rounded-full border border-border bg-surface p-1.5 shadow-lg">
      <button type="button" class="rounded-full px-3 py-2 text-xs font-medium text-text hover:bg-surface-2" onclick={onNewGroup}> + Group </button>
      <button type="button" class="rounded-full px-3 py-2 text-xs text-text-muted hover:bg-surface-2 hover:text-text" onclick={onSettings}> Settings </button>
      <button type="button" class="rounded-full bg-primary px-3 py-2 text-xs font-medium text-white hover:bg-primary-hover" onclick={() => editMode.set(false)}> Save </button>
    </div>

    <button
      type="button"
      class="pointer-events-auto flex h-11 items-center gap-1.5 rounded-full border border-danger/40 bg-surface px-4 text-xs font-medium text-danger shadow-lg hover:bg-danger-soft"
      onclick={onDeleteDashboard}
      title="Delete dashboard"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18" /><path d="M8 6V4h8v2" /><path d="M19 6l-1 14H6L5 6" /></svg>
      <span class="hidden sm:inline">Delete</span>
    </button>
  </div>
{:else}
  <div class="fixed {bottom} left-1/2 z-30 -translate-x-1/2">
    <button
      type="button"
      class="flex h-11 items-center rounded-full border border-border bg-surface px-5 text-xs font-medium text-text shadow-lg hover:bg-surface-2"
      onclick={() => editMode.set(true)}
    >
      Edit
    </button>
  </div>
{/if}
