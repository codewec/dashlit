<script lang="ts">
  import { onMount } from 'svelte';
  import { DragDropProvider, DragOverlay } from '@dnd-kit-svelte/svelte';
  import { move } from '@dnd-kit/helpers';
  import { push, replace } from 'svelte-spa-router';
  import { api, type Dashboard, type Group, type Item, type ItemSize, type Layout, type Width } from '../lib/api';
  import { user, editMode, searchQuery, currentDashboard } from '../lib/stores';
  import AppLayout from '../layouts/AppLayout.svelte';
  import GroupCard from '../components/GroupCard.svelte';
  import ItemCard from '../components/ItemCard.svelte';
  import Modal from '../components/Modal.svelte';
  import ConfirmModal from '../components/ConfirmModal.svelte';
  import IconPicker from '../components/IconPicker.svelte';

  let { params = { slug: '' } } = $props<{ params?: { slug?: string } }>();

  let dashboard = $state<Dashboard | null>(null);
  let groups = $state<Group[]>([]);
  let loading = $state(true);
  let error = $state('');
  let dashList = $state<{ id: string; name: string; slug: string }[]>([]);

  let groupOpen = $state(false);
  let itemOpen = $state(false);
  let dashOpen = $state(false);
  let confirmOpen = $state(false);
  let confirmMsg = $state('');
  let confirmAction = $state<(() => Promise<void>) | null>(null);

  let editingGroup = $state<Group | null>(null);
  let editingItem = $state<Item | null>(null);
  let targetGroupId = $state('');

  let gTitle = $state('');
  let gDesc = $state('');
  let gIcon = $state('');
  let gItemSize = $state<ItemSize>('1x1');
  let gTitleErr = $state(false);

  let iTitle = $state('');
  let iDesc = $state('');
  let iUrl = $state('');
  let iIcon = $state('mdi:link');
  let iTitleErr = $state(false);
  let iUrlErr = $state(false);

  let dName = $state('');
  let dSlug = $state('');
  let dPrivacy = $state<'public' | 'private' | 'users'>('private');
  let dLayout = $state<Layout>('rows');
  let dWidth = $state<Width>('default');
  let creatingNew = $state(false);

  const filteredGroups = $derived.by(() => {
    const q = $searchQuery.trim().toLowerCase();
    if (!q) return groups;
    return groups
      .map((g) => ({
        ...g,
        items: (g.items || []).filter((it) => it.title.toLowerCase().includes(q) || (it.description || '').toLowerCase().includes(q)),
      }))
      .filter((g) => (g.items?.length || 0) > 0 || g.title.toLowerCase().includes(q));
  });

  const itemsByGroup = $derived.by(() => {
    const map: Record<string, Item[]> = {};
    for (const g of filteredGroups) {
      map[g.id] = [...(g.items || [])].sort((a, b) => a.position - b.position);
    }
    return map;
  });

  async function load() {
    loading = true;
    error = '';
    try {
      if ($user) {
        const list = await api.listDashboards();
        dashList = list.map((x) => ({ id: x.id, name: x.name, slug: x.slug }));
      } else {
        dashList = [];
      }

      const slug = params?.slug;
      let d: Dashboard | null = null;

      if (slug) {
        d = await api.getDashboard(slug);
      } else {
        // root: prefer main, else first available
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
      } else {
        dashboard = d;
        groups = (d.groups || []).slice().sort((a, b) => a.position - b.position);
        currentDashboard.set(d);
      }
    } catch (e: any) {
      if (/login|unauthorized|access denied/i.test(e.message || '')) {
        replace('/login');
        return;
      }
      error = e.message || 'Failed to load';
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

  function askConfirm(message: string, action: () => Promise<void>) {
    confirmMsg = message;
    confirmAction = action;
    confirmOpen = true;
  }

  function openNewGroup() {
    editingGroup = null;
    gTitle = '';
    gDesc = '';
    gIcon = '';
    gItemSize = '1x1';
    gTitleErr = false;
    groupOpen = true;
  }
  function openEditGroup(g: Group) {
    editingGroup = g;
    gTitle = g.title;
    gDesc = g.description || '';
    gIcon = g.icon || '';
    gItemSize = g.itemSize || '1x1';
    gTitleErr = false;
    groupOpen = true;
  }
  async function saveGroup() {
    gTitleErr = !gTitle.trim();
    if (gTitleErr || !dashboard) return;
    const payload = {
      title: gTitle.trim(),
      description: gDesc,
      icon: gIcon,
      itemSize: gItemSize,
    };
    if (editingGroup) await api.updateGroup(editingGroup.id, payload);
    else await api.createGroup(dashboard.id, { ...payload, position: groups.length });
    groupOpen = false;
    await load();
  }

  function openNewItem(g: Group) {
    editingItem = null;
    targetGroupId = g.id;
    iTitle = '';
    iDesc = '';
    iUrl = '';
    iIcon = 'mdi:link';
    iTitleErr = false;
    iUrlErr = false;
    itemOpen = true;
  }
  function openEditItem(item: Item) {
    editingItem = item;
    targetGroupId = item.groupId;
    iTitle = item.title;
    iDesc = item.description || '';
    iUrl = item.url;
    iIcon = item.icon;
    iTitleErr = false;
    iUrlErr = false;
    itemOpen = true;
  }
  async function saveItem() {
    iTitleErr = !iTitle.trim();
    iUrlErr = !iUrl.trim();
    if (iTitleErr || iUrlErr) return;
    const payload = {
      title: iTitle.trim(),
      description: iDesc,
      url: iUrl.trim(),
      icon: iIcon || 'mdi:link',
    };
    if (editingItem) await api.updateItem(editingItem.id, payload);
    else await api.createItem(targetGroupId, { ...payload, position: (itemsByGroup[targetGroupId] || []).length });
    itemOpen = false;
    await load();
  }

  function openDashSettings() {
    creatingNew = false;
    if (dashboard) {
      dName = dashboard.name;
      dSlug = dashboard.slug;
      dPrivacy = dashboard.privacy;
      dLayout = dashboard.layout;
      dWidth = dashboard.width || 'default';
    } else {
      dName = 'Home';
      dSlug = 'home';
      dPrivacy = 'private';
      dLayout = 'rows';
      dWidth = 'default';
    }
    dashOpen = true;
  }
  function openCreateDashboard() {
    dName = '';
    dSlug = '';
    dPrivacy = 'private';
    dLayout = 'rows';
    dWidth = 'default';
    // force create path in saveDash
    creatingNew = true;
    dashOpen = true;
  }
  async function saveDash() {
    if (!dName.trim() || !dSlug.trim()) return;
    if (creatingNew || !dashboard) {
      if (!$user) return;
      const d = await api.createDashboard({
        name: dName,
        slug: dSlug,
        privacy: dPrivacy,
        layout: dLayout,
        width: dWidth,
      } as any);
      creatingNew = false;
      dashOpen = false;
      push('/' + d.slug);
      return;
    }
    await api.updateDashboard(dashboard.id, {
      name: dName,
      slug: dSlug,
      privacy: dPrivacy,
      layout: dLayout,
      width: dWidth,
    } as any);
    dashOpen = false;
    if (dSlug !== dashboard.slug) push('/' + dSlug);
    else await load();
  }

  async function persistLayout() {
    if (!dashboard) return;
    const sortedGroups = groups.map((g, i) => ({ id: g.id, position: i }));
    const items: { id: string; groupId: string; position: number }[] = [];
    for (const g of groups) {
      (g.items || []).forEach((it, i) => items.push({ id: it.id, groupId: g.id, position: i }));
    }
    await api.updateLayout(dashboard.id, { groups: sortedGroups, items });
  }

  function onDragOver(event: any) {
    if (!$editMode) return;
    const { source } = event.operation;
    if (source?.type === 'column') {
      const from = groups.findIndex((g) => g.id === source.id);
      const targetId = event.operation.target?.id;
      if (from < 0 || !targetId) return;
      const to = groups.findIndex((g) => g.id === targetId);
      if (to < 0 || from === to) return;
      const next = [...groups];
      const [moved] = next.splice(from, 1);
      next.splice(to, 0, moved);
      groups = next;
      return;
    }
    const bag: Record<string, Item[]> = {};
    for (const g of groups) bag[g.id] = [...(g.items || [])];
    const nextBag = move(bag, event) as Record<string, Item[]>;
    groups = groups.map((g) => ({
      ...g,
      items: (nextBag[g.id] || []).map((it, i) => ({ ...it, groupId: g.id, position: i })),
    }));
  }

  async function onDragEnd() {
    if (!$editMode || !dashboard) return;
    try {
      await persistLayout();
    } catch {}
  }
</script>

{#if loading}
  <div class="flex min-h-dvh items-center justify-center text-sm text-[var(--color-text-subtle)]">Loading…</div>
{:else if dashboard}
  <AppLayout
    dashboards={dashList}
    currentSlug={dashboard.slug}
    wide={dashboard.width === 'wide'}
    onCreateDashboard={openCreateDashboard}
    onDeleteDashboard={() =>
      askConfirm(`Delete dashboard “${dashboard?.name}”?`, async () => {
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
      })}
  >
    <DragDropProvider {onDragOver} {onDragEnd}>
      <div
        class={dashboard.layout === 'columns'
          ? dashboard.width === 'wide'
            ? 'grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4'
            : 'grid grid-cols-1 gap-4 md:grid-cols-2'
          : dashboard.layout === 'masonry'
            ? dashboard.width === 'wide'
              ? 'columns-1 gap-4 space-y-4 sm:columns-2 lg:columns-3 xl:columns-4 [&>*]:mb-4 [&>*]:break-inside-avoid'
              : 'columns-1 gap-4 space-y-4 md:columns-2 [&>*]:mb-4 [&>*]:break-inside-avoid'
            : 'flex flex-col gap-4'}
      >
        {#each filteredGroups as group, gIndex (group.id)}
          <div class={dashboard.layout === 'columns' ? 'min-w-0' : dashboard.layout === 'masonry' ? 'break-inside-avoid' : ''}>
            <GroupCard
              {group}
              index={gIndex}
              layout={dashboard.layout}
              wide={dashboard.width === 'wide'}
              onEdit={openEditGroup}
              onDelete={(g) =>
                askConfirm(`Delete group “${g.title}” and all its items?`, async () => {
                  await api.deleteGroup(g.id);
                  await load();
                })}
              onAddItem={openNewItem}
            >
              {#each itemsByGroup[group.id] || [] as item, iIndex (item.id)}
                <ItemCard
                  {item}
                  index={iIndex}
                  groupId={group.id}
                  itemSize={group.itemSize}
                  onEdit={openEditItem}
                  onDelete={(it) =>
                    askConfirm(`Delete “${it.title}”?`, async () => {
                      await api.deleteItem(it.id);
                      await load();
                    })}
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
            <GroupCard group={source.data.group} index={0} layout={dashboard?.layout} wide={dashboard?.width === 'wide'} isOverlay />
          {/if}
        {/snippet}
      </DragOverlay>
    </DragDropProvider>

    {#if $editMode}
      <div class="fixed bottom-5 left-1/2 z-30 flex -translate-x-1/2 gap-2 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-3 py-2 shadow-xl">
        <button type="button" class="rounded-full bg-[var(--color-primary)] px-4 py-1.5 text-xs font-medium text-white" onclick={openNewGroup}>+ Group</button>
        <button type="button" class="rounded-full px-4 py-1.5 text-xs text-[var(--color-text)] hover:bg-[var(--color-surface-2)]" onclick={openDashSettings}>Settings</button>
      </div>
    {/if}
  </AppLayout>
{:else}
  <AppLayout dashboards={dashList} currentSlug="" onCreateDashboard={openCreateDashboard}>
    <div class="flex flex-col items-center gap-3 py-20">
      <p class="text-sm text-[var(--color-text-muted)]">{error || 'No dashboards yet.'}</p>
      {#if $user}
        <button type="button" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-4 py-2 text-sm text-white" onclick={openCreateDashboard}> Create dashboard </button>
      {/if}
    </div>
  </AppLayout>
{/if}

<Modal bind:open={groupOpen} title={editingGroup ? 'Edit group' : 'New group'}>
  <form
    class="space-y-3"
    onsubmit={(e) => {
      e.preventDefault();
      saveGroup();
    }}
  >
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Title</span>
      <input
        class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] {gTitleErr
          ? 'field-error'
          : ''}"
        bind:value={gTitle}
      />
      {#if gTitleErr}<span class="mt-1 text-xs text-[var(--color-danger)]">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Description</span>
      <input
        class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]"
        bind:value={gDesc}
      />
    </label>
    <div>
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Icon</span>
      <IconPicker bind:value={gIcon} />
    </div>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Item size in this group</span>
      <select class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={gItemSize}>
        <option value="1x1">1×1 — icon only</option>
        <option value="1x2">1×2 — icon, title, description</option>
      </select>
    </label>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-[var(--radius-btn)] px-3 py-2 text-sm text-[var(--color-text-muted)]" onclick={() => (groupOpen = false)}>Cancel</button>
      <button type="submit" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>

<Modal bind:open={itemOpen} title={editingItem ? 'Edit item' : 'New item'}>
  <form
    class="space-y-3"
    onsubmit={(e) => {
      e.preventDefault();
      saveItem();
    }}
  >
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Title</span>
      <input
        class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] {iTitleErr
          ? 'field-error'
          : ''}"
        bind:value={iTitle}
      />
      {#if iTitleErr}<span class="mt-1 text-xs text-[var(--color-danger)]">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">URL</span>
      <input
        class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] {iUrlErr
          ? 'field-error'
          : ''}"
        bind:value={iUrl}
      />
      {#if iUrlErr}<span class="mt-1 text-xs text-[var(--color-danger)]">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Description</span>
      <input
        class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]"
        bind:value={iDesc}
      />
    </label>
    <div>
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Icon</span>
      <IconPicker bind:value={iIcon} />
    </div>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-[var(--radius-btn)] px-3 py-2 text-sm text-[var(--color-text-muted)]" onclick={() => (itemOpen = false)}>Cancel</button>
      <button type="submit" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>

<Modal bind:open={dashOpen} title={creatingNew || !dashboard ? 'New dashboard' : 'Dashboard settings'}>
  <form
    class="space-y-3"
    onsubmit={(e) => {
      e.preventDefault();
      saveDash();
    }}
  >
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Name</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={dName} required />
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Slug</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={dSlug} required />
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Privacy</span>
      <select class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={dPrivacy}>
        <option value="public">Public</option>
        <option value="users">Authenticated users</option>
        <option value="private">Private</option>
      </select>
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Layout</span>
      <select class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={dLayout}>
        <option value="rows">Rows</option>
        <option value="columns">Columns</option>
        <option value="masonry">Masonry</option>
      </select>
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Width</span>
      <select class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={dWidth}>
        <option value="default">Default</option>
        <option value="wide">Wide</option>
      </select>
    </label>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-[var(--radius-btn)] px-3 py-2 text-sm text-[var(--color-text-muted)]" onclick={() => (dashOpen = false)}>Cancel</button>
      <button type="submit" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>

<ConfirmModal
  bind:open={confirmOpen}
  title="Confirm"
  message={confirmMsg}
  onConfirm={async () => {
    await confirmAction?.();
  }}
/>
