const TOKEN_KEY = 'bd_token';

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY);
}
export function setToken(token: string | null) {
  if (token) localStorage.setItem(TOKEN_KEY, token);
  else localStorage.removeItem(TOKEN_KEY);
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = { ...(options.headers as Record<string, string>) };
  const token = getToken();
  if (token) headers['Authorization'] = `Bearer ${token}`;
  if (options.body && !(options.body instanceof FormData)) headers['Content-Type'] = 'application/json';
  const res = await fetch(`/api${path}`, { ...options, headers, credentials: 'include' });
  if (!res.ok) {
    let msg = res.statusText;
    try { msg = (await res.json()).error || msg; } catch {}
    throw new Error(msg);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export type User = { id: string; username: string; role: 'admin' | 'user'; authMethod: 'password' | 'oidc' };
export type AuthConfig = {
  passwordLoginEnabled: boolean;
  passwordRegistrationEnabled: boolean;
  oidcEnabled: boolean;
  oidcButtonTitle: string;
  legacyDashboardAvailable: boolean;
};
export type ItemSize = '1x1' | '1x2';
export type Layout = 'rows' | 'columns' | 'masonry';
export type Width = 'default' | 'wide';

export type Item = {
  id: string; groupId: string; title: string; description: string;
  url: string; icon: string; iconDark: string; pingEnabled: boolean;
  pingOnlyDown: boolean; pingUrl: string; pingSkipTls: boolean; position: number;
};
export type Group = {
  id: string; dashboardId: string; title: string; description: string;
  icon: string; iconDark: string; itemSize: ItemSize; position: number; items?: Item[];
};
export type Dashboard = {
  id: string; ownerId: string; name: string; slug: string;
  description: string; icon: string; iconDark: string;
  layout: Layout; width: Width; privacy: 'public' | 'private' | 'users';
  cleanMode: boolean; isMain: boolean; isDefault: boolean; groups?: Group[]; owner?: User;
};
export type AdminUser = {
  id: string; username: string; role: 'admin' | 'user';
  hasPassword: boolean; hasOIDC: boolean; dashboardCount: number;
};
export type AdminOverview = {
  users: AdminUser[];
  dashboards: Dashboard[];
  flags: Record<string, boolean>;
};
export type IconSearchResult = {
  name: string;
  icon: string;
  iconDark?: string;
  source: 'selfh.st' | 'Iconify';
  variant?: 'color' | 'monochrome';
};
export type SystemInfo = {
  version: string;
  commit: string;
  latestVersion?: string;
  updateAvailable?: boolean;
  releaseUrl?: string;
};

export const api = {
  systemInfo: () => request<SystemInfo>('/system/version'),
  authConfig: () => request<AuthConfig>('/auth/config'),
  login: (username: string, password: string) =>
    request<{ token: string; user: User }>('/auth/login', { method: 'POST', body: JSON.stringify({ username, password }) }),
  register: (username: string, password: string, importLegacy = false) =>
    request<{ token: string; user: User }>('/auth/register', { method: 'POST', body: JSON.stringify({ username, password, importLegacy }) }),
  logout: () => request('/auth/logout', { method: 'POST' }),
  me: () => request<User>('/auth/me'),
  updateProfile: (data: { username: string; newPassword?: string }) =>
    request<User>('/auth/profile', { method: 'PUT', body: JSON.stringify(data) }),
  listDashboards: () => request<Dashboard[]>('/dashboards'),
  getDashboard: (idOrSlug: string) => request<Dashboard>(`/dashboards/${idOrSlug}`),
  getMain: () => request<Dashboard | null>('/dashboards/main'),
  createDashboard: (data: Partial<Dashboard>) =>
    request<Dashboard>('/dashboards', { method: 'POST', body: JSON.stringify(data) }),
  updateDashboard: (id: string, data: Partial<Dashboard>) =>
    request<Dashboard>(`/dashboards/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteDashboard: (id: string) => request(`/dashboards/${id}`, { method: 'DELETE' }),
  setMain: (id: string) => request(`/dashboards/${id}/set-main`, { method: 'POST' }),
  setDefault: (id: string, isDefault: boolean) =>
    request(`/dashboards/${id}/set-default`, { method: 'POST', body: JSON.stringify({ isDefault }) }),
  cloneDashboard: (id: string) =>
    request<Dashboard>(`/dashboards/${id}/clone`, { method: 'POST' }),
  exportDashboard: async (id: string) => {
    const headers: Record<string, string> = {};
    const token = getToken();
    if (token) headers['Authorization'] = `Bearer ${token}`;
    const res = await fetch(`/api/dashboards/${id}/export`, { headers, credentials: 'include' });
    if (!res.ok) throw new Error('Export failed');
    return res.blob();
  },
  importDashboard: (data: unknown) =>
    request<Dashboard>('/dashboards/import', { method: 'POST', body: JSON.stringify(data) }),
  cloneGroup: (id: string) =>
    request<Group>(`/groups/${id}/clone`, { method: 'POST' }),
  cloneItem: (id: string) =>
    request<Item>(`/items/${id}/clone`, { method: 'POST' }),
  pingItem: (id: string) => request<{ reachable: boolean }>(`/items/${id}/ping`),
  createGroup: (dashboardId: string, data: Partial<Group> & { title: string }) =>
    request<Group>(`/dashboards/${dashboardId}/groups`, { method: 'POST', body: JSON.stringify(data) }),
  updateGroup: (id: string, data: Partial<Group>) =>
    request<Group>(`/groups/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteGroup: (id: string) => request(`/groups/${id}`, { method: 'DELETE' }),
  createItem: (groupId: string, data: Partial<Item>) =>
    request<Item>(`/groups/${groupId}/items`, { method: 'POST', body: JSON.stringify(data) }),
  updateItem: (id: string, data: Partial<Item> & { groupId?: string }) =>
    request<Item>(`/items/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  deleteItem: (id: string) => request(`/items/${id}`, { method: 'DELETE' }),
  updateLayout: (dashboardId: string, data: {
    groups: { id: string; position: number }[];
    items: { id: string; groupId: string; position: number }[];
  }) => request(`/dashboards/${dashboardId}/layout`, { method: 'PUT', body: JSON.stringify(data) }),
  uploadIcon: async (file: File) => {
    const fd = new FormData();
    fd.append('icon', file);
    return request<{ id: string; icon: string; url: string }>('/icons/upload', { method: 'POST', body: fd });
  },
  searchIconifyIcons: (query: string) =>
    request<IconSearchResult[]>(`/icons/search/iconify?q=${encodeURIComponent(query)}`),
  searchSelfhstIcons: (query: string) =>
    request<IconSearchResult[]>(`/icons/search/selfhst?q=${encodeURIComponent(query)}`),
  adminOverview: () => request<AdminOverview>('/admin/overview'),
  adminUpdateUser: (id: string, data: { username: string; newPassword?: string; resetOIDC?: boolean }) =>
    request(`/admin/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
  adminDeleteUser: (id: string) => request(`/admin/users/${id}`, { method: 'DELETE' }),
};
