<script lang="ts">
  import { iconSrc, isIconifyIcon, resolveIcon } from '../lib/icon-helpers';
  import { resolvedTheme } from '../lib/stores';

  let { icon = '', iconDark = '', size = 24, alt = '', theme = 'auto', class: className = '' }: { icon?: string; iconDark?: string; size?: number; alt?: string; theme?: 'auto' | 'light' | 'dark'; class?: string } = $props();

  const isDark = $derived(theme === 'dark' || (theme === 'auto' && !['crema', 'latte'].includes($resolvedTheme)));
  const resolved = $derived(resolveIcon(icon, iconDark, isDark));
  const src = $derived(iconSrc(resolved));
  const isIconify = $derived(isIconifyIcon(resolved));
</script>

{#if src}
  <img {src} width={size} height={size} {alt} class="shrink-0 object-contain {isDark && isIconify ? 'brightness-0 invert' : ''} {className}" style="width:{size}px;height:{size}px" loading="lazy" />
{:else}
  <span class="inline-flex shrink-0 items-center justify-center rounded-md bg-surface-2 text-text-subtle {className}" style="width:{size}px;height:{size}px;font-size:11px">?</span>
{/if}
