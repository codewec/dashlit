<script lang="ts">
  import type { Snippet } from 'svelte'
  let { children }: { children?: Snippet } = $props()
</script>

<div class="flex min-h-dvh items-center justify-center p-4">
  <div class="w-full max-w-sm">
    {@render children?.()}
  </div>
</div>
