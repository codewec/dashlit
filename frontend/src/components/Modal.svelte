<script lang="ts">
  import { Dialog } from 'bits-ui';
  import type { Snippet } from 'svelte';
  import { cn } from '../lib/cn';

  let {
    open = $bindable(false),
    title = '',
    description = '',
    class: className = '',
    children,
    footer,
    onOpenChange,
  }: {
    open?: boolean;
    title?: string;
    description?: string;
    class?: string;
    children?: Snippet;
    footer?: Snippet;
    onOpenChange?: (v: boolean) => void;
  } = $props();

  function handleOpenChange(v: boolean) {
    open = v;
    onOpenChange?.(v);
  }
</script>

<Dialog.Root bind:open onOpenChange={handleOpenChange}>
  <Dialog.Portal>
    <Dialog.Overlay class="dialog-overlay fixed inset-0 z-50 bg-black/55" />
    <Dialog.Content
      class={cn(
        'dialog-content fixed left-1/2 top-1/2 z-50 w-[min(28rem,calc(100vw-1.5rem))] -translate-x-1/2 -translate-y-1/2',
        'rounded-card border border-border bg-surface p-5 shadow-2xl outline-none',
        'max-h-[calc(100dvh-1.5rem)] overflow-y-auto',
        className,
      )}
    >
      {#if title}
        <Dialog.Title class="text-lg font-semibold tracking-tight text-text">
          {title}
        </Dialog.Title>
      {/if}
      {#if description}
        <Dialog.Description class="mt-1 text-sm text-text-muted">
          {description}
        </Dialog.Description>
      {/if}
      <div class="mt-4">
        {@render children?.()}
      </div>
      {#if footer}
        <div class="mt-5 flex justify-end gap-2">
          {@render footer()}
        </div>
      {/if}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
