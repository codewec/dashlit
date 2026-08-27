<script lang="ts">
  import { iconSrc, resolveIcon } from '../lib/api';
  import { theme } from '../lib/stores';

  let { icon = '', iconDark = '', size = 24, alt = '', class: className = '' }: { icon?: string; iconDark?: string; size?: number; alt?: string; class?: string } = $props();

  const isDark = $derived($theme === 'dark' || ($theme === 'system' && typeof window !== 'undefined' && window.matchMedia('(prefers-color-scheme: dark)').matches));
  const resolved = $derived(resolveIcon(icon, iconDark, isDark));
  const src = $derived(iconSrc(resolved));
</script>

{#if src}
  <img {src} width={size} height={size} {alt} class="shrink-0 object-contain {className}" style="width:{size}px;height:{size}px" loading="lazy" />
{:else}
  <span
    class="inline-flex shrink-0 items-center justify-center rounded-md bg-[var(--color-surface-2)] text-[var(--color-text-subtle)] {className}"
    style="width:{size}px;height:{size}px;font-size:11px">?</span
  >
{/if}
