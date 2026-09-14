<script lang="ts">
  import { DropdownMenu } from 'bits-ui'
  import { theme, setTheme, selectCustomTheme, openCustomThemeEditor, customTheme, hasCustomTheme } from '../lib/stores'
  import { customThemeSwatch, themeOptions } from '../lib/themes'

  let { inset = false }: { inset?: boolean } = $props()
  const systemTheme = themeOptions[0]
  const lightThemes = themeOptions.filter((option) => option.mode === 'light')
  const darkThemes = themeOptions.filter((option) => option.mode === 'dark')
  const customOption = themeOptions.find((option) => option.value === 'custom')

  const customSwatch = $derived(customThemeSwatch($customTheme))
  const itemPad = $derived(inset ? 'pl-7' : 'pl-2.5')
</script>

{#snippet optionItem(option: (typeof themeOptions)[number], swatch = option.swatch)}
  <DropdownMenu.Item
    class="flex cursor-pointer items-center gap-2 rounded-lg py-2 pr-2.5 text-sm text-text outline-none data-highlighted:bg-surface-2 {itemPad}"
    onSelect={() => {
      if (option.value === 'custom') selectCustomTheme()
      else setTheme(option.value)
    }}
  >
    <span class="h-3.5 w-3.5 shrink-0 rounded-full border border-border" style:background={swatch}></span>
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
<DropdownMenu.Separator class="my-1 h-px bg-border-soft" />
{#if $hasCustomTheme && customOption}
  {@render optionItem(customOption, customSwatch)}
{/if}
<DropdownMenu.Item
  class="flex cursor-pointer items-center gap-2 rounded-lg py-2 pr-2.5 text-sm text-text outline-none data-highlighted:bg-surface-2 {itemPad}"
  onSelect={() => openCustomThemeEditor()}
>
  <svg class="h-3.5 w-3.5 shrink-0" viewBox="0 0 16 16" aria-hidden="true">
    <path d="M0 0h16v16H0z" fill="none" />
    <path
      fill="currentColor"
      d="M15.825.12a.5.5 0 0 1 .132.584c-1.53 3.43-4.743 8.17-7.095 10.64a6.1 6.1 0 0 1-2.373 1.534c-.018.227-.06.538-.16.868c-.201.659-.667 1.479-1.708 1.74a8.1 8.1 0 0 1-3.078.132a4 4 0 0 1-.562-.135a1.4 1.4 0 0 1-.466-.247a.7.7 0 0 1-.204-.288a.62.62 0 0 1 .004-.443c.095-.245.316-.38.461-.452c.394-.197.625-.453.867-.826c.095-.144.184-.297.287-.472l.117-.198c.151-.255.326-.54.546-.848c.528-.739 1.201-.925 1.746-.896q.19.012.348.048c.062-.172.142-.38.238-.608c.261-.619.658-1.419 1.187-2.069c2.176-2.67 6.18-6.206 9.117-8.104a.5.5 0 0 1 .596.04M4.705 11.912a1.2 1.2 0 0 0-.419-.1c-.246-.013-.573.05-.879.479c-.197.275-.355.532-.5.777l-.105.177c-.106.181-.213.362-.32.528a3.4 3.4 0 0 1-.76.861c.69.112 1.736.111 2.657-.12c.559-.139.843-.569.993-1.06a3 3 0 0 0 .126-.75zm1.44.026c.12-.04.277-.1.458-.183a5.1 5.1 0 0 0 1.535-1.1c1.9-1.996 4.412-5.57 6.052-8.631c-2.59 1.927-5.566 4.66-7.302 6.792c-.442.543-.795 1.243-1.042 1.826c-.121.288-.214.54-.275.72v.001l.575.575zm-4.973 3.04l.007-.005zm3.582-3.043l.002.001h-.002z"
    />
  </svg>
  <span class="flex-1">Edit colors…</span>
</DropdownMenu.Item>
