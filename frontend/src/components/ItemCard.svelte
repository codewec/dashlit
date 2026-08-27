<script lang="ts">
  import { useSortable } from '@dnd-kit-svelte/svelte/sortable';
  import Icon from './Icon.svelte';
  import type { Item } from '../lib/api';
  import { editMode } from '../lib/stores';
  import { cn } from '../lib/cn';

  let {
    item,
    index,
    groupId,
    isOverlay = false,
    onEdit,
    onDelete,
  }: {
    item: Item;
    index: number;
    groupId: string;
    isOverlay?: boolean;
    onEdit?: (item: Item) => void;
    onDelete?: (item: Item) => void;
  } = $props();

  const { ref, handleRef, isDragging } = useSortable({
    id: item.id,
    index: () => index,
    type: 'item',
    accept: 'item',
    group: () => groupId,
    data: () => ({ group: groupId, item }),
    disabled: () => !$editMode,
  });
</script>

<div class="relative" {@attach ref}>
  <a
    href={$editMode ? undefined : item.url}
    target={$editMode ? undefined : '_blank'}
    rel="noopener"
    class={cn(
      'group relative flex items-center gap-3 rounded-2xl border border-[var(--color-border-soft)] bg-[var(--color-surface)] p-3 transition',
      'hover:border-[var(--color-primary)]/40 hover:bg-[var(--color-surface-2)]',
      item.size === '1x2' && 'min-h-[4.5rem]',
      isDragging.current && !isOverlay && 'invisible',
      isOverlay && 'shadow-xl ring-2 ring-[var(--color-primary)]/30'
    )}
    onclick={(e) => $editMode && e.preventDefault()}
  >
    {#if $editMode}
      <button
        type="button"
        class="cursor-grab touch-none text-[var(--color-text-subtle)] hover:text-[var(--color-text-muted)]"
        {@attach handleRef}
        aria-label="Drag"
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><circle cx="9" cy="7" r="1.5"/><circle cx="15" cy="7" r="1.5"/><circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/><circle cx="9" cy="17" r="1.5"/><circle cx="15" cy="17" r="1.5"/></svg>
      </button>
    {/if}

    <Icon icon={item.icon} size={item.size === '1x2' ? 36 : 28} class="shrink-0 rounded-lg" />

    <div class="min-w-0 flex-1">
      <div class="truncate text-sm font-medium text-[var(--color-text)]">{item.title}</div>
      {#if item.size === '1x2' && item.description}
        <div class="truncate text-xs text-[var(--color-text-muted)]">{item.description}</div>
      {/if}
    </div>

    {#if $editMode}
      <div class="absolute right-1.5 top-1.5 flex gap-0.5 opacity-0 transition group-hover:opacity-100 focus-within:opacity-100">
        <button
          type="button"
          class="rounded-md bg-[var(--color-surface)]/95 p-1 text-[var(--color-text-muted)] shadow-sm ring-1 ring-[var(--color-border)] hover:text-[var(--color-text)]"
          onclick={(e) => { e.preventDefault(); onEdit?.(item); }}
          aria-label="Edit"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z"/></svg>
        </button>
        <button
          type="button"
          class="rounded-md bg-[var(--color-surface)]/95 p-1 text-[var(--color-text-muted)] shadow-sm ring-1 ring-[var(--color-border)] hover:text-[var(--color-danger)]"
          onclick={(e) => { e.preventDefault(); onDelete?.(item); }}
          aria-label="Delete"
        >
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18"/><path d="M8 6V4h8v2"/><path d="M19 6l-1 14H6L5 6"/></svg>
        </button>
      </div>
    {/if}
  </a>

  {#if !isOverlay && isDragging.current}
    <div class="absolute inset-0 rounded-2xl border-2 border-dashed border-[var(--color-primary)]/50 bg-[var(--color-primary)]/5"></div>
  {/if}
</div>
