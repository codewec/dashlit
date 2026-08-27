<script lang="ts">
  import Modal from './Modal.svelte';

  let {
    open = $bindable(false),
    title = 'Confirm',
    message = 'Are you sure?',
    confirmLabel = 'Delete',
    danger = true,
    onConfirm,
  }: {
    open?: boolean;
    title?: string;
    message?: string;
    confirmLabel?: string;
    danger?: boolean;
    onConfirm?: () => void | Promise<void>;
  } = $props();

  let loading = $state(false);

  async function confirm() {
    loading = true;
    try {
      await onConfirm?.();
      open = false;
    } finally {
      loading = false;
    }
  }
</script>

<Modal bind:open {title} description={message}>
  {#snippet footer()}
    <button type="button" class="rounded-btn border border-border bg-transparent px-3.5 py-2 text-sm text-text hover:bg-surface-2" onclick={() => (open = false)} disabled={loading}> Cancel </button>
    <button
      type="button"
      class="rounded-btn px-3.5 py-2 text-sm font-medium text-white {danger ? 'bg-danger hover:opacity-90' : 'bg-primary hover:bg-primary-hover'}"
      onclick={confirm}
      disabled={loading}
    >
      {loading ? '…' : confirmLabel}
    </button>
  {/snippet}
</Modal>
