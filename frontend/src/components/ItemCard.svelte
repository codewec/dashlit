<script lang="ts">
  import { useSortable } from '@dnd-kit-svelte/svelte/sortable';
  import { DropdownMenu } from 'bits-ui';
  import Icon from './Icon.svelte';
  import type { Item, ItemSize } from '../lib/api';
  import { editMode } from '../lib/stores';
  import { cn } from '../lib/cn';
  import { api } from '../lib/api';

  let {
    item,
    index,
    groupId,
    itemSize = '1x1',
    isOverlay = false,
    canModify = true,
    onEdit,
    onDelete,
    onClone,
  }: {
    item: Item;
    index: number;
    groupId: string;
    itemSize?: ItemSize;
    isOverlay?: boolean;
    canModify?: boolean;
    onEdit?: (item: Item) => void;
    onDelete?: (item: Item) => void;
    onClone?: (item: Item) => void;
  } = $props();

  const { ref, handleRef, isDragging } = useSortable({
    id: () => item.id,
    index: () => index,
    type: 'item',
    accept: 'item',
    group: () => groupId,
    data: () => ({ group: groupId, item }),
    disabled: () => !$editMode || !canModify,
  });

  let pingReachable = $state<boolean | null>(null);

  $effect(() => {
    if (!item.pingEnabled || isOverlay) {
      pingReachable = null;
      return;
    }
    let active = true;
    const check = async () => {
      try {
        const result = await api.pingItem(item.id);
        if (active) pingReachable = result.reachable;
      } catch {
        if (active) pingReachable = false;
      }
    };
    void check();
    const interval = window.setInterval(check, 30_000);
    return () => {
      active = false;
      window.clearInterval(interval);
    };
  });
</script>

<div class="relative h-full w-full" {@attach ref}>
  <a
    href={$editMode ? undefined : item.url}
    target={$editMode ? undefined : '_blank'}
    rel="noopener"
    class={cn(
      'group relative flex border border-border-soft bg-surface transition',
      'hover:border-primary/40 hover:bg-surface-2',
      itemSize === '1x1' ? 'aspect-square w-full flex-col items-center justify-center rounded-2xl p-2' : 'min-h-[4.25rem] w-full items-center gap-3 rounded-2xl p-3',
      isDragging.current && !isOverlay && 'invisible',
      isOverlay && 'shadow-xl ring-2 ring-primary/30',
    )}
    onclick={(e) => $editMode && e.preventDefault()}
    title={itemSize === '1x1' ? item.title : undefined}
  >
    {#if $editMode}
      <button type="button" disabled={!canModify} class="absolute left-1 top-1 z-10 cursor-grab touch-none rounded p-0.5 text-text-subtle hover:bg-surface-2 hover:text-text-muted disabled:cursor-not-allowed disabled:opacity-40" {@attach handleRef} aria-label="Drag">
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"
          ><circle cx="9" cy="7" r="1.5" /><circle cx="15" cy="7" r="1.5" /><circle cx="9" cy="12" r="1.5" /><circle cx="15" cy="12" r="1.5" /><circle cx="9" cy="17" r="1.5" /><circle
            cx="15"
            cy="17"
            r="1.5"
          /></svg
        >
      </button>
    {/if}

    {#if item.pingEnabled && pingReachable !== null && (!item.pingOnlyDown || !pingReachable)}
      {#if itemSize === '1x1'}
        <span
          class="absolute right-2 top-2 z-[5] h-2.5 w-2.5 rounded-full shadow-sm ring-2 ring-surface {pingReachable ? 'bg-success' : 'bg-danger'}"
          title={pingReachable ? 'URL is available' : 'URL is unavailable'}
          aria-label={pingReachable ? 'Online' : 'Offline'}
        ></span>
      {:else}
        <span
          class="absolute right-1.5 top-1.5 z-[5] inline-flex items-center gap-1 rounded-full px-1.5 py-0.5 text-[9px] font-medium shadow-sm {pingReachable ? 'bg-success/15 text-success' : 'bg-danger-soft text-danger'}"
          title={pingReachable ? 'URL is available' : 'URL is unavailable'}
        >
          <span class="h-1.5 w-1.5 rounded-full {pingReachable ? 'bg-success' : 'bg-danger'}"></span>
          {pingReachable ? 'Online' : 'Offline'}
        </span>
      {/if}
    {/if}

    {#if itemSize === '1x1'}
      <Icon icon={item.icon} iconDark={item.iconDark} size={40} class="shrink-0 rounded-xl" />
    {:else}
      <Icon icon={item.icon} iconDark={item.iconDark} size={28} class="shrink-0 rounded-lg" />
      <div class="min-w-0 flex-1">
        <div class="truncate text-sm font-medium text-text">{item.title}</div>
        {#if item.description}
          <div class="truncate text-xs text-text-muted">{item.description}</div>
        {/if}
      </div>
    {/if}

    {#if $editMode}
      <div
        class={cn(
          'absolute z-10 opacity-100 transition sm:opacity-0 sm:group-hover:opacity-100 sm:focus-within:opacity-100',
          itemSize === '1x1' ? 'bottom-1 right-1' : 'bottom-1.5 right-1.5',
        )}
      >
        <DropdownMenu.Root>
          <DropdownMenu.Trigger disabled={!canModify} class="rounded-md bg-surface/95 p-1 text-text-muted shadow-sm ring-1 ring-border hover:text-text disabled:cursor-not-allowed disabled:opacity-40" aria-label="Item actions" onclick={(e) => e.preventDefault()}>
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><circle cx="5" cy="12" r="1.75" /><circle cx="12" cy="12" r="1.75" /><circle cx="19" cy="12" r="1.75" /></svg>
          </DropdownMenu.Trigger>
          <DropdownMenu.Portal>
            <DropdownMenu.Content class="z-50 min-w-36 overflow-hidden rounded-xl border border-border bg-surface p-1 shadow-xl outline-none" sideOffset={6} align="end">
              <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => onEdit?.(item)}>Edit</DropdownMenu.Item>
              <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => onClone?.(item)}>Clone</DropdownMenu.Item>
              <DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
              <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-danger outline-none data-[highlighted]:bg-danger-soft" onSelect={() => onDelete?.(item)}>Delete</DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Portal>
        </DropdownMenu.Root>
      </div>
    {/if}
  </a>

  {#if !isOverlay && isDragging.current}
    <div class="absolute inset-0 rounded-2xl border-2 border-dashed border-primary/50 bg-primary/5"></div>
  {/if}
</div>
