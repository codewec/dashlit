<script lang="ts">
  import { DropdownMenu } from 'bits-ui';
  import { theme } from '../lib/stores';
  import { themeLabel } from '../lib/themes';
  import ThemeItems from './ThemeItems.svelte';
</script>

<DropdownMenu.Root>
  <DropdownMenu.Trigger
    class="fixed bottom-5 right-5 z-40 flex h-10 w-10 items-center justify-center rounded-full border border-border bg-surface text-text-muted shadow-lg hover:bg-surface-2 hover:text-text"
    title="Theme: {themeLabel($theme)}"
    aria-label="Choose theme"
  >
    <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><circle cx="12" cy="12" r="9" /><path d="M12 3a9 9 0 0 0 0 18c-2.2-2.5-2.2-15.5 0-18Z" /></svg>
  </DropdownMenu.Trigger>
  <DropdownMenu.Portal>
    <DropdownMenu.Content class="z-50 min-w-44 rounded-xl border border-border bg-surface p-1 shadow-xl outline-none" sideOffset={8} side="top" align="end">
      <ThemeItems />
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>
