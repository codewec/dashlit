<script lang="ts">
  import { iconSrc, resolveIcon } from '../lib/api';
  import { resolvedTheme } from '../lib/stores';

  let { icon = '', iconDark = '', size = 24, alt = '', class: className = '' }: { icon?: string; iconDark?: string; size?: number; alt?: string; class?: string } = $props();

  const isDark = $derived(!['crema', 'latte'].includes($resolvedTheme));
  const resolved = $derived(resolveIcon(icon, iconDark, isDark));
  const src = $derived(iconSrc(resolved));
</script>

{#if src}
  <img {src} width={size} height={size} {alt} class="shrink-0 object-contain {className}" style="width:{size}px;height:{size}px" loading="lazy" />
{:else}
  <span class="inline-flex shrink-0 items-center justify-center rounded-md bg-surface-2 text-text-subtle {className}" style="width:{size}px;height:{size}px;font-size:11px">?</span>
{/if}
