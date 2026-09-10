<script lang="ts">
  import { searchQuery } from '../lib/stores'

  let input: HTMLInputElement
</script>

<div class="relative w-full max-w-xs">
  <input
    bind:this={input}
    data-dashboard-filter
    type="text"
    aria-label="Filter dashboard"
    placeholder="Filter…"
    bind:value={$searchQuery}
    class="w-full rounded-btn border border-border bg-surface py-1.5 pl-3 pr-9 text-sm outline-none placeholder:text-text-subtle focus:border-primary"
  />
  {#if $searchQuery}
    <button
      type="button"
      aria-label="Clear filter"
      title="Clear filter (Esc)"
      class="absolute inset-y-0 right-0 flex w-9 items-center justify-center rounded-btn text-text-muted hover:text-text focus-visible:outline-primary"
      onclick={() => {
        searchQuery.set('')
        input.focus()
      }}
    >
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
        <path d="m6 6 12 12M6 18 18 6" />
      </svg>
    </button>
  {/if}
</div>
