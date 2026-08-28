<script lang="ts">
  import type { GroupForm } from '../lib/dashboard-helpers';
  import Modal from './Modal.svelte';
  import IconField from './IconField.svelte';

  let {
    open = $bindable(false),
    form = $bindable(),
    editing = false,
    onSave,
  }: {
    open?: boolean;
    form: GroupForm;
    editing?: boolean;
    onSave: () => void | Promise<void>;
  } = $props();

  let titleErr = $state(false);

  async function submit(e: Event) {
    e.preventDefault();
    titleErr = !form.title.trim();
    if (titleErr) return;
    await onSave();
  }
</script>

<Modal bind:open title={editing ? 'Edit group' : 'New group'}>
  <form class="space-y-3" onsubmit={submit}>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Title</span>
      <input class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary {titleErr ? 'field-error' : ''}" bind:value={form.title} />
      {#if titleErr}<span class="mt-1 text-xs text-danger">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Description</span>
      <input class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary" bind:value={form.description} />
    </label>
    <div>
      <span class="mb-1 block text-xs text-text-muted">Icon</span>
      <IconField bind:value={form.icon} bind:valueDark={form.iconDark} />
    </div>
    <div>
      <span class="mb-1 block text-xs text-text-muted">Item size in this group</span>
      <div class="grid grid-cols-2 gap-2">
        <button type="button" class="flex min-h-20 items-center justify-center gap-3 rounded-btn border px-3 py-2 text-sm transition {form.itemSize === '1x1' ? 'border-primary bg-primary/10 text-primary' : 'border-border bg-bg-elevated text-text-muted hover:bg-surface-2'}" onclick={() => (form.itemSize = '1x1')}>
          <span class="h-8 w-8 rounded-md border-2 border-current" aria-hidden="true"></span>
          <span>1×1</span>
        </button>
        <button type="button" class="flex min-h-20 items-center justify-center gap-3 rounded-btn border px-3 py-2 text-sm transition {form.itemSize === '1x2' ? 'border-primary bg-primary/10 text-primary' : 'border-border bg-bg-elevated text-text-muted hover:bg-surface-2'}" onclick={() => (form.itemSize = '1x2')}>
          <span class="h-8 w-14 rounded-md border-2 border-current" aria-hidden="true"></span>
          <span>1×2</span>
        </button>
      </div>
    </div>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={() => (open = false)}>Cancel</button>
      <button type="submit" class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>
