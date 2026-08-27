<script lang="ts">
  import { useSortable } from '@dnd-kit-svelte/svelte/sortable';
  import { CollisionPriority } from '@dnd-kit/abstract';
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
    children,
    onEdit,
    onDelete,
    onClone,
    onAddItem,
  }: {
    group: Group;
    index: number;
    layout?: Layout;
    wide?: boolean;
    isOverlay?: boolean;
    children?: Snippet;
    onEdit?: (g: Group) => void;
    onDelete?: (g: Group) => void;
    onClone?: (g: Group) => void;
    onAddItem?: (g: Group) => void;
  } = $props();

  const { ref, handleRef, isDragging } = useSortable({
    id: () => group.id,
    index: () => index,
    type: 'column',
    accept: ['item', 'column'],
    collisionPriority: CollisionPriority.Low,
    data: () => ({ group }),
    disabled: () => !$editMode,
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
        class="absolute left-1.5 top-1.5 z-10 cursor-grab touch-none rounded p-0.5 text-text-subtle hover:bg-surface-2 hover:text-text-muted"
        {@attach handleRef}
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

    <header class="mb-3 flex items-start gap-2">
      {#if group.icon}
        <Icon icon={group.icon} iconDark={group.iconDark} size={22} class="mt-0.5 shrink-0 rounded-md" />
      {/if}
      <div class="min-w-0 flex-1">
        <h3 class="truncate text-sm font-semibold tracking-tight text-text">{group.title}</h3>
        {#if group.description}
          <p class="mt-0.5 line-clamp-2 text-xs text-text-muted">{group.description}</p>
        {/if}
      </div>
      {#if $editMode}
        <div class="flex shrink-0 items-center gap-0.5">
          <button type="button" class="rounded-md p-1 text-text-muted hover:bg-surface-2 hover:text-text" onclick={() => onAddItem?.(group)} title="Add item">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
          </button>
          <button type="button" class="rounded-md p-1 text-text-muted hover:bg-surface-2 hover:text-text" onclick={() => onEdit?.(group)} title="Edit group">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9" /><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" /></svg>
          </button>
          <button type="button" class="rounded-md p-1 text-text-muted hover:bg-surface-2 hover:text-text" onclick={() => onClone?.(group)} title="Clone group">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
              ><rect x="9" y="9" width="13" height="13" rx="2" /><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" /></svg
            >
          </button>
          <button type="button" class="rounded-md p-1 text-text-muted hover:bg-surface-2 hover:text-danger" onclick={() => onDelete?.(group)} title="Delete group">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18" /><path d="M8 6V4h8v2" /><path d="M19 6l-1 14H6L5 6" /></svg>
          </button>
        </div>
      {/if}
    </header>
    <div class={itemsClass}>
      {@render children?.()}
    </div>
  </div>
  {#if !isOverlay && isDragging.current}
    <div class="absolute inset-0 rounded-card border-2 border-dashed border-primary/40 bg-primary/5"></div>
  {/if}
</section>
