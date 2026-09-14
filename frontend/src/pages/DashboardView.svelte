<script lang="ts">
  import { tick, untrack } from 'svelte'
  import { derived } from 'svelte/store'
  import { push, replace } from '../lib/router'
  import { api, type Dashboard, type Group, type Item } from '../lib/api'
  import { user, editMode, currentDashboard, searchQuery } from '../lib/stores'
  import DashboardFilter from '../components/DashboardFilter.svelte'
  import {
    type DashListItem,
    type GroupForm,
    type ItemForm,
    type DashboardForm,
    toDashList,
    emptyGroupForm,
    groupToForm,
    emptyItemForm,
    itemToForm,
    emptyDashboardForm,
    dashboardToForm,
    buildLayoutPayload,
    pageContainerClass,
  } from '../lib/dashboard-helpers'
  import AppLayout from '../layouts/AppLayout.svelte'
  import DashboardBoard from '../components/DashboardBoard.svelte'
  import EditFabs from '../components/EditFabs.svelte'
  import CleanHeader from '../components/CleanHeader.svelte'
  import GroupFormModal from '../components/GroupFormModal.svelte'
  import ItemFormModal from '../components/ItemFormModal.svelte'
  import DashboardFormModal from '../components/DashboardFormModal.svelte'
  import ConfirmModal from '../components/ConfirmModal.svelte'
  import CopyGroupModal from '../components/CopyGroupModal.svelte'
  import { toast, toastError } from '../lib/toasts'

  let { params = { slug: '' } } = $props<{ params?: { slug?: string } }>()

  let dashboard = $state<Dashboard | null>(null)
  let groups = $state<Group[]>([])
  let loading = $state(true)
  let error = $state('')
  let notFound = $state(false)
  let dashList = $state<DashListItem[]>([])

  let editingGroup = $state<Group | null>(null)
  let editingItem = $state<Item | null>(null)

  let groupOpen = $state(false)
  let itemOpen = $state(false)
  let dashOpen = $state(false)
  let groupForm = $state<GroupForm>(emptyGroupForm())
  let itemForm = $state<ItemForm>(emptyItemForm())
  let dashForm = $state<DashboardForm>(emptyDashboardForm())

  let confirmOpen = $state(false)
  let confirmMsg = $state('')
  let confirmAction = $state<(() => Promise<void>) | null>(null)
  let copyGroupOpen = $state(false)
  let groupToCopy = $state<Group | null>(null)
  let copyingDashboardId = $state('')
  const authenticatedUserId = derived(user, (currentUser) => currentUser?.id ?? null)

  async function handleFilterKeydown(event: KeyboardEvent) {
    if (loading || !dashboard || event.defaultPrevented || event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return
    if (groupOpen || itemOpen || dashOpen || confirmOpen || copyGroupOpen) return
    if (document.querySelector('[role="dialog"], [role="alertdialog"], [role="menu"], [role="listbox"]')) return

    const target = event.target
    const isFilter = target instanceof HTMLElement && target.matches('[data-dashboard-filter]')
    if (
      target instanceof HTMLElement &&
      !isFilter &&
      target.closest('input, textarea, select, [contenteditable]:not([contenteditable="false"]), [role="textbox"]')
    )
      return

    if (event.key === 'Escape') {
      searchQuery.set('')
      event.preventDefault()
      return
    }
    if (isFilter || event.key.length !== 1 || !/\S/u.test(event.key)) return

    event.preventDefault()
    searchQuery.update((query) => query + event.key)
    await tick()
    const input = document.querySelector<HTMLInputElement>('[data-dashboard-filter]')
    input?.focus()
    input?.setSelectionRange(input.value.length, input.value.length)
  }
  const canModifyDashboard = $derived(!!dashboard && !!$user && (dashboard.ownerId === $user.id || $user.role === 'admin'))
  const copyGroupTargets = $derived(
    dashList.filter((candidate) => candidate.id !== dashboard?.id && !!$user && (candidate.ownerId === $user.id || $user.role === 'admin')),
  )
  const usedItemHotkeys = $derived(
    groups.flatMap((group) => (group.items ?? []).filter((item) => item.id !== editingItem?.id && !!item.hotkey).map((item) => item.hotkey)),
  )
  const ownDashboards = $derived(dashList.filter((candidate) => candidate.ownerId === $user?.id))
  const dashboardHotkeys = $derived(ownDashboards.filter((candidate) => !!candidate.hotkey).map((candidate) => candidate.hotkey))
  const dashboardFormConflicts = $derived([
    ...(dashboard?.ownerId === $user?.id
      ? groups.flatMap((group) => (group.items ?? []).filter((item) => !!item.hotkey).map((item) => item.hotkey))
      : []),
    ...ownDashboards
      .filter((candidate) => !!candidate.hotkey && (dashForm.creating || candidate.id !== dashboard?.id))
      .map((candidate) => candidate.hotkey),
  ])

  function askConfirm(message: string, action: () => Promise<void>) {
    confirmMsg = message
    confirmAction = action
    confirmOpen = true
  }

  async function load() {
    loading = true
    error = ''
    notFound = false
    try {
      if ($user) {
        dashList = toDashList(await api.listDashboards())
      } else {
        dashList = []
      }

      const slug = params?.slug
      let d: Dashboard | null = null

      notFound = false
      if (slug) {
        try {
          d = await api.getDashboard(slug)
        } catch (e: unknown) {
          const msg = e instanceof Error ? e.message : ''
          if (/login|unauthorized|access denied/i.test(msg)) {
            replace('/login')
            return
          }
          notFound = true
          dashboard = null
          groups = []
          return
        }
      } else {
        const owned = $user ? dashList.filter((candidate) => candidate.ownerId === $user?.id) : []
        if (owned.length > 0) {
          const preferred = owned.find((candidate) => candidate.isDefault) ?? owned[0]
          replace('/' + preferred.slug)
          return
        } else {
          try {
            d = await api.getMain()
          } catch {
            d = null
          }
          if (!d && $user && dashList.length > 0) {
            replace('/' + dashList[0].slug)
            return
          }
          if (!d && !$user) {
            replace('/login')
            return
          }
        }
      }

      if (!d) {
        if (!$user) {
          replace('/login')
          return
        }
        error = 'No dashboards yet.'
        dashboard = null
        groups = []
        currentDashboard.set(null)
      } else {
        dashboard = d
        groups = (d.groups ?? []).slice().sort((a, b) => a.position - b.position)
        currentDashboard.set(d)
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to load'
      if (/login|unauthorized|access denied/i.test(msg)) {
        replace('/login')
        return
      }
      error = msg
      dashboard = null
      groups = []
    } finally {
      loading = false
    }
  }

  $effect(() => {
    void params?.slug
    void $authenticatedUserId
    untrack(() => void load())
  })

  /* —— group —— */
  function openNewGroup() {
    editingGroup = null
    groupForm = emptyGroupForm()
    groupOpen = true
  }

  function createFirstGroup() {
    editMode.set(true)
    openNewGroup()
  }
  function openEditGroup(g: Group) {
    editingGroup = g
    groupForm = groupToForm(g)
    groupOpen = true
  }
  async function saveGroup() {
    if (!dashboard) return
    const payload = {
      title: groupForm.title.trim(),
      description: groupForm.description,
      icon: groupForm.icon,
      iconDark: groupForm.iconDark,
      itemSize: groupForm.itemSize,
      openInNewTab: groupForm.openInNewTab,
    }
    try {
      if (editingGroup) {
        const updated = await api.updateGroup(editingGroup.id, payload)
        groups = groups.map((group) => (group.id === updated.id ? { ...updated, items: group.items ?? [] } : group))
      } else {
        const created = await api.createGroup(dashboard.id, { ...payload, position: groups.length })
        groups = [...groups, { ...created, items: created.items ?? [] }]
      }
      groupOpen = false
    } catch (e: unknown) {
      toastError(e, editingGroup ? 'Could not update group' : 'Could not create group')
    }
  }

  async function deleteGroup(group: Group) {
    await api.deleteGroup(group.id)
    groups = groups.filter((candidate) => candidate.id !== group.id)
  }

  /* —— item —— */
  function openNewItem(g: Group) {
    editingItem = null
    itemForm = emptyItemForm(g.id)
    itemOpen = true
  }
  function openEditItem(item: Item) {
    editingItem = item
    itemForm = itemToForm(item)
    itemOpen = true
  }
  async function saveItem() {
    const payload = {
      title: itemForm.title.trim(),
      description: itemForm.description,
      url: itemForm.url.trim(),
      icon: itemForm.icon || 'mdi:link',
      iconDark: itemForm.iconDark,
      pingEnabled: itemForm.pingEnabled,
      pingOnlyDown: itemForm.pingOnlyDown,
      pingUrl: itemForm.pingUrl.trim(),
      pingSkipTls: itemForm.pingSkipTls,
      hotkey: itemForm.hotkey,
      hotkeyLabel: itemForm.hotkeyLabel,
      openInNewTab: itemForm.openInNewTab,
    }
    try {
      let saved: Item
      if (editingItem) saved = await api.updateItem(editingItem.id, payload)
      else {
        const pos = groups.find((g) => g.id === itemForm.groupId)?.items?.length ?? 0
        saved = await api.createItem(itemForm.groupId, { ...payload, position: pos })
      }
      groups = groups.map((group) => {
        const items = group.items ?? []
        if (group.id === saved.groupId) return { ...group, items: [...items.filter((item) => item.id !== saved.id), saved] }
        if (items.some((item) => item.id === saved.id)) return { ...group, items: items.filter((item) => item.id !== saved.id) }
        return group
      })
      itemOpen = false
    } catch (e: unknown) {
      toastError(e, editingItem ? 'Could not update item' : 'Could not create item')
    }
  }

  async function deleteItem(item: Item) {
    await api.deleteItem(item.id)
    groups = groups.map((group) =>
      group.id === item.groupId ? { ...group, items: (group.items ?? []).filter((candidate) => candidate.id !== item.id) } : group,
    )
  }

  /* —— dashboard —— */
  function openDashSettings() {
    dashForm = dashboard ? dashboardToForm(dashboard) : emptyDashboardForm(false)
    dashOpen = true
  }
  function openCreateDashboard() {
    dashForm = emptyDashboardForm(true)
    dashOpen = true
  }
  async function saveDash() {
    if (!dashForm.name.trim() || !dashForm.slug.trim()) return
    const body = {
      name: dashForm.name,
      slug: dashForm.slug,
      description: dashForm.description,
      icon: dashForm.icon,
      iconDark: dashForm.iconDark,
      privacy: dashForm.privacy,
      layout: dashForm.layout,
      width: dashForm.width,
      cleanMode: dashForm.cleanMode,
      openInNewTab: dashForm.openInNewTab,
      hotkey: dashForm.hotkey,
      hotkeyLabel: dashForm.hotkeyLabel,
    }
    try {
      if (dashForm.creating || !dashboard) {
        if (!$user) return
        const d = await api.createDashboard(body)
        if (dashForm.isDefault) await api.setDefault(d.id, true)
        dashOpen = false
        toast.success('Dashboard created')
        push('/' + d.slug)
        return
      }
      const prevSlug = dashboard.slug
      const wasDefault = dashboard.isDefault
      const updated = await api.updateDashboard(dashboard.id, body)
      if (dashboard.ownerId === $user?.id && dashForm.isDefault !== wasDefault) {
        await api.setDefault(dashboard.id, dashForm.isDefault)
      }
      const nextIsDefault = dashboard.ownerId === $user?.id ? dashForm.isDefault : wasDefault
      const nextDashboard = { ...dashboard, ...updated, isDefault: nextIsDefault }
      dashboard = nextDashboard
      currentDashboard.set(nextDashboard)
      dashList = dashList.map((candidate) => {
        if (nextIsDefault && candidate.ownerId === nextDashboard.ownerId) candidate = { ...candidate, isDefault: false }
        return candidate.id === nextDashboard.id
          ? {
              ...candidate,
              name: nextDashboard.name,
              slug: nextDashboard.slug,
              description: nextDashboard.description,
              icon: nextDashboard.icon,
              iconDark: nextDashboard.iconDark,
              hotkey: nextDashboard.hotkey,
              hotkeyLabel: nextDashboard.hotkeyLabel,
              isDefault: nextDashboard.isDefault,
            }
          : candidate
      })
      dashOpen = false
      if (dashForm.slug !== prevSlug) push('/' + dashForm.slug)
    } catch (e: unknown) {
      toastError(e, dashForm.creating ? 'Could not create dashboard' : 'Could not update dashboard')
    }
  }

  async function deleteDashboard() {
    if (!dashboard) return
    const id = dashboard.id
    await api.deleteDashboard(id)
    editMode.set(false)
    const rest = dashList.filter((d) => d.id !== id)
    dashList = rest
    if (rest.length > 0) push('/' + rest[0].slug)
    else {
      dashboard = null
      groups = []
      replace('/')
    }
  }

  async function cloneDashboard() {
    if (!dashboard) return
    try {
      const d = await api.cloneDashboard(dashboard.id)
      toast.success('Dashboard cloned')
      push('/' + d.slug)
    } catch (e: unknown) {
      toastError(e, 'Could not clone dashboard')
    }
  }

  async function exportDashboard() {
    if (!dashboard) return
    try {
      const blob = await api.exportDashboard(dashboard.id)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${dashboard.slug}.dashlit.json`
      a.click()
      URL.revokeObjectURL(url)
      toast.success('Dashboard exported')
    } catch (e: unknown) {
      toastError(e, 'Could not export dashboard')
    }
  }

  function importDashboard() {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'application/json,.json'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (!file) return
      try {
        const text = await file.text()
        const data = JSON.parse(text)
        const d = await api.importDashboard(data)
        toast.success('Dashboard imported')
        push('/' + d.slug)
      } catch (e: unknown) {
        toastError(e, 'Could not import dashboard')
      }
    }
    input.click()
  }

  async function cloneGroup(g: Group) {
    try {
      const created = await api.cloneGroup(g.id)
      groups = [...groups, created]
      await tick()

      const element = document.querySelector<HTMLElement>(`[data-dashboard-group="${CSS.escape(created.id)}"]`)
      if (element) {
        const bounds = element.getBoundingClientRect()
        if (bounds.top < 0 || bounds.bottom > window.innerHeight) {
          element.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
        }
      }

      toast.success('Group cloned')
    } catch (e: unknown) {
      toastError(e, 'Could not clone group')
    }
  }

  function openCopyGroup(g: Group) {
    groupToCopy = g
    copyingDashboardId = ''
    copyGroupOpen = true
  }

  async function copyGroupToDashboard(target: DashListItem) {
    if (!groupToCopy) return
    copyingDashboardId = target.id
    try {
      await api.cloneGroupToDashboard(groupToCopy.id, target.id)
      copyGroupOpen = false
      toast.success(`Group copied to ${target.name}`)
    } catch (e: unknown) {
      toastError(e, 'Could not copy group')
    } finally {
      copyingDashboardId = ''
    }
  }

  async function cloneItem(item: Item) {
    try {
      const created = await api.cloneItem(item.id)
      groups = groups.map((x) => (x.id === created.groupId ? { ...x, items: [...(x.items ?? []), created] } : x))
    } catch (e: unknown) {
      toastError(e, 'Could not clone item')
    }
  }

  async function persistLayout() {
    if (!dashboard) return
    try {
      await api.updateLayout(dashboard.id, buildLayoutPayload(groups))
    } catch (e: unknown) {
      toastError(e, 'Could not save dashboard layout')
    }
  }
</script>

<svelte:window onkeydown={handleFilterKeydown} />

{#if loading}
  <div class="flex min-h-dvh items-center justify-center text-sm text-text-subtle">Loading…</div>
{:else if notFound}
  <div class="flex min-h-dvh flex-col items-center justify-center gap-3 px-4 text-center">
    <p class="text-5xl font-semibold tracking-tight text-text">404</p>
    <p class="text-sm text-text-muted">Page not found</p>
    <a href="/" class="mt-2 text-sm text-primary hover:underline">Go to home</a>
  </div>
{:else if dashboard}
  {@const d = dashboard}
  {#if d.cleanMode}
    {#if $searchQuery}
      <div class="fixed left-1/2 top-4 z-40 w-[min(20rem,calc(100vw-2rem))] -translate-x-1/2 rounded-btn bg-surface p-2 shadow-xl">
        <DashboardFilter />
      </div>
    {/if}
    <div class={pageContainerClass(d.width === 'wide', !!$user && $editMode)}>
      <CleanHeader dashboard={d} dashboards={dashList} />
      <DashboardBoard
        dashboard={d}
        dashboards={ownDashboards}
        canModify={canModifyDashboard}
        bind:groups
        onEditGroup={openEditGroup}
        onDeleteGroup={(g) =>
          askConfirm(`Delete group “${g.title}” and all its items?`, async () => {
            await deleteGroup(g)
          })}
        onCloneGroup={cloneGroup}
        onCopyGroupToDashboard={openCopyGroup}
        onAddItem={openNewItem}
        onEditItem={openEditItem}
        onDeleteItem={(it) =>
          askConfirm(`Delete “${it.title}”?`, async () => {
            await deleteItem(it)
          })}
        onCloneItem={cloneItem}
        onLayoutChange={persistLayout}
        onCreateFirstGroup={createFirstGroup}
      />
      {#if $user}
        <EditFabs
          hideEditButton
          canModify={canModifyDashboard}
          onCreateDashboard={openCreateDashboard}
          onCloneDashboard={cloneDashboard}
          onExport={exportDashboard}
          onImport={importDashboard}
          onNewGroup={openNewGroup}
          onSettings={openDashSettings}
          onDeleteDashboard={() => askConfirm(`Delete dashboard “${d.name}”?`, deleteDashboard)}
          onSave={() => editMode.set(false)}
        />
      {/if}
    </div>
  {:else}
    <AppLayout dashboards={dashList} currentSlug={d.slug} wide={d.width === 'wide'} reserveControls={!!$user} showEdit>
      <DashboardBoard
        dashboard={d}
        dashboards={ownDashboards}
        canModify={canModifyDashboard}
        bind:groups
        onEditGroup={openEditGroup}
        onDeleteGroup={(g) =>
          askConfirm(`Delete group “${g.title}” and all its items?`, async () => {
            await deleteGroup(g)
          })}
        onCloneGroup={cloneGroup}
        onCopyGroupToDashboard={openCopyGroup}
        onAddItem={openNewItem}
        onEditItem={openEditItem}
        onDeleteItem={(it) =>
          askConfirm(`Delete “${it.title}”?`, async () => {
            await deleteItem(it)
          })}
        onCloneItem={cloneItem}
        onLayoutChange={persistLayout}
        onCreateFirstGroup={createFirstGroup}
      />
      {#if $user}
        <EditFabs
          raised
          hideEditButton
          canModify={canModifyDashboard}
          onCreateDashboard={openCreateDashboard}
          onCloneDashboard={cloneDashboard}
          onExport={exportDashboard}
          onImport={importDashboard}
          onNewGroup={openNewGroup}
          onSettings={openDashSettings}
          onDeleteDashboard={() => askConfirm(`Delete dashboard “${d.name}”?`, deleteDashboard)}
          onSave={() => editMode.set(false)}
        />
      {/if}
    </AppLayout>
  {/if}
{:else}
  <AppLayout dashboards={dashList} currentSlug="">
    <div class="flex flex-col items-center gap-3 py-20">
      <p class="text-sm text-text-muted">{error || 'No dashboards yet.'}</p>
      {#if $user}
        <div class="flex flex-wrap items-center justify-center gap-2">
          <button
            type="button"
            class="flex items-center gap-2 rounded-btn bg-primary px-4 py-2 text-sm text-white hover:bg-primary-hover"
            onclick={openCreateDashboard}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"
              ><path d="M12 5v14M5 12h14" /></svg
            >
            Create dashboard
          </button>
          <button
            type="button"
            class="flex items-center gap-2 rounded-btn border border-border bg-surface px-4 py-2 text-sm text-text hover:bg-surface-2"
            onclick={importDashboard}
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"
              ><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><path d="m17 8-5-5-5 5" /><path d="M12 3v12" /></svg
            >
            Import dashboard
          </button>
        </div>
      {/if}
    </div>
  </AppLayout>
{/if}

<GroupFormModal bind:open={groupOpen} bind:form={groupForm} editing={!!editingGroup} onSave={saveGroup} />
<ItemFormModal bind:open={itemOpen} bind:form={itemForm} editing={!!editingItem} itemHotkeys={usedItemHotkeys} {dashboardHotkeys} onSave={saveItem} />
<DashboardFormModal
  bind:open={dashOpen}
  bind:form={dashForm}
  canSetDefault={!!$user && (dashForm.creating || dashboard?.ownerId === $user.id)}
  conflictingHotkeys={dashboardFormConflicts}
  onSave={saveDash}
/>
<ConfirmModal
  bind:open={confirmOpen}
  title="Confirm"
  message={confirmMsg}
  onConfirm={async () => {
    await confirmAction?.()
  }}
/>
<CopyGroupModal bind:open={copyGroupOpen} group={groupToCopy} dashboards={copyGroupTargets} {copyingDashboardId} onSelect={copyGroupToDashboard} />
