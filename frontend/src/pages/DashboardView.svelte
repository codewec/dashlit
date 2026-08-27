<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { push, replace } from 'svelte-spa-router';
  import { api, type Dashboard, type Group, type Item } from '../lib/api';
  import { user, editMode, currentDashboard } from '../lib/stores';
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
  } from '../lib/dashboard-helpers';
  import AppLayout from '../layouts/AppLayout.svelte';
  import DashboardBoard from '../components/DashboardBoard.svelte';
  import EditFabs from '../components/EditFabs.svelte';
  import CleanHeader from '../components/CleanHeader.svelte';
  import GroupFormModal from '../components/GroupFormModal.svelte';
  import ItemFormModal from '../components/ItemFormModal.svelte';
  import DashboardFormModal from '../components/DashboardFormModal.svelte';
  import ConfirmModal from '../components/ConfirmModal.svelte';

  let { params = { slug: '' } } = $props<{ params?: { slug?: string } }>();

  let dashboard = $state<Dashboard | null>(null);
  let groups = $state<Group[]>([]);
  let loading = $state(true);
  let error = $state('');
  let notFound = $state(false);
  let dashList = $state<DashListItem[]>([]);

  let editingGroup = $state<Group | null>(null);
  let editingItem = $state<Item | null>(null);

  let groupOpen = $state(false);
  let itemOpen = $state(false);
  let dashOpen = $state(false);
  let groupForm = $state<GroupForm>(emptyGroupForm());
  let itemForm = $state<ItemForm>(emptyItemForm());
  let dashForm = $state<DashboardForm>(emptyDashboardForm());

  let confirmOpen = $state(false);
  let confirmMsg = $state('');
  let confirmAction = $state<(() => Promise<void>) | null>(null);

  function askConfirm(message: string, action: () => Promise<void>) {
    confirmMsg = message;
    confirmAction = action;
    confirmOpen = true;
  }

  async function load() {
    loading = true;
    error = '';
    notFound = false;
    try {
      if ($user) {
        dashList = toDashList(await api.listDashboards());
      } else {
        dashList = [];
      }

      const slug = params?.slug;
      let d: Dashboard | null = null;

      notFound = false;
      if (slug) {
        try {
          d = await api.getDashboard(slug);
        } catch (e: unknown) {
          const msg = e instanceof Error ? e.message : '';
          if (/login|unauthorized|access denied/i.test(msg)) {
            replace('/login');
            return;
          }
          notFound = true;
          dashboard = null;
          groups = [];
          return;
        }
      } else {
        try {
          d = await api.getMain();
        } catch {
          d = null;
        }
        if (!d && $user && dashList.length > 0) {
          replace('/' + dashList[0].slug);
          return;
        }
        if (!d && !$user) {
          replace('/login');
          return;
        }
      }

      if (!d) {
        if (!$user) {
          replace('/login');
          return;
        }
        error = 'No dashboards yet.';
        dashboard = null;
        groups = [];
        currentDashboard.set(null);
      } else {
        dashboard = d;
        groups = (d.groups ?? []).slice().sort((a, b) => a.position - b.position);
        currentDashboard.set(d);
      }
    } catch (e: unknown) {
      const msg = e instanceof Error ? e.message : 'Failed to load';
      if (/login|unauthorized|access denied/i.test(msg)) {
        replace('/login');
        return;
      }
      error = msg;
      dashboard = null;
      groups = [];
    } finally {
      loading = false;
    }
  }

  onMount(load);
  $effect(() => {
    void params?.slug;
    void $user;
    load();
  });

  /* —— group —— */
  function openNewGroup() {
    editingGroup = null;
    groupForm = emptyGroupForm();
    groupOpen = true;
  }
  function openEditGroup(g: Group) {
    editingGroup = g;
    groupForm = groupToForm(g);
    groupOpen = true;
  }
  async function saveGroup() {
    if (!dashboard) return;
    const payload = {
      title: groupForm.title.trim(),
      description: groupForm.description,
      icon: groupForm.icon,
      iconDark: groupForm.iconDark,
      itemSize: groupForm.itemSize,
    };
    if (editingGroup) await api.updateGroup(editingGroup.id, payload);
    else await api.createGroup(dashboard.id, { ...payload, position: groups.length });
    groupOpen = false;
    await load();
  }

  /* —— item —— */
  function openNewItem(g: Group) {
    editingItem = null;
    itemForm = emptyItemForm(g.id);
    itemOpen = true;
  }
  function openEditItem(item: Item) {
    editingItem = item;
    itemForm = itemToForm(item);
    itemOpen = true;
  }
  async function saveItem() {
    const payload = {
      title: itemForm.title.trim(),
      description: itemForm.description,
      url: itemForm.url.trim(),
      icon: itemForm.icon || 'mdi:link',
      iconDark: itemForm.iconDark,
    };
    if (editingItem) await api.updateItem(editingItem.id, payload);
    else {
      const pos = groups.find((g) => g.id === itemForm.groupId)?.items?.length ?? 0;
      await api.createItem(itemForm.groupId, { ...payload, position: pos });
    }
    itemOpen = false;
    await load();
  }

  /* —— dashboard —— */
  function openDashSettings() {
    dashForm = dashboard ? dashboardToForm(dashboard) : emptyDashboardForm(false);
    dashOpen = true;
  }
  function openCreateDashboard() {
    dashForm = emptyDashboardForm(true);
    dashOpen = true;
  }
  async function saveDash() {
    if (!dashForm.name.trim() || !dashForm.slug.trim()) return;
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
    };
    if (dashForm.creating || !dashboard) {
      if (!$user) return;
      const d = await api.createDashboard(body);
      dashOpen = false;
      push('/' + d.slug);
      return;
    }
    const prevSlug = dashboard.slug;
    await api.updateDashboard(dashboard.id, body);
    dashOpen = false;
    if (dashForm.slug !== prevSlug) push('/' + dashForm.slug);
    else await load();
  }

  async function deleteDashboard() {
    if (!dashboard) return;
    const id = dashboard.id;
    await api.deleteDashboard(id);
    editMode.set(false);
    const rest = dashList.filter((d) => d.id !== id);
    dashList = rest;
    if (rest.length > 0) push('/' + rest[0].slug);
    else {
      dashboard = null;
      groups = [];
      replace('/');
    }
  }

  async function cloneDashboard() {
    if (!dashboard) return;
    const d = await api.cloneDashboard(dashboard.id);
    push('/' + d.slug);
  }

  async function exportDashboard() {
    if (!dashboard) return;
    const blob = await api.exportDashboard(dashboard.id);
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${dashboard.slug}.dashlit.json`;
    a.click();
    URL.revokeObjectURL(url);
  }

  function importDashboard() {
    const input = document.createElement('input');
    input.type = 'file';
    input.accept = 'application/json,.json';
    input.onchange = async () => {
      const file = input.files?.[0];
      if (!file) return;
      try {
        const text = await file.text();
        const data = JSON.parse(text);
        const d = await api.importDashboard(data);
        push('/' + d.slug);
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : 'Import failed';
        alert(msg);
      }
    };
    input.click();
  }

  async function cloneGroup(g: Group) {
    const created = await api.cloneGroup(g.id);
    groups = [...groups, created];
    await tick();

    const element = document.querySelector<HTMLElement>(
      `[data-dashboard-group="${CSS.escape(created.id)}"]`,
    );
    if (!element) return;

    const bounds = element.getBoundingClientRect();
    if (bounds.top < 0 || bounds.bottom > window.innerHeight) {
      element.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
  }

  async function cloneItem(item: Item) {
    const created = await api.cloneItem(item.id);
    groups = groups.map((x) =>
      x.id === created.groupId ? { ...x, items: [...(x.items ?? []), created] } : x,
    );
  }

  async function persistLayout() {
    if (!dashboard) return;
    try {
      await api.updateLayout(dashboard.id, buildLayoutPayload(groups));
    } catch {
      /* ignore transient DnD errors */
    }
  }
</script>

{#if loading}
  <div class="flex min-h-dvh items-center justify-center text-sm text-text-subtle">Loading…</div>
{:else if notFound}
  <div class="flex min-h-dvh flex-col items-center justify-center gap-3 px-4 text-center">
    <p class="text-5xl font-semibold tracking-tight text-text">404</p>
    <p class="text-sm text-text-muted">Page not found</p>
    <a href="#/" class="mt-2 text-sm text-primary hover:underline">Go to home</a>
  </div>
{:else if dashboard}
  {@const d = dashboard}
  {#if d.cleanMode}
    <div class={pageContainerClass(d.width === 'wide')}>
      <CleanHeader dashboard={d} dashboards={dashList} />
      <DashboardBoard
        dashboard={d}
        bind:groups
        onEditGroup={openEditGroup}
        onDeleteGroup={(g) =>
          askConfirm(`Delete group “${g.title}” and all its items?`, async () => {
            await api.deleteGroup(g.id);
            await load();
          })}
        onCloneGroup={cloneGroup}
        onAddItem={openNewItem}
        onEditItem={openEditItem}
        onDeleteItem={(it) =>
          askConfirm(`Delete “${it.title}”?`, async () => {
            await api.deleteItem(it.id);
            await load();
          })}
        onCloneItem={cloneItem}
        onLayoutChange={persistLayout}
      />
      {#if $user}
        <EditFabs
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
    <AppLayout dashboards={dashList} currentSlug={d.slug} wide={d.width === 'wide'}>
      <DashboardBoard
        dashboard={d}
        bind:groups
        onEditGroup={openEditGroup}
        onDeleteGroup={(g) =>
          askConfirm(`Delete group “${g.title}” and all its items?`, async () => {
            await api.deleteGroup(g.id);
            await load();
          })}
        onCloneGroup={cloneGroup}
        onAddItem={openNewItem}
        onEditItem={openEditItem}
        onDeleteItem={(it) =>
          askConfirm(`Delete “${it.title}”?`, async () => {
            await api.deleteItem(it.id);
            await load();
          })}
        onCloneItem={cloneItem}
        onLayoutChange={persistLayout}
      />
      {#if $user}
        <EditFabs
          raised
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
        <button type="button" class="rounded-btn bg-primary px-4 py-2 text-sm text-white" onclick={openCreateDashboard}> Create dashboard </button>
      {/if}
    </div>
  </AppLayout>
{/if}

<GroupFormModal bind:open={groupOpen} bind:form={groupForm} editing={!!editingGroup} onSave={saveGroup} />
<ItemFormModal bind:open={itemOpen} bind:form={itemForm} editing={!!editingItem} onSave={saveItem} />
<DashboardFormModal bind:open={dashOpen} bind:form={dashForm} onSave={saveDash} />
<ConfirmModal
  bind:open={confirmOpen}
  title="Confirm"
  message={confirmMsg}
  onConfirm={async () => {
    await confirmAction?.();
  }}
/>
