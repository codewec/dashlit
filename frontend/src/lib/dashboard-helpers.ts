import type { Dashboard, Group, Item, ItemSize, Layout, Width } from './api';

export type DashListItem = {
  id: string;
  name: string;
  slug: string;
  icon?: string;
  iconDark?: string;
};

export function toDashList(list: Dashboard[]): DashListItem[] {
  return list.map((d) => ({
    id: d.id,
    name: d.name,
    slug: d.slug,
    icon: d.icon,
    iconDark: d.iconDark,
  }));
}

export function filterGroups(groups: Group[], query: string): Group[] {
  const q = query.trim().toLowerCase();
  if (!q) return groups;
  return groups
    .map((g) => ({
      ...g,
      items: (g.items ?? []).filter((it) => it.title.toLowerCase().includes(q) || (it.description || '').toLowerCase().includes(q)),
    }))
    .filter((g) => (g.items?.length ?? 0) > 0 || g.title.toLowerCase().includes(q));
}

export function itemsByGroupMap(groups: Group[]): Record<string, Item[]> {
  const map: Record<string, Item[]> = {};
  for (const g of groups) {
    map[g.id] = [...(g.items ?? [])].sort((a, b) => a.position - b.position);
  }
  return map;
}

export function groupsOuterClass(layout: Layout, wide: boolean): string {
  if (layout === 'columns') {
    return wide ? 'grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4' : 'grid grid-cols-1 gap-4 md:grid-cols-2';
  }
  if (layout === 'masonry') {
    return wide ? 'columns-1 gap-4 space-y-4 sm:columns-2 lg:columns-3 xl:columns-4 [&>*]:mb-4 [&>*]:break-inside-avoid' : 'columns-1 gap-4 space-y-4 md:columns-2 [&>*]:mb-4 [&>*]:break-inside-avoid';
  }
  return 'flex flex-col gap-4';
}

export function groupCellClass(layout: Layout): string {
  if (layout === 'columns') return 'min-w-0';
  if (layout === 'masonry') return 'break-inside-avoid';
  return '';
}

export function buildLayoutPayload(groups: Group[]) {
  const sortedGroups = groups.map((g, i) => ({ id: g.id, position: i }));
  const items: { id: string; groupId: string; position: number }[] = [];
  for (const g of groups) {
    (g.items ?? []).forEach((it, i) => items.push({ id: it.id, groupId: g.id, position: i }));
  }
  return { groups: sortedGroups, items };
}

export function reorderGroups(groups: Group[], fromId: string, toId: string): Group[] {
  const from = groups.findIndex((g) => g.id === fromId);
  const to = groups.findIndex((g) => g.id === toId);
  if (from < 0 || to < 0 || from === to) return groups;
  const next = [...groups];
  const [moved] = next.splice(from, 1);
  next.splice(to, 0, moved);
  return next;
}

export function applyItemMove(groups: Group[], bag: Record<string, Item[]>): Group[] {
  return groups.map((g) => ({
    ...g,
    items: (bag[g.id] || []).map((it, i) => ({ ...it, groupId: g.id, position: i })),
  }));
}

/* —— form drafts —— */

export type GroupForm = {
  title: string;
  description: string;
  icon: string;
  iconDark: string;
  itemSize: ItemSize;
};

export type ItemForm = {
  title: string;
  description: string;
  url: string;
  icon: string;
  iconDark: string;
  groupId: string;
};

export type DashboardForm = {
  name: string;
  slug: string;
  description: string;
  icon: string;
  iconDark: string;
  privacy: Dashboard['privacy'];
  layout: Layout;
  width: Width;
  cleanMode: boolean;
  creating: boolean;
};

export function emptyGroupForm(): GroupForm {
  return { title: '', description: '', icon: '', iconDark: '', itemSize: '1x1' };
}

export function groupToForm(g: Group): GroupForm {
  return {
    title: g.title,
    description: g.description || '',
    icon: g.icon || '',
    iconDark: g.iconDark || '',
    itemSize: g.itemSize || '1x1',
  };
}

export function emptyItemForm(groupId = ''): ItemForm {
  return {
    title: '',
    description: '',
    url: '',
    icon: 'mdi:link',
    iconDark: '',
    groupId,
  };
}

export function itemToForm(item: Item): ItemForm {
  return {
    title: item.title,
    description: item.description || '',
    url: item.url,
    icon: item.icon,
    iconDark: item.iconDark || '',
    groupId: item.groupId,
  };
}

export function emptyDashboardForm(creating = true): DashboardForm {
  return {
    name: '',
    slug: '',
    description: '',
    icon: '',
    iconDark: '',
    privacy: 'private',
    layout: 'rows',
    width: 'default',
    cleanMode: false,
    creating,
  };
}

export function dashboardToForm(d: Dashboard): DashboardForm {
  return {
    name: d.name,
    slug: d.slug,
    description: d.description || '',
    icon: d.icon || '',
    iconDark: d.iconDark || '',
    privacy: d.privacy,
    layout: d.layout,
    width: d.width || 'default',
    cleanMode: !!d.cleanMode,
    creating: false,
  };
}

export function pageContainerClass(wide: boolean): string {
  return wide ? 'mx-auto max-w-none px-4 pb-20 pt-6' : 'mx-auto max-w-6xl px-4 pb-20 pt-6';
}
