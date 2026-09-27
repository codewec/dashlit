<script lang="ts">
  import type { Snippet } from 'svelte'
  import logoUrl from '../assets/dashlit.svg'

  const currentYear = new Date().getFullYear()

  let { children }: { children?: Snippet } = $props()
</script>

<div class="flex min-h-dvh flex-col p-4">
  <div class="flex flex-1 items-center justify-center">
    <div class="w-full max-w-sm">
      {@render children?.()}
    </div>
  </div>

  <div class="flex items-center justify-center gap-2 pt-4 text-xs text-text-subtle">
    <img src={logoUrl} alt="" class="h-3.5 w-3.5 opacity-70" width="14" height="14" />
    <span>DashLit · © 2025 – {currentYear} ·</span>
    <a
      href="https://github.com/codewec/dashlit"
      target="_blank"
      rel="noopener noreferrer"
      class="opacity-50 transition-opacity hover:opacity-100"
      aria-label="DashLit on GitHub"
      title="GitHub"
    >
      <img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/codewec/dashlit" />
    </a>
  </div>
</div>
