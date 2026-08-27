<script lang="ts">
  import { DragDropProvider, DragOverlay } from '@dnd-kit-svelte/svelte';
  import { move } from '@dnd-kit/helpers';
  import type { Dashboard, Group, Item } from '../lib/api';
  import { editMode } from '../lib/stores';
  import { filterGroups, itemsByGroupMap, groupsOuterClass, groupCellClass, reorderGroups, applyItemMove } from '../lib/dashboard-helpers';
  import GroupCard from './GroupCard.svelte';
  import ItemCard from './ItemCard.svelte';
  import { searchQuery } from '../lib/stores';

  let {
    dashboard,
    groups = $bindable([]),
    onEditGroup,
    onDeleteGroup,
    onCloneGroup,
    onAddItem,
    onEditItem,
    onDeleteItem,
    onCloneItem,
    onLayoutChange,
  }: {
    dashboard: Dashboard;
    groups: Group[];
    onEditGroup: (g: Group) => void;
    onDeleteGroup: (g: Group) => void;
    onCloneGroup: (g: Group) => void;
    onAddItem: (g: Group) => void;
    onEditItem: (item: Item) => void;
    onDeleteItem: (item: Item) => void;
    onCloneItem: (item: Item) => void;
    onLayoutChange: () => void | Promise<void>;
  } = $props();

  const filtered = $derived(filterGroups(groups, $searchQuery));
  const byGroup = $derived(itemsByGroupMap(filtered));
  const outerClass = $derived(groupsOuterClass(dashboard.layout, dashboard.width === 'wide'));
  const cellClass = $derived(groupCellClass(dashboard.layout));

  function onDragOver(event: any) {
    if (!$editMode) return;
    const { source } = event.operation;
    if (source?.type === 'column') {
      const targetId = event.operation.target?.id as string | undefined;
      if (!targetId) return;
      groups = reorderGroups(groups, source.id as string, targetId);
      return;
    }
    const bag: Record<string, Item[]> = {};
    for (const g of groups) bag[g.id] = [...(g.items ?? [])];
    const nextBag = move(bag, event) as Record<string, Item[]>;
    groups = applyItemMove(groups, nextBag);
  }

  async function onDragEnd() {
    if (!$editMode) return;
    await onLayoutChange();
  }
</script>

<DragDropProvider {onDragOver} {onDragEnd}>
  <div class={outerClass}>
    {#each filtered as group, gIndex (group.id)}
      <div class={cellClass}>
        <GroupCard {group} index={gIndex} layout={dashboard.layout} wide={dashboard.width === 'wide'} onEdit={onEditGroup} onDelete={onDeleteGroup} onClone={onCloneGroup} {onAddItem}>
          {#each byGroup[group.id] || [] as item, iIndex (item.id)}
            <ItemCard {item} index={iIndex} groupId={group.id} itemSize={group.itemSize} onEdit={onEditItem} onDelete={onDeleteItem} onClone={onCloneItem} />
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
