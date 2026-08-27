<script lang="ts">
  import { onMount } from 'svelte';
  import { DragDropProvider, DragOverlay } from '@dnd-kit-svelte/svelte';
  import { move } from '@dnd-kit/helpers';
  import { push, replace } from 'svelte-spa-router';
  import { api, setToken, type Dashboard, type Group, type Item, type ItemSize, type Layout, type Width } from '../lib/api';
  import { user, editMode, searchQuery, currentDashboard, theme, applyTheme } from '../lib/stores';
  import AppLayout from '../layouts/AppLayout.svelte';
  import GroupCard from '../components/GroupCard.svelte';
  import ItemCard from '../components/ItemCard.svelte';
  import Modal from '../components/Modal.svelte';
  import ConfirmModal from '../components/ConfirmModal.svelte';
  import IconField from '../components/IconField.svelte';
  import Icon from '../components/Icon.svelte';
  import NavMenu from '../components/NavMenu.svelte';
  import { Switch } from 'bits-ui';

  let { params = { slug: '' } } = $props<{ params?: { slug?: string } }>();

  let dashboard = $state<Dashboard | null>(null);
  let groups = $state<Group[]>([]);
  let loading = $state(true);
  let error = $state('');
  let dashList = $state<{ id: string; name: string; slug: string; icon?: string; iconDark?: string }[]>([]);

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
  let gIconDark = $state('');
  let gItemSize = $state<ItemSize>('1x1');
  let gTitleErr = $state(false);

  let iTitle = $state('');
  let iDesc = $state('');
  let iUrl = $state('');
  let iIcon = $state('mdi:link');
  let iIconDark = $state('');
  let iTitleErr = $state(false);
  let iUrlErr = $state(false);

  let dName = $state('');
  let dDesc = $state('');
  let dIcon = $state('');
  let dIconDark = $state('');
  let dClean = $state(false);
  let dCleanStr = $state('off');
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
        dashList = list.map((x) => ({ id: x.id, name: x.name, slug: x.slug, icon: x.icon, iconDark: x.iconDark }));
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
    gIconDark = '';
    gItemSize = '1x1';
    gTitleErr = false;
    groupOpen = true;
  }
  function openEditGroup(g: Group) {
    editingGroup = g;
    gTitle = g.title;
    gDesc = g.description || '';
    gIcon = g.icon || '';
    gIconDark = g.iconDark || '';
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
      iconDark: gIconDark,
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
    iIconDark = '';
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
    iIconDark = item.iconDark || '';
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
      iconDark: iIconDark,
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
      dDesc = dashboard.description || '';
      dIcon = dashboard.icon || '';
      dIconDark = dashboard.iconDark || '';
      dPrivacy = dashboard.privacy;
      dLayout = dashboard.layout;
      dWidth = dashboard.width || 'default';
      dClean = !!dashboard.cleanMode;
      dCleanStr = dashboard.cleanMode ? 'on' : 'off';
    } else {
      dName = 'Home';
      dSlug = 'home';
      dDesc = '';
      dIcon = '';
      dIconDark = '';
      dPrivacy = 'private';
      dLayout = 'rows';
      dWidth = 'default';
      dClean = false;
      dCleanStr = 'off';
    }
    dashOpen = true;
  }
  function openCreateDashboard() {
    dName = '';
    dSlug = '';
    dDesc = '';
    dIcon = '';
    dIconDark = '';
    dPrivacy = 'private';
    dLayout = 'rows';
    dWidth = 'default';
    dClean = false;
    dCleanStr = 'off';
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
        description: dDesc,
        icon: dIcon,
        iconDark: dIconDark,
        privacy: dPrivacy,
        layout: dLayout,
        width: dWidth,
        cleanMode: dCleanStr === 'on',
      } as any);
      creatingNew = false;
      dashOpen = false;
      push('/' + d.slug);
      return;
    }
    await api.updateDashboard(dashboard.id, {
      name: dName,
      slug: dSlug,
      description: dDesc,
      icon: dIcon,
      iconDark: dIconDark,
      privacy: dPrivacy,
      layout: dLayout,
      width: dWidth,
      cleanMode: dCleanStr === 'on',
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
  {#snippet boardBody()}
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
  {/snippet}

  {#snippet editFabs()}
    {#if $user}
      {#if $editMode}
        <div class="pointer-events-none fixed bottom-5 left-4 right-4 z-30 flex items-center justify-between gap-3">
          <!-- create dashboard -->
          <button
            type="button"
            class="pointer-events-auto flex h-11 items-center gap-1.5 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-4 text-xs font-medium text-[var(--color-text)] shadow-lg hover:bg-[var(--color-surface-2)]"
            onclick={openCreateDashboard}
            title="New dashboard"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14" /></svg>
            <span class="hidden sm:inline">Dashboard</span>
          </button>

          <!-- center: group / settings / done -->
          <div class="pointer-events-auto flex items-center gap-1 rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] p-1.5 shadow-lg">
            <button type="button" class="rounded-full px-3 py-2 text-xs font-medium text-[var(--color-text)] hover:bg-[var(--color-surface-2)]" onclick={openNewGroup}> + Group </button>
            <button type="button" class="rounded-full px-3 py-2 text-xs text-[var(--color-text-muted)] hover:bg-[var(--color-surface-2)] hover:text-[var(--color-text)]" onclick={openDashSettings}>
              Settings
            </button>
            <button type="button" class="rounded-full bg-[var(--color-primary)] px-3 py-2 text-xs font-medium text-white hover:bg-[var(--color-primary-hover)]" onclick={() => editMode.set(false)}>
              Save
            </button>
          </div>

          <!-- delete dashboard -->
          <button
            type="button"
            class="pointer-events-auto flex h-11 items-center gap-1.5 rounded-full border border-[var(--color-danger)]/40 bg-[var(--color-surface)] px-4 text-xs font-medium text-[var(--color-danger)] shadow-lg hover:bg-[var(--color-danger-soft)]"
            onclick={() =>
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
            title="Delete dashboard"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M3 6h18" /><path d="M8 6V4h8v2" /><path d="M19 6l-1 14H6L5 6" /></svg>
            <span class="hidden sm:inline">Delete</span>
          </button>
        </div>
      {:else}
        <div class="fixed bottom-5 left-1/2 z-30 -translate-x-1/2">
          <button
            type="button"
            class="flex h-11 items-center rounded-full border border-[var(--color-border)] bg-[var(--color-surface)] px-5 text-xs font-medium text-[var(--color-text)] shadow-lg hover:bg-[var(--color-surface-2)]"
            onclick={() => editMode.set(true)}
          >
            Edit
          </button>
        </div>
      {/if}
    {/if}
  {/snippet}

  {#if dashboard.cleanMode}
    <div class={dashboard.width === 'wide' ? 'mx-auto max-w-none px-4 py-6' : 'mx-auto max-w-6xl px-4 py-6'}>
      <div class="mb-6 flex items-start gap-3">
        {#if dashboard.icon}
          <Icon icon={dashboard.icon} iconDark={dashboard.iconDark} size={40} class="mt-0.5 shrink-0 rounded-xl" />
        {/if}
        <div class="min-w-0 flex-1">
          <h1 class="text-xl font-semibold tracking-tight text-[var(--color-text)]">{dashboard.name}</h1>
          {#if dashboard.description}
            <p class="mt-1 text-sm text-[var(--color-text-muted)]">{dashboard.description}</p>
          {/if}
        </div>
        <div class="shrink-0">
          <NavMenu dashboards={dashList} currentSlug={dashboard.slug} />
        </div>
      </div>
      {@render boardBody()}
      {@render editFabs()}
    </div>
  {:else}
    <AppLayout dashboards={dashList} currentSlug={dashboard.slug} wide={dashboard.width === 'wide'}>
      {@render boardBody()}
      {@render editFabs()}
    </AppLayout>
  {/if}
{:else}
  <AppLayout dashboards={dashList} currentSlug="">
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
      <IconField bind:value={gIcon} bind:valueDark={gIconDark} />
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
      <IconField bind:value={iIcon} bind:valueDark={iIconDark} defaultIcon="mdi:link" />
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
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Description</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={dDesc} />
    </label>
    <div>
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Icon</span>
      <IconField bind:value={dIcon} bind:valueDark={dIconDark} />
    </div>
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
      <div class="space-y-3">
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
      </div>
      <div class="flex flex-col justify-center gap-4">
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm text-[var(--color-text)]">Clean mode</span>
          <Switch.Root
            checked={dCleanStr === 'on'}
            onCheckedChange={(v) => (dCleanStr = v ? 'on' : 'off')}
            class="peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border border-transparent bg-[var(--color-border)] transition data-[state=checked]:bg-[var(--color-primary)]"
          >
            <Switch.Thumb class="pointer-events-none block h-4 w-4 rounded-full bg-white shadow transition-transform translate-x-0.5 data-[state=checked]:translate-x-[1.1rem]" />
          </Switch.Root>
        </div>
        <div class="flex items-center justify-between gap-3">
          <span class="text-sm text-[var(--color-text)]">Wide mode</span>
          <Switch.Root
            checked={dWidth === 'wide'}
            onCheckedChange={(v) => (dWidth = v ? 'wide' : 'default')}
            class="peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border border-transparent bg-[var(--color-border)] transition data-[state=checked]:bg-[var(--color-primary)]"
          >
            <Switch.Thumb class="pointer-events-none block h-4 w-4 rounded-full bg-white shadow transition-transform translate-x-0.5 data-[state=checked]:translate-x-[1.1rem]" />
          </Switch.Root>
        </div>
      </div>
    </div>
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
