<script lang="ts">
  import { iconSrc } from '../lib/api';

  let {
    icon = '',
    size = 24,
    alt = '',
    class: className = '',
  }: { icon?: string; size?: number; alt?: string; class?: string } = $props();

  const src = $derived(iconSrc(icon));
</script>

{#if src}
  <img
    {src}
    width={size}
    height={size}
    alt={alt}
    class="object-contain {className}"
    loading="lazy"
  />
{:else}
  <span
    class="inline-flex items-center justify-center rounded-md bg-[var(--color-surface-2)] text-[var(--color-text-subtle)] {className}"
    style="width:{size}px;height:{size}px;font-size:11px"
  >?</span>
{/if}
