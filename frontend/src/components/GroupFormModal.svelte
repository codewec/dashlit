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
    <label class="block">
      <span class="mb-1 block text-xs text-text-muted">Item size in this group</span>
      <select class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm" bind:value={form.itemSize}>
        <option value="1x1">1×1 — icon only</option>
        <option value="1x2">1×2 — icon, title, description</option>
      </select>
    </label>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={() => (open = false)}>Cancel</button>
      <button type="submit" class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>
