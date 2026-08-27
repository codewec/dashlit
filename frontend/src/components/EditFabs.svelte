<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import { onMount } from 'svelte';
  import { editMode } from '../lib/stores';

  let {
    raised = false,
    onCreateDashboard,
    onCloneDashboard,
    onExport,
    onImport,
    onNewGroup,
    onSettings,
    onDeleteDashboard,
    onSave,
  }: {
    raised?: boolean;
    onCreateDashboard: () => void;
    onCloneDashboard: () => void;
    onExport: () => void;
    onImport: () => void;
    onNewGroup: () => void;
    onSettings: () => void;
    onDeleteDashboard: () => void;
    onSave: () => void | Promise<void>;
  } = $props();

  let bottomOffset = $state(20);

  onMount(() => {
    if (!raised) return;

    const footer = document.querySelector('footer');
    if (!footer) return;

    const updateBottomOffset = () => {
      const visibleFooterHeight = Math.max(0, window.innerHeight - footer.getBoundingClientRect().top);
      bottomOffset = 20 + visibleFooterHeight;
    };

    updateBottomOffset();
    window.addEventListener('scroll', updateBottomOffset, { passive: true });
    window.addEventListener('resize', updateBottomOffset);

    return () => {
      window.removeEventListener('scroll', updateBottomOffset);
      window.removeEventListener('resize', updateBottomOffset);
    };
  });
</script>

<!-- Reserve room for the fixed controls on touch-sized screens. -->
<div class={$editMode ? 'h-32 sm:hidden' : 'h-20 sm:hidden'} aria-hidden="true"></div>

{#if $editMode}
  <div class="pointer-events-none fixed left-2 right-2 z-30 grid grid-cols-2 items-center gap-2 sm:left-4 sm:right-4 sm:flex sm:justify-between sm:gap-3" style:bottom="{bottomOffset}px">
    <div class="pointer-events-auto order-2 flex items-center gap-1.5 sm:order-none">
      <button
        type="button"
        class="flex h-11 items-center gap-1.5 rounded-full border border-border bg-surface px-3 text-xs font-medium text-text shadow-lg hover:bg-surface-2 sm:px-4"
        onclick={onCreateDashboard}
        title="New dashboard"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
        <span class="hidden sm:inline">New</span>
      </button>
      <button
        type="button"
        class="flex h-11 w-11 items-center justify-center rounded-full border border-border bg-surface text-text shadow-lg hover:bg-surface-2"
        onclick={onCloneDashboard}
        title="Clone dashboard"
      >
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
          ><rect x="9" y="9" width="13" height="13" rx="2" /><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" /></svg
        >
      </button>
      <DropdownMenu.Root>
        <DropdownMenu.Trigger class="flex h-11 w-11 items-center justify-center rounded-full border border-border bg-surface text-text shadow-lg hover:bg-surface-2" title="Import / Export">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
            ><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><path d="M7 10l5 5 5-5" /><path d="M12 15V3" /></svg
          >
        </DropdownMenu.Trigger>
        <DropdownMenu.Portal>
          <DropdownMenu.Content class="z-50 min-w-40 overflow-hidden rounded-xl border border-border bg-surface p-1 shadow-xl outline-none" sideOffset={8} side="top">
            <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={onExport}>
              Export
            </DropdownMenu.Item>
            <DropdownMenu.Item class="flex cursor-pointer items-center gap-2 rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={onImport}>
              Import
            </DropdownMenu.Item>
          </DropdownMenu.Content>
        </DropdownMenu.Portal>
      </DropdownMenu.Root>
    </div>

    <div class="pointer-events-auto order-1 col-span-2 flex items-center gap-1 justify-self-center rounded-full border border-border bg-surface p-1.5 shadow-lg sm:order-none sm:col-span-1">
      <button type="button" class="rounded-full px-3 py-2 text-xs font-medium text-text hover:bg-surface-2" onclick={onNewGroup}> + Group </button>
      <button type="button" class="rounded-full px-3 py-2 text-xs text-text-muted hover:bg-surface-2 hover:text-text" onclick={onSettings}> Settings </button>
      <button type="button" class="rounded-full bg-primary px-3 py-2 text-xs font-medium text-white hover:bg-primary-hover" onclick={() => onSave()}> Save </button>
    </div>

    <button
      type="button"
      class="pointer-events-auto order-3 flex h-11 w-11 items-center justify-center justify-self-end rounded-full border border-danger/40 bg-surface p-0 text-xs font-medium text-danger shadow-lg hover:bg-danger-soft sm:order-none sm:w-auto sm:gap-1.5 sm:px-4"
      onclick={onDeleteDashboard}
      title="Delete dashboard"
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18" /><path d="M8 6V4h8v2" /><path d="M19 6l-1 14H6L5 6" /></svg>
      <span class="hidden sm:inline">Delete</span>
    </button>
  </div>
{:else}
  <div class="group fixed bottom-5 left-1/2 z-30 flex h-11 w-28 -translate-x-1/2 items-end justify-center sm:bottom-0 sm:h-20">
    <button
      type="button"
      class="flex h-11 items-center rounded-full border border-border bg-surface px-5 text-xs font-medium text-text shadow-lg transition-transform duration-200 ease-out hover:bg-surface-2 sm:translate-y-[1.375rem] sm:group-hover:-translate-y-5 sm:group-focus-within:-translate-y-5"
      onclick={() => editMode.set(true)}
    >
      Edit
    </button>
  </div>
{/if}
