<script lang="ts">
  import { useSortable } from '@dnd-kit-svelte/svelte/sortable';
  import { CollisionPriority } from '@dnd-kit/abstract';
  import type { Snippet } from 'svelte';
  import type { Group } from '../lib/api';
  import { editMode } from '../lib/stores';
  import { cn } from '../lib/cn';

  let {
    group,
    index,
    isOverlay = false,
    children,
    onEdit,
    onDelete,
    onAddItem,
  }: {
    group: Group;
    index: number;
    isOverlay?: boolean;
    children?: Snippet;
    onEdit?: (g: Group) => void;
    onDelete?: (g: Group) => void;
    onAddItem?: (g: Group) => void;
  } = $props();

  const { ref, handleRef, isDragging } = useSortable({
    id: group.id,
    index: () => index,
    type: 'column',
    accept: ['item', 'column'],
    collisionPriority: CollisionPriority.Low,
    data: () => ({ group }),
    disabled: () => !$editMode,
  });
</script>

<section class="relative" {@attach ref}>
  <div
    class={cn(
      'rounded-[var(--radius-card)] border border-[var(--color-border-soft)] bg-[var(--color-surface)]/80 p-4 backdrop-blur-sm',
      isDragging.current && !isOverlay && 'invisible',
      isOverlay && 'shadow-2xl ring-2 ring-[var(--color-primary)]/25'
    )}
  >
    <header class="mb-3 flex items-center gap-2">
      {#if $editMode}
        <button
          type="button"
          class="cursor-grab touch-none text-[var(--color-text-subtle)]"
          {@attach handleRef}
          aria-label="Drag group"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><circle cx="9" cy="7" r="1.5"/><circle cx="15" cy="7" r="1.5"/><circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/><circle cx="9" cy="17" r="1.5"/><circle cx="15" cy="17" r="1.5"/></svg>
        </button>
      {/if}
      <h3 class="min-w-0 flex-1 truncate text-sm font-semibold tracking-tight text-[var(--color-text)]">
        {group.title}
      </h3>
      {#if $editMode}
        <div class="flex items-center gap-0.5">
          <button type="button" class="rounded-md p-1 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-text)]" onclick={() => onAddItem?.(group)} title="Add item">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg>
          </button>
          <button type="button" class="rounded-md p-1 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-text)]" onclick={() => onEdit?.(group)} title="Edit group">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>
          </button>
          <button type="button" class="rounded-md p-1 text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-danger)]" onclick={() => onDelete?.(group)} title="Delete group">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M19 6l-1 14H6L5 6"/></svg>
          </button>
        </div>
      {/if}
    </header>
    <div class="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
      {@render children?.()}
    </div>
  </div>
  {#if !isOverlay && isDragging.current}
    <div class="absolute inset-0 rounded-[var(--radius-card)] border-2 border-dashed border-[var(--color-primary)]/40 bg-[var(--color-primary)]/5"></div>
  {/if}
</section>
