<script lang="ts">
  import { AlertDialog } from 'bits-ui';

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

<AlertDialog.Root bind:open>
  <AlertDialog.Portal>
    <AlertDialog.Overlay class="dialog-overlay fixed inset-0 z-50 bg-black/55" />
    <AlertDialog.Content class="dialog-content fixed left-1/2 top-1/2 z-50 w-[min(28rem,calc(100vw-1.5rem))] -translate-x-1/2 -translate-y-1/2 rounded-card border border-border bg-surface p-5 shadow-2xl outline-none">
      <AlertDialog.Title class="text-lg font-semibold tracking-tight text-text">{title}</AlertDialog.Title>
      <AlertDialog.Description class="mt-1 text-sm text-text-muted">{message}</AlertDialog.Description>
      <div class="mt-5 flex justify-end gap-2">
        <AlertDialog.Cancel class="rounded-btn border border-border bg-transparent px-3.5 py-2 text-sm text-text hover:bg-surface-2" disabled={loading}>Cancel</AlertDialog.Cancel>
        <AlertDialog.Action
          type="button"
          class="rounded-btn px-3.5 py-2 text-sm font-medium text-white {danger ? 'bg-danger hover:opacity-90' : 'bg-primary hover:bg-primary-hover'}"
          onclick={(event) => {
            event.preventDefault();
            void confirm();
          }}
          disabled={loading}
        >
          {loading ? '…' : confirmLabel}
        </AlertDialog.Action>
      </div>
    </AlertDialog.Content>
  </AlertDialog.Portal>
</AlertDialog.Root>
