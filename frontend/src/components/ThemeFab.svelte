<script lang="ts">
  import { theme, applyTheme } from '../lib/stores';

  function toggleTheme() {
    const order: Array<'light' | 'dark' | 'system'> = ['dark', 'light', 'system'];
    const next = order[(order.indexOf($theme) + 1) % order.length];
    theme.set(next);
    applyTheme(next);
  }
</script>

<button
  type="button"
  class="fixed bottom-5 right-5 z-40 flex h-10 w-10 items-center justify-center rounded-full border border-border bg-surface text-base text-text-muted shadow-lg hover:bg-surface-2 hover:text-text"
  onclick={toggleTheme}
  title="Theme: {$theme}"
  aria-label="Toggle theme"
>
  {$theme === 'dark' ? '☾' : $theme === 'light' ? '☀' : '◐'}
</button>
