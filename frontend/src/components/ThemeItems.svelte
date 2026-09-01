<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { theme, setTheme } from '../lib/stores'
  import { themeOptions } from '../lib/themes'

  let { inset = false }: { inset?: boolean } = $props()
  const systemTheme = themeOptions[0]
  const lightThemes = themeOptions.filter((option) => option.mode === 'light')
  const darkThemes = themeOptions.filter((option) => option.mode === 'dark')
</script>

{#snippet optionItem(option: (typeof themeOptions)[number])}
  <DropdownMenu.Item
    class="flex cursor-pointer items-center gap-2 rounded-lg py-2 pr-2.5 text-sm text-text outline-none data-highlighted:bg-surface-2 {inset
      ? 'pl-7'
      : 'pl-2.5'}"
    onSelect={() => setTheme(option.value)}
  >
    <span class="h-3.5 w-3.5 shrink-0 rounded-full border border-border" style:background={option.swatch}></span>
    <span class="flex-1">{option.label}</span>
    {#if $theme === option.value}<span class="text-primary">✓</span>{/if}
  </DropdownMenu.Item>
{/snippet}

{@render optionItem(systemTheme)}
<DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
<div class="px-2.5 py-1 text-[10px] font-medium uppercase tracking-wide text-text-subtle">Light</div>
{#each lightThemes as option}
  {@render optionItem(option)}
{/each}
<DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
<div class="px-2.5 py-1 text-[10px] font-medium uppercase tracking-wide text-text-subtle">Dark</div>
{#each darkThemes as option}
  {@render optionItem(option)}
{/each}
