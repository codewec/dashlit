<script lang="ts">
  import { onMount } from 'svelte';
  import { DragDropProvider, DragOverlay } from '@dnd-kit-svelte/svelte';
  import { move } from '@dnd-kit/helpers';
  import { push } from 'svelte-spa-router';
  import { api, type Dashboard, type Group, type Item } from '../lib/api';
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
  let needLogin = $state(false);
  let dashList = $state<{ id: string; name: string; slug: string }[]>([]);

  // modals
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
  let gTitleErr = $state(false);
  let iTitle = $state('');
  let iDesc = $state('');
  let iUrl = $state('');
  let iIcon = $state('mdi:link');
  let iSize = $state<'1x1' | '1x2'>('1x1');
  let iTitleErr = $state(false);
  let iUrlErr = $state(false);

  let dName = $state('');
  let dSlug = $state('');
  let dPrivacy = $state<'public' | 'private' | 'users'>('private');
  let dLayout = $state<'rows' | 'columns'>('rows');

  const filteredGroups = $derived.by(() => {
    const q = $searchQuery.trim().toLowerCase();
    if (!q) return groups;
    return groups
      .map((g) => ({
        ...g,
        items: (g.items || []).filter(
          (it) => it.title.toLowerCase().includes(q) || (it.description || '').toLowerCase().includes(q)
        ),
      }))
      .filter((g) => (g.items?.length || 0) > 0 || g.title.toLowerCase().includes(q));
  });

  // DnD structure: Record<groupId, items[]>
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
    needLogin = false;
    try {
      const slug = params?.slug;
      const d = slug ? await api.getDashboard(slug) : await api.getMain();
      if (!d) {
        error = 'No dashboard yet.';
        dashboard = null;
        groups = [];
      } else {
        dashboard = d;
        groups = (d.groups || []).slice().sort((a, b) => a.position - b.position);
        currentDashboard.set(d);
      }
      if ($user) {
        const list = await api.listDashboards();
        dashList = list.map((x) => ({ id: x.id, name: x.name, slug: x.slug }));
      }
    } catch (e: any) {
      if (/login|unauthorized/i.test(e.message || '')) needLogin = true;
      else error = e.message || 'Failed to load';
      dashboard = null;
      groups = [];
    } finally {
      loading = false;
    }
  }

  onMount(load);
  $effect(() => {
    void params?.slug;
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
    gTitleErr = false;
    groupOpen = true;
  }
  function openEditGroup(g: Group) {
    editingGroup = g;
    gTitle = g.title;
    gTitleErr = false;
    groupOpen = true;
  }
  async function saveGroup() {
    gTitleErr = !gTitle.trim();
    if (gTitleErr || !dashboard) return;
    if (editingGroup) await api.updateGroup(editingGroup.id, { title: gTitle.trim() });
    else await api.createGroup(dashboard.id, { title: gTitle.trim(), position: groups.length });
    groupOpen = false;
    await load();
  }

  function openNewItem(g: Group) {
    editingItem = null;
    targetGroupId = g.id;
    iTitle = ''; iDesc = ''; iUrl = ''; iIcon = 'mdi:link'; iSize = '1x1';
    iTitleErr = false; iUrlErr = false;
    itemOpen = true;
  }
  function openEditItem(item: Item) {
    editingItem = item;
    targetGroupId = item.groupId;
    iTitle = item.title; iDesc = item.description || ''; iUrl = item.url;
    iIcon = item.icon; iSize = item.size;
    iTitleErr = false; iUrlErr = false;
    itemOpen = true;
  }
  async function saveItem() {
    iTitleErr = !iTitle.trim();
    iUrlErr = !iUrl.trim();
    if (iTitleErr || iUrlErr) return;
    if (editingItem) {
      await api.updateItem(editingItem.id, {
        title: iTitle.trim(), description: iDesc, url: iUrl.trim(), icon: iIcon || 'mdi:link', size: iSize,
      });
    } else {
      await api.createItem(targetGroupId, {
        title: iTitle.trim(), description: iDesc, url: iUrl.trim(), icon: iIcon || 'mdi:link', size: iSize,
        position: (itemsByGroup[targetGroupId] || []).length,
      });
    }
    itemOpen = false;
    await load();
  }

  function openDashSettings() {
    if (dashboard) {
      dName = dashboard.name; dSlug = dashboard.slug;
      dPrivacy = dashboard.privacy; dLayout = dashboard.layout;
    } else {
      dName = 'Home'; dSlug = 'home'; dPrivacy = 'private'; dLayout = 'rows';
    }
    dashOpen = true;
  }
  async function saveDash() {
    if (!dName.trim() || !dSlug.trim()) return;
    if (!dashboard && $user) {
      const d = await api.createDashboard({ name: dName, slug: dSlug, privacy: dPrivacy, layout: dLayout });
      dashOpen = false;
      push('/' + d.slug);
      return;
    }
    if (!dashboard) return;
    await api.updateDashboard(dashboard.id, { name: dName, slug: dSlug, privacy: dPrivacy, layout: dLayout } as any);
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
      // reorder groups
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
    // items via move helper — adapt structure
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
    try { await persistLayout(); } catch {}
  }
</script>

{#if needLogin}
  <!-- bare message — login uses its own route/layout -->
  <div class="flex min-h-dvh flex-col items-center justify-center gap-4 p-6">
    <p class="text-[var(--color-text-muted)]">This dashboard requires authentication.</p>
    <a href="#/login" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-4 py-2 text-sm font-medium text-white">Sign in</a>
  </div>
{:else}
  <AppLayout dashboards={dashList}>
    {#if loading}
      <p class="py-20 text-center text-sm text-[var(--color-text-subtle)]">Loading…</p>
    {:else if error}
      <div class="flex flex-col items-center gap-3 py-20">
        <p class="text-sm text-[var(--color-text-muted)]">{error}</p>
        {#if $user}
          <button type="button" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-4 py-2 text-sm text-white" onclick={openDashSettings}>
            Create dashboard
          </button>
        {/if}
      </div>
    {:else if dashboard}
      <DragDropProvider {onDragOver} onDragEnd={onDragEnd}>
        <div class="flex flex-col gap-4 {dashboard.layout === 'columns' ? 'md:flex-row md:flex-wrap md:items-start' : ''}">
          {#each filteredGroups as group, gIndex (group.id)}
            <div class={dashboard.layout === 'columns' ? 'min-w-[16rem] flex-1' : ''}>
              <GroupCard
                {group}
                index={gIndex}
                onEdit={openEditGroup}
                onDelete={(g) => askConfirm(`Delete group “${g.title}” and all its items?`, async () => { await api.deleteGroup(g.id); await load(); })}
                onAddItem={openNewItem}
              >
                {#each (itemsByGroup[group.id] || []) as item, iIndex (item.id)}
                  <ItemCard
                    {item}
                    index={iIndex}
                    groupId={group.id}
                    onEdit={openEditItem}
                    onDelete={(it) => askConfirm(`Delete “${it.title}”?`, async () => { await api.deleteItem(it.id); await load(); })}
                  />
                {/each}
              </GroupCard>
            </div>
          {/each}
        </div>

        <DragOverlay>
          {#snippet children(source)}
            {#if source?.data?.item}
              <ItemCard item={source.data.item} index={0} groupId={source.data.group} isOverlay />
            {:else if source?.data?.group}
              <GroupCard group={source.data.group} index={0} isOverlay />
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
    {/if}
  </AppLayout>
{/if}

<!-- Group modal -->
<Modal bind:open={groupOpen} title={editingGroup ? 'Edit group' : 'New group'}>
  <form onsubmit={(e) => { e.preventDefault(); saveGroup(); }}>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Title</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] {gTitleErr ? 'field-error' : ''}" bind:value={gTitle} />
      {#if gTitleErr}<span class="mt-1 text-xs text-[var(--color-danger)]">Required</span>{/if}
    </label>
    <div class="mt-5 flex justify-end gap-2">
      <button type="button" class="rounded-[var(--radius-btn)] px-3 py-2 text-sm text-[var(--color-text-muted)]" onclick={() => (groupOpen = false)}>Cancel</button>
      <button type="submit" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>

<!-- Item modal -->
<Modal bind:open={itemOpen} title={editingItem ? 'Edit item' : 'New item'} class="max-w-md">
  <form class="space-y-3" onsubmit={(e) => { e.preventDefault(); saveItem(); }}>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Title</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] {iTitleErr ? 'field-error' : ''}" bind:value={iTitle} />
      {#if iTitleErr}<span class="mt-1 text-xs text-[var(--color-danger)]">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">URL</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)] {iUrlErr ? 'field-error' : ''}" bind:value={iUrl} />
      {#if iUrlErr}<span class="mt-1 text-xs text-[var(--color-danger)]">Required</span>{/if}
    </label>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Description</span>
      <input class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm outline-none focus:border-[var(--color-primary)]" bind:value={iDesc} />
    </label>
    <div>
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Icon</span>
      <IconPicker bind:value={iIcon} seed={iTitle} />
    </div>
    <label class="block">
      <span class="mb-1 block text-xs text-[var(--color-text-muted)]">Size</span>
      <select class="w-full rounded-[var(--radius-btn)] border border-[var(--color-border)] bg-[var(--color-bg-elevated)] px-3 py-2 text-sm" bind:value={iSize}>
        <option value="1x1">1×1 — compact</option>
        <option value="1x2">1×2 — with description</option>
      </select>
    </label>
    <div class="flex justify-end gap-2 pt-2">
      <button type="button" class="rounded-[var(--radius-btn)] px-3 py-2 text-sm text-[var(--color-text-muted)]" onclick={() => (itemOpen = false)}>Cancel</button>
      <button type="submit" class="rounded-[var(--radius-btn)] bg-[var(--color-primary)] px-3 py-2 text-sm font-medium text-white">Save</button>
    </div>
  </form>
</Modal>

<!-- Dashboard settings -->
<Modal bind:open={dashOpen} title="Dashboard settings">
  <form class="space-y-3" onsubmit={(e) => { e.preventDefault(); saveDash(); }}>
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
  onConfirm={async () => { await confirmAction?.(); }}
/>
