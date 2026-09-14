<script lang="ts">
  import { DragDropProvider, DragOverlay, KeyboardSensor, PointerSensor } from '@dnd-kit-svelte/svelte'
  import { move } from '@dnd-kit/helpers'
  import { push } from '../lib/router'

  import type { Dashboard, Group, Item } from '../lib/api'
  import { editMode, user } from '../lib/stores'
  import {
    filterGroups,
    itemsByGroupMap,
    groupsOuterClass,
    groupCellClass,
    reorderGroups,
    applyItemMove,
    type DashListItem,
  } from '../lib/dashboard-helpers'
  import GroupCard from './GroupCard.svelte'
  import ItemCard from './ItemCard.svelte'
  import { searchQuery } from '../lib/stores'
  import { hotkeyKeyLabel, hotkeyMatches, modifierStateFromEvent, shouldShowHotkey, type ModifierState } from '../lib/hotkeys'

  let {
    dashboard,
    dashboards = [],
    groups = $bindable([]),
    canModify = true,
    onEditGroup,
    onDeleteGroup,
    onCloneGroup,
    onCopyGroupToDashboard,
    onAddItem,
    onEditItem,
    onDeleteItem,
    onCloneItem,
    onLayoutChange,
    onCreateFirstGroup,
  }: {
    dashboard: Dashboard
    dashboards?: DashListItem[]
    groups: Group[]
    canModify?: boolean
    onEditGroup: (g: Group) => void
    onDeleteGroup: (g: Group) => void
    onCloneGroup: (g: Group) => void
    onCopyGroupToDashboard: (g: Group) => void
    onAddItem: (g: Group) => void
    onEditItem: (item: Item) => void
    onDeleteItem: (item: Item) => void
    onCloneItem: (item: Item) => void
    onLayoutChange: () => void | Promise<void>
    onCreateFirstGroup?: () => void
  } = $props()

  const filtered = $derived(filterGroups(groups, $searchQuery))
  const byGroup = $derived(itemsByGroupMap(filtered))
  const outerClass = $derived(groupsOuterClass(dashboard.layout, dashboard.width === 'wide'))
  const cellClass = $derived(groupCellClass(dashboard.layout))
  // Sortables can only be activated through their dedicated handles, which
  // already use touch-action: none. Avoid the default 250 ms / 5 px touch
  // constraint: in mobile emulation (and during normal finger movement) it
  // commonly cancels the drag before it can start.
  const sensors = [PointerSensor.configure({ activationConstraints: () => undefined }), KeyboardSensor]

  let activeItemId = $state('')
  let heldModifiers = $state<ModifierState>({ ctrl: false, alt: false, shift: false, meta: false })
  const hasHeldModifier = $derived(heldModifiers.ctrl || heldModifiers.alt || heldModifiers.shift || heldModifiers.meta)
  const ownDashboards = $derived(dashboards.filter((candidate) => !!$user && candidate.ownerId === $user.id))
  const matchingDashboards = $derived(ownDashboards.filter((candidate) => candidate.hotkey && shouldShowHotkey(candidate.hotkey, heldModifiers)))
  const heldModifierLabels = $derived.by(() => {
    const labels: string[] = []
    if (heldModifiers.ctrl) labels.push('Ctrl')
    if (heldModifiers.alt) labels.push('Alt')
    if (heldModifiers.shift) labels.push('Shift')
    if (heldModifiers.meta) labels.push('Meta')
    return labels
  })

  // ID первого элемента на доске — точка входа для Tab, когда ничего не выбрано
  const firstItemId = $derived.by(() => {
    for (const group of filtered) {
      const items = byGroup[group.id]
      if (items?.length) return items[0].id
    }
    return ''
  })

  // Сбрасываем выделение, если отфильтрованный активный элемент исчезает
  const filteredItemIds = $derived.by(() => {
    const ids = new Set<string>()
    for (const group of filtered) {
      for (const item of byGroup[group.id] || []) ids.add(item.id)
    }
    return ids
  })

  $effect(() => {
    if (activeItemId && !filteredItemIds.has(activeItemId)) activeItemId = ''
  })

  $effect(() => {
    void dashboard.id
    activeItemId = ''
  })

  // При фильтрации первый результат сразу готов к открытию через Enter.
  $effect(() => {
    activeItemId = $searchQuery.trim() ? firstItemId : ''
  })

  $effect(() => {
    if ($editMode) {
      activeItemId = ''
      clearHeldModifiers()
    }
  })

  function clearSelection() {
    activeItemId = ''
  }

  function clearHeldModifiers() {
    heldModifiers = { ctrl: false, alt: false, shift: false, meta: false }
  }

  function handleWindowKeyup(event: KeyboardEvent) {
    if ($editMode) {
      clearHeldModifiers()
      return
    }
    heldModifiers = modifierStateFromEvent(event)
  }

  // Навигация по реальным DOM-позициям элементов — корректна при любом CSS-layout
  function getItemElements() {
    return Array.from(document.querySelectorAll<HTMLAnchorElement>('[data-dashboard-item]'))
  }

  function navigateByDOM(fromId: string, direction: 'left' | 'right' | 'up' | 'down'): string | null {
    const allEls = getItemElements()
    if (allEls.length === 0) return null

    const currentEl = allEls.find((element) => element.dataset.itemId === fromId)
    if (!currentEl) return allEls[0]?.dataset.itemId ?? null

    const cr = currentEl.getBoundingClientRect()
    const cx = (cr.left + cr.right) / 2
    const cy = (cr.top + cr.bottom) / 2

    const candidates = allEls
      .filter((el) => el.dataset.itemId !== fromId)
      .map((el) => {
        const rect = el.getBoundingClientRect()
        const dx = (rect.left + rect.right) / 2 - cx
        const dy = (rect.top + rect.bottom) / 2 - cy
        return { id: el.dataset.itemId!, dx, dy }
      })
      .filter(({ dx, dy }) => {
        if (direction === 'right') return dx > 1
        if (direction === 'left') return dx < -1
        if (direction === 'down') return dy > 1
        return dy < -1
      })

    if (candidates.length === 0) return null

    // Все карточки считаются точками по центру. Размеры 1x1 и 1x2 не влияют
    // на доступность кандидата; отклонение от оси лишь повышает его score.
    candidates.sort((a, b) => {
      const aPrimary = direction === 'left' || direction === 'right' ? Math.abs(a.dx) : Math.abs(a.dy)
      const bPrimary = direction === 'left' || direction === 'right' ? Math.abs(b.dx) : Math.abs(b.dy)
      const aSecondary = direction === 'left' || direction === 'right' ? Math.abs(a.dy) : Math.abs(a.dx)
      const bSecondary = direction === 'left' || direction === 'right' ? Math.abs(b.dy) : Math.abs(b.dx)
      return aPrimary + aSecondary * 2 - (bPrimary + bSecondary * 2)
    })

    return candidates[0]?.id ?? null
  }

  function handleWindowKeydown(event: KeyboardEvent) {
    if (event.isComposing) return

    const hasOpenOverlay = !!document.querySelector('[role="dialog"], [role="alertdialog"], [role="menu"], [role="listbox"]')

    // Открытый диалог обрабатывает Escape самостоятельно. В остальных случаях
    // Escape снимает выделение и завершает режим редактирования.
    if (event.key === 'Escape') {
      clearSelection()
      if ($editMode && !hasOpenOverlay) {
        event.preventDefault()
        editMode.set(false)
      }
      return
    }
    if ($editMode || hasOpenOverlay) {
      clearHeldModifiers()
      return
    }

    heldModifiers = modifierStateFromEvent(event)
    const target = event.target
    const isFilter = target instanceof HTMLElement && target.matches('[data-dashboard-filter]')
    if (
      target instanceof HTMLElement &&
      !isFilter &&
      target.closest('input, textarea, select, [contenteditable]:not([contenteditable="false"]), [role="textbox"]')
    )
      return

    if (!event.repeat && (event.ctrlKey || event.altKey || event.shiftKey || event.metaKey)) {
      const targetDashboard = ownDashboards.find((candidate) => candidate.hotkey && hotkeyMatches(event, candidate.hotkey))
      if (targetDashboard) {
        event.preventDefault()
        clearSelection()
        if (targetDashboard.slug !== dashboard.slug) push('/' + targetDashboard.slug)
        return
      }

      const matchingIds = filtered.flatMap((group) =>
        (byGroup[group.id] || []).filter((item) => item.hotkey && hotkeyMatches(event, item.hotkey)).map((item) => item.id),
      )
      if (matchingIds.length > 0) {
        const matchingItems = filtered.flatMap((group) => (byGroup[group.id] || []).filter((item) => matchingIds.includes(item.id)))
        event.preventDefault()
        event.stopPropagation()
        clearSelection()
        window.getSelection()?.removeAllRanges()
        for (const item of matchingItems) window.open(item.url, '_blank', 'noopener')
        requestAnimationFrame(() => window.getSelection()?.removeAllRanges())
        return
      }
    }

    if (event.defaultPrevented || event.ctrlKey || event.metaKey || event.altKey || event.shiftKey) return

    const dirs: Record<string, 'left' | 'right' | 'up' | 'down'> = {
      ArrowLeft: 'left',
      ArrowRight: 'right',
      ArrowUp: 'up',
      ArrowDown: 'down',
    }
    const direction = dirs[event.key]
    if (direction) {
      const elements = getItemElements()
      if (elements.length === 0) return
      event.preventDefault()

      if (!activeItemId) {
        activeItemId = elements[0]?.dataset.itemId ?? ''
        return
      }

      const nextId = navigateByDOM(activeItemId, direction)
      if (nextId) activeItemId = nextId
      return
    }

    if (event.key === 'Enter' && activeItemId) {
      const selected = getItemElements().find((element) => element.dataset.itemId === activeItemId)
      if (!selected) return
      event.preventDefault()
      clearSelection()
      selected.click()
      return
    }

    if (activeItemId) clearSelection()
  }

  function onDragOver(event: any) {
    if (!$editMode || !canModify) return
    const { source } = event.operation
    if (source?.type === 'column') {
      const targetId = event.operation.target?.id as string | undefined
      if (!targetId) return
      groups = reorderGroups(groups, source.id as string, targetId)
      return
    }
    const bag: Record<string, Item[]> = {}
    for (const g of groups) bag[g.id] = [...(g.items ?? [])]
    const nextBag = move(bag, event) as Record<string, Item[]>
    groups = applyItemMove(groups, nextBag)
  }

  async function onDragEnd() {
    if (!$editMode || !canModify) return
    await onLayoutChange()
  }
</script>

<svelte:window onkeydown={handleWindowKeydown} onkeyup={handleWindowKeyup} onpointerdown={clearSelection} onblur={clearHeldModifiers} />

<div
  class="pointer-events-none fixed left-1/2 z-40 -translate-x-1/2 items-center gap-1.5 rounded-xl border border-border bg-bg-elevated/95 px-3 py-2 shadow-xl backdrop-blur-md {hasHeldModifier &&
  !$editMode
    ? 'hidden sm:flex'
    : 'hidden'} {dashboard.cleanMode ? 'bottom-5' : 'bottom-16'}"
  aria-hidden="true"
>
  {#each heldModifierLabels as label, index}
    {#if index > 0}<span class="text-xs text-text-subtle">+</span>{/if}
    <kbd class="min-w-9 rounded-md border border-border bg-surface px-2 py-1 text-center text-xs font-semibold text-text shadow-sm">{label}</kbd>
  {/each}
</div>

{#if groups.length === 0}
  <div class="flex flex-col items-center gap-3 py-20 text-center">
    <p class="text-sm font-medium text-text">This dashboard is empty.</p>
    <p class="max-w-sm text-sm text-text-muted">Create the first group to start adding services and links.</p>
    {#if canModify}
      <button
        type="button"
        class="rounded-btn bg-primary px-4 py-2 text-sm font-medium text-white hover:bg-primary-hover"
        onclick={onCreateFirstGroup}
      >
        Add group
      </button>
    {/if}
  </div>
{:else}
  <div>
    <DragDropProvider {sensors} {onDragOver} {onDragEnd}>
      <div class={outerClass}>
        {#each filtered as group, gIndex (group.id)}
          <div class={cellClass} data-dashboard-group={group.id}>
            <GroupCard
              {group}
              index={gIndex}
              layout={dashboard.layout}
              wide={dashboard.width === 'wide'}
              {canModify}
              onEdit={onEditGroup}
              onDelete={onDeleteGroup}
              onClone={onCloneGroup}
              onCopyTo={onCopyGroupToDashboard}
              {onAddItem}
            >
              {#each byGroup[group.id] || [] as item, iIndex (item.id)}
                <ItemCard
                  {item}
                  index={iIndex}
                  groupId={group.id}
                  itemSize={group.itemSize}
                  tabIndex={activeItemId ? (item.id === activeItemId ? 0 : -1) : item.id === firstItemId ? 0 : -1}
                  isKeyboardActive={!!activeItemId && item.id === activeItemId}
                  hotkeyHint={!$editMode && shouldShowHotkey(item.hotkey, heldModifiers) ? hotkeyKeyLabel(item.hotkey, item.hotkeyLabel) : ''}
                  isHotkeyDimmed={!$editMode && hasHeldModifier && !shouldShowHotkey(item.hotkey, heldModifiers)}
                  {canModify}
                  onEdit={onEditItem}
                  onDelete={onDeleteItem}
                  onClone={onCloneItem}
                />
              {/each}
            </GroupCard>
          </div>
        {/each}
      </div>

      <DragOverlay>
        {#snippet children(source)}
          {#if source?.data?.item}
            {@const g = groups.find((x) => x.id === source.data.group)}
            <ItemCard item={source.data.item} index={0} groupId={source.data.group} itemSize={g?.itemSize ?? '1x1'} isOverlay />
          {:else if source?.data?.group}
            <GroupCard group={source.data.group} index={0} layout={dashboard.layout} wide={dashboard.width === 'wide'} isOverlay />
          {/if}
        {/snippet}
      </DragOverlay>
    </DragDropProvider>
    {#if matchingDashboards.length > 0 && !$editMode}
      <aside class="fixed right-4 top-20 z-30 hidden w-60 sm:block">
        <div class="rounded-card border border-border-soft bg-surface/95 p-3 shadow-xl backdrop-blur-md">
          <div class="mb-2 text-[10px] font-semibold uppercase tracking-wider text-text-subtle">Dashboard shortcuts</div>
          <div class="space-y-1.5">
            {#each matchingDashboards as candidate}
              <div
                class="flex items-center gap-2 rounded-btn px-2.5 py-2 {candidate.slug === dashboard.slug
                  ? 'bg-primary/10 text-primary'
                  : 'bg-surface-2 text-text'}"
              >
                <kbd class="min-w-7 shrink-0 rounded border border-border bg-bg-elevated px-1.5 py-0.5 text-center text-xs font-semibold"
                  >{hotkeyKeyLabel(candidate.hotkey, candidate.hotkeyLabel)}</kbd
                >
                <span class="min-w-0 flex-1 truncate text-xs font-medium">{candidate.name}</span>
                {#if candidate.slug === dashboard.slug}<span class="text-[9px] text-text-subtle">current</span>{/if}
              </div>
            {/each}
          </div>
        </div>
      </aside>
    {/if}
  </div>
{/if}
