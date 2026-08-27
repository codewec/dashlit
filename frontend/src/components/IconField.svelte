<script lang="ts">
  import { Dialog } from 'bits-ui';
  import Icon from './Icon.svelte';
  import IconPicker from './IconPicker.svelte';

  let {
    value = $bindable(''),
    valueDark = $bindable(''),
    defaultIcon = '',
  }: {
    value?: string;
    valueDark?: string;
    defaultIcon?: string;
  } = $props();

  let openLight = $state(false);
  let openDark = $state(false);
  let draft = $state('');

  function openPicker(which: 'light' | 'dark') {
    draft = which === 'light' ? value || defaultIcon : valueDark;
    if (which === 'light') openLight = true;
    else openDark = true;
  }

  function applyLight() {
    value = draft || defaultIcon;
    openLight = false;
  }
  function applyDark() {
    valueDark = draft;
    openDark = false;
  }

  function clearLight(e: MouseEvent) {
    e.stopPropagation();
    value = defaultIcon;
  }
  function clearDark(e: MouseEvent) {
    e.stopPropagation();
    valueDark = '';
  }
</script>

<div class="flex gap-3">
  <!-- primary / light -->
  <div class="flex flex-1 flex-col items-center gap-1">
    <span class="text-[11px] text-text-muted">Default</span>
    <button
      type="button"
      class="relative flex h-14 w-14 items-center justify-center rounded-xl border border-border bg-bg-elevated hover:border-primary"
      onclick={() => openPicker('light')}
      title="Choose default icon"
    >
      <Icon icon={value || defaultIcon} size={28} />
      {#if value && value !== defaultIcon}
        <span
          role="button"
          tabindex="0"
          class="absolute -right-1.5 -top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-surface text-[10px] text-text-muted ring-1 ring-border hover:text-danger"
          onclick={clearLight}
          onkeydown={(e) => e.key === 'Enter' && clearLight(e as any)}>×</span
        >
      {/if}
    </button>
  </div>

  <!-- dark optional -->
  <div class="flex flex-1 flex-col items-center gap-1">
    <span class="text-[11px] text-text-muted">Dark</span>
    <button
      type="button"
      class="relative flex h-14 w-14 items-center justify-center rounded-xl border border-dashed border-border bg-bg-elevated hover:border-primary"
      onclick={() => openPicker('dark')}
      title="Choose dark theme icon"
    >
      {#if valueDark}
        <Icon icon={valueDark} size={28} />
        <span
          role="button"
          tabindex="0"
          class="absolute -right-1.5 -top-1.5 flex h-5 w-5 items-center justify-center rounded-full bg-surface text-[10px] text-text-muted ring-1 ring-border hover:text-danger"
          onclick={clearDark}
          onkeydown={(e) => e.key === 'Enter' && clearDark(e as any)}>×</span
        >
      {:else}
        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" class="text-text-subtle">
          <rect x="3" y="3" width="18" height="18" rx="4" />
          <path d="M12 8v8M8 12h8" />
        </svg>
      {/if}
    </button>
  </div>
</div>

<!-- nested: light -->
<Dialog.Root bind:open={openLight}>
  <Dialog.Portal>
    <Dialog.Overlay class="dialog-overlay fixed inset-0 z-[60] bg-black/50" />
    <Dialog.Content class="dialog-content fixed left-1/2 top-1/2 z-[60] w-[min(22rem,calc(100vw-1.5rem))] -translate-x-1/2 -translate-y-1/2 rounded-card border border-border bg-surface p-4 shadow-2xl outline-none">
      <Dialog.Title class="text-sm font-semibold text-text">Default icon</Dialog.Title>
      <div class="mt-3">
        {#key openLight}
          <IconPicker bind:value={draft} />
        {/key}
      </div>
      <div class="mt-4 flex justify-end gap-2">
        <button type="button" class="rounded-btn px-3 py-1.5 text-xs text-text-muted" onclick={() => (openLight = false)}>Cancel</button>
        <button type="button" class="rounded-btn bg-primary px-3 py-1.5 text-xs font-medium text-white" onclick={applyLight}>Apply</button>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>

<!-- nested: dark -->
<Dialog.Root bind:open={openDark}>
  <Dialog.Portal>
    <Dialog.Overlay class="dialog-overlay fixed inset-0 z-[60] bg-black/50" />
    <Dialog.Content class="dialog-content fixed left-1/2 top-1/2 z-[60] w-[min(22rem,calc(100vw-1.5rem))] -translate-x-1/2 -translate-y-1/2 rounded-card border border-border bg-surface p-4 shadow-2xl outline-none">
      <Dialog.Title class="text-sm font-semibold text-text">Dark theme icon</Dialog.Title>
      <div class="mt-3">
        {#key openDark}
          <IconPicker bind:value={draft} />
        {/key}
      </div>
      <div class="mt-4 flex justify-end gap-2">
        <button type="button" class="rounded-btn px-3 py-1.5 text-xs text-text-muted" onclick={() => (openDark = false)}>Cancel</button>
        <button type="button" class="rounded-btn bg-primary px-3 py-1.5 text-xs font-medium text-white" onclick={applyDark}>Apply</button>
      </div>
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
