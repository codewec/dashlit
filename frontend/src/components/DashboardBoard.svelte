<script lang="ts">
  import { DragDropProvider, DragOverlay, KeyboardSensor, PointerSensor } from '@dnd-kit-svelte/svelte'
  import { move } from '@dnd-kit/helpers'

  import type { Dashboard, Group, Item } from '../lib/api'
  import { editMode } from '../lib/stores'
  import { filterGroups, itemsByGroupMap, groupsOuterClass, groupCellClass, reorderGroups, applyItemMove } from '../lib/dashboard-helpers'
  import GroupCard from './GroupCard.svelte'
  import ItemCard from './ItemCard.svelte'
  import { searchQuery } from '../lib/stores'

  let {
    dashboard,
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
  let boardElement = $state<HTMLElement | null>(null)

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
    if ($editMode) activeItemId = ''
  })

  function clearSelection() {
    activeItemId = ''
  }

  // Навигация по реальным DOM-позициям элементов — корректна при любом CSS-layout
  function getItemElements() {
    return boardElement ? Array.from(boardElement.querySelectorAll<HTMLAnchorElement>('[data-dashboard-item]')) : []
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
    if (event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return

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
    if ($editMode || event.defaultPrevented || hasOpenOverlay) return

    const target = event.target
    const isFilter = target instanceof HTMLElement && target.matches('[data-dashboard-filter]')
    if (
      target instanceof HTMLElement &&
      !isFilter &&
      target.closest('input, textarea, select, [contenteditable]:not([contenteditable="false"]), [role="textbox"]')
    ) return

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

<svelte:window onkeydown={handleWindowKeydown} onpointerdown={clearSelection} />

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
  <DragDropProvider {sensors} {onDragOver} {onDragEnd}>
    <div class={outerClass} bind:this={boardElement}>
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
                tabIndex={activeItemId ? (item.id === activeItemId ? 0 : -1) : (item.id === firstItemId ? 0 : -1)}
                isKeyboardActive={!!activeItemId && item.id === activeItemId}
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
{/if}
