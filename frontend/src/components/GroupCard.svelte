<script lang="ts">
  import { useSortable } from '@dnd-kit-svelte/svelte/sortable';
  import { CollisionPriority } from '@dnd-kit/abstract';
  import { DropdownMenu } from 'bits-ui';
  import type { Snippet } from 'svelte';
  import type { Group, Layout } from '../lib/api';
  import { editMode } from '../lib/stores';
  import { cn } from '../lib/cn';
  import Icon from './Icon.svelte';

  let {
    group,
    index,
    layout = 'rows',
    wide = false,
    isOverlay = false,
    canModify = true,
    children,
    onEdit,
    onDelete,
    onClone,
    onCopyTo,
    onAddItem,
  }: {
    group: Group;
    index: number;
    layout?: Layout;
    wide?: boolean;
    isOverlay?: boolean;
    canModify?: boolean;
    children?: Snippet;
    onEdit?: (g: Group) => void;
    onDelete?: (g: Group) => void;
    onClone?: (g: Group) => void;
    onCopyTo?: (g: Group) => void;
    onAddItem?: (g: Group) => void;
  } = $props();

  const { ref, handleRef, isDragging } = useSortable({
    id: () => group.id,
    index: () => index,
    type: 'column',
    accept: ['item', 'column'],
    collisionPriority: CollisionPriority.Low,
    data: () => ({ group }),
    disabled: () => !$editMode || !canModify,
  });

  // Item grid is the same visual language everywhere.
  // 1x1: fixed min size so squares don't grow with container width.
  // 1x2: rows → 4 (default) / 6 (wide); columns & masonry → one per row.
  const itemsClass = $derived.by(() => {
    if (group.itemSize === '1x1') {
      return 'grid grid-cols-[repeat(auto-fill,minmax(4.5rem,1fr))] gap-2';
    }
    if (layout === 'columns' || layout === 'masonry') {
      return 'grid grid-cols-1 gap-2';
    }
    // rows + 1x2
    return wide ? 'grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6' : 'grid grid-cols-1 gap-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4';
  });

  function addFirstItem() {
    editMode.set(true);
    onAddItem?.(group);
  }
</script>

<section class="relative h-full w-full" {@attach ref}>
  <div
    class={cn(
      'relative h-full rounded-card border border-border-soft bg-surface/80 p-4 pt-5 backdrop-blur-sm',
      isDragging.current && !isOverlay && 'invisible',
      isOverlay && 'shadow-2xl ring-2 ring-primary/25',
    )}
  >
    {#if $editMode}
      <button
        type="button"
        class="absolute left-1.5 top-1.5 z-10 cursor-grab touch-none rounded p-0.5 text-text-subtle hover:bg-surface-2 hover:text-text-muted disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
        {@attach handleRef}
        disabled={!canModify}
        aria-label="Drag group"
      >
        <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor"
          ><circle cx="9" cy="7" r="1.5" /><circle cx="15" cy="7" r="1.5" /><circle cx="9" cy="12" r="1.5" /><circle cx="15" cy="12" r="1.5" /><circle cx="9" cy="17" r="1.5" /><circle
            cx="15"
            cy="17"
            r="1.5"
          /></svg
        >
      </button>
    {/if}

    <header class={cn('mb-3 flex gap-2', group.description ? 'items-start' : 'items-center', $editMode && 'pr-14')}>
      {#if group.icon}
        <Icon icon={group.icon} iconDark={group.iconDark} size={22} class={cn('shrink-0 rounded-md', group.description && 'mt-0.5')} />
      {/if}
      <div class="min-w-0 flex-1">
        <h3 class="truncate text-sm font-semibold tracking-tight text-text">{group.title}</h3>
        {#if group.description}
          <p class="mt-0.5 line-clamp-2 text-xs text-text-muted">{group.description}</p>
        {/if}
      </div>
      {#if $editMode}
        <div class="absolute right-3 top-4 flex items-center gap-0.5">
          <button
            type="button"
            disabled={!canModify}
            class="rounded-md p-1 text-text-muted hover:bg-surface-2 hover:text-text disabled:cursor-not-allowed disabled:opacity-40"
            onclick={() => onAddItem?.(group)}
            title="Add item"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
          </button>
          <DropdownMenu.Root>
            <DropdownMenu.Trigger
              disabled={!canModify}
              class="rounded-md p-1 text-text-muted hover:bg-surface-2 hover:text-text disabled:cursor-not-allowed disabled:opacity-40"
              title="Group actions"
              aria-label="Group actions"
            >
              <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"
                ><circle cx="5" cy="12" r="1.75" /><circle cx="12" cy="12" r="1.75" /><circle cx="19" cy="12" r="1.75" /></svg
              >
            </DropdownMenu.Trigger>
            <DropdownMenu.Portal>
              <DropdownMenu.Content class="z-50 min-w-36 overflow-hidden rounded-xl border border-border bg-surface p-1 shadow-xl outline-none" sideOffset={6} align="end">
                <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => onEdit?.(group)}>Edit</DropdownMenu.Item
                >
                <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => onClone?.(group)}
                  >Clone</DropdownMenu.Item
                >
                <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-text outline-none data-[highlighted]:bg-surface-2" onSelect={() => onCopyTo?.(group)}
                  >Copy to dashboard…</DropdownMenu.Item
                >
                <DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
                <DropdownMenu.Item class="cursor-pointer rounded-lg px-2.5 py-2 text-sm text-danger outline-none data-[highlighted]:bg-danger-soft" onSelect={() => onDelete?.(group)}
                  >Delete</DropdownMenu.Item
                >
              </DropdownMenu.Content>
            </DropdownMenu.Portal>
          </DropdownMenu.Root>
        </div>
      {/if}
    </header>
    {#if !isOverlay && (group.items?.length ?? 0) === 0}
      <div class="flex flex-col items-center gap-2 pb-8 text-center">
        <p class="text-xs text-text-muted">This group is empty.</p>
        {#if canModify}
          <button type="button" class="rounded-btn bg-primary px-3 py-1.5 text-xs font-medium text-white hover:bg-primary-hover" onclick={addFirstItem}> Add item </button>
        {/if}
      </div>
    {:else}
      <div class={itemsClass}>
        {@render children?.()}
      </div>
    {/if}
  </div>
  {#if !isOverlay && isDragging.current}
    <div class="absolute inset-0 rounded-card border-2 border-dashed border-primary/40 bg-primary/5"></div>
  {/if}
</section>
