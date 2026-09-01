<script lang="ts">
  import { onMount } from 'svelte';
  import { replace } from 'svelte-spa-router';
  import { api, type AdminOverview, type AdminUser, type Dashboard } from '../lib/api';
  import { user } from '../lib/stores';
  import { toast, toastError } from '../lib/toasts';
  import { toDashList, type DashListItem } from '../lib/dashboard-helpers';
  import AppLayout from '../layouts/AppLayout.svelte';
  import Modal from '../components/Modal.svelte';
  import ConfirmModal from '../components/ConfirmModal.svelte';
  import Icon from '../components/Icon.svelte';

  let overview = $state<AdminOverview | null>(null);
  let dashList = $state<DashListItem[]>([]);
  let loading = $state(true);
  let pageError = $state('');
  let editOpen = $state(false);
  let editingUser = $state<AdminUser | null>(null);
  let editUsername = $state('');
  let editPassword = $state('');
  let resetOIDC = $state(false);
  let savingUser = $state(false);
  let confirmOpen = $state(false);
  let confirmMessage = $state('');
  let confirmAction = $state<(() => Promise<void>) | null>(null);

  const flagDescriptions: Record<string, string> = {
    DEV_MODE: 'Enables development-mode behavior.',
    OIDC_INSECURE_SKIP_TLS_VERIFY: 'Disables TLS certificate verification for OIDC requests.',
    UPDATE_CHECK_ENABLED: 'Checks GitHub Releases for newer stable DashLit versions.',
    DISABLE_PASSWORD_REGISTRATION: 'Prevents new users from registering with a password.',
    DISABLE_OIDC_REGISTRATION: 'Prevents creation of new users through OIDC.',
    DISABLE_PASSWORD_LOGIN: 'Disables password sign-in and password registration.',
    DISABLE_OIDC_USER_MERGE: 'Prevents OIDC sign-in from linking to an existing user.',
  };

  async function load() {
    loading = true;
    pageError = '';
    try {
      overview = await api.adminOverview();
      dashList = toDashList(overview.dashboards);
    } catch (error: unknown) {
      pageError = error instanceof Error ? error.message : 'Could not load administration data';
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    if (!$user || $user.role !== 'admin') {
      replace($user ? '/' : '/login');
      return;
    }
    void load();
  });

  function accountMethod(account: AdminUser): string {
    if (account.hasPassword && account.hasOIDC) return 'Password + OIDC';
    if (account.hasOIDC) return 'OIDC';
    if (account.hasPassword) return 'Password';
    return 'None';
  }

  function openUser(userToEdit: AdminUser) {
    editingUser = userToEdit;
    editUsername = userToEdit.username;
    editPassword = '';
    resetOIDC = false;
    editOpen = true;
  }

  async function saveUser(event: SubmitEvent) {
    event.preventDefault();
    if (!editingUser) return;
    if (editingUser.hasOIDC && !resetOIDC) {
      toast.error('Confirm that the OIDC link will be reset');
      return;
    }
    savingUser = true;
    try {
      await api.adminUpdateUser(editingUser.id, {
        username: editUsername.trim(),
        newPassword: editPassword || undefined,
        resetOIDC: editingUser.hasOIDC ? resetOIDC : false,
      });
      editOpen = false;
      toast.success('User updated');
      await load();
    } catch (error: unknown) {
      toastError(error, 'Could not update user');
    } finally {
      savingUser = false;
    }
  }

  function ask(message: string, action: () => Promise<void>) {
    confirmMessage = message;
    confirmAction = action;
    confirmOpen = true;
  }

  async function deleteUser(account: AdminUser) {
    await api.adminDeleteUser(account.id);
    toast.success('User deleted');
    await load();
  }

  async function deleteDashboard(dashboard: Dashboard) {
    await api.deleteDashboard(dashboard.id);
    toast.success('Dashboard deleted');
    await load();
  }

  async function setSystemDefault(dashboard: Dashboard) {
    try {
      await api.setMain(dashboard.id);
      toast.success('System default dashboard updated');
      await load();
    } catch (error: unknown) {
      toastError(error, 'Could not set system default dashboard');
    }
  }
</script>

<AppLayout dashboards={dashList} currentSlug="" showSearch={false}>
  <div class="space-y-6">
    <div>
      <h1 class="text-xl font-semibold tracking-tight text-text">Administration</h1>
      <p class="mt-1 text-sm text-text-muted">Manage users, dashboards, and runtime policy flags.</p>
    </div>

    {#if loading}
      <p class="py-16 text-center text-sm text-text-muted">Loading…</p>
    {:else if pageError}
      <p class="rounded-card bg-danger-soft px-4 py-3 text-sm text-danger">{pageError}</p>
    {:else if overview}
      <section class="rounded-card border border-border-soft bg-surface">
        <div class="border-b border-border-soft px-5 py-4"><h2 class="text-sm font-semibold text-text">Users</h2></div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[42rem] text-left text-sm">
            <thead class="text-xs text-text-muted"><tr class="border-b border-border-soft"><th class="px-5 py-3 font-medium">Login</th><th class="px-5 py-3 font-medium">Role</th><th class="px-5 py-3 font-medium">Sign-in methods</th><th class="px-5 py-3 font-medium">Dashboards</th><th class="px-5 py-3 text-right font-medium">Actions</th></tr></thead>
            <tbody>
              {#each overview.users as account (account.id)}
                <tr class="border-b border-border-soft last:border-0 hover:bg-surface-2">
                  <td class="px-5 py-3 font-medium text-text">{account.username}</td>
                  <td class="px-5 py-3 capitalize text-text-muted">{account.role}</td>
                  <td class="px-5 py-3 text-text-muted">{accountMethod(account)}</td>
                  <td class="px-5 py-3 text-text-muted">{account.dashboardCount}</td>
                  <td class="px-5 py-3"><div class="flex justify-end gap-2"><button type="button" class="rounded-btn border border-border px-2.5 py-1.5 text-xs hover:bg-surface" onclick={() => openUser(account)}>Edit</button><button type="button" disabled={account.id === $user?.id} class="rounded-btn px-2.5 py-1.5 text-xs text-danger hover:bg-danger-soft disabled:cursor-not-allowed disabled:opacity-40" onclick={() => ask(`Delete user “${account.username}” and all their dashboards?`, () => deleteUser(account))}>Delete</button></div></td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-card border border-border-soft bg-surface">
        <div class="border-b border-border-soft px-5 py-4">
          <h2 class="text-sm font-semibold text-text">All dashboards</h2>
          <p class="mt-1 text-xs text-text-muted">The system default is shown to unauthenticated visitors when public, or to signed-in users without their own dashboards when visibility is “users”. A private dashboard cannot be the system default.</p>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[58rem] text-left text-sm">
            <thead class="text-xs text-text-muted"><tr class="border-b border-border-soft"><th class="px-5 py-3 font-medium">Dashboard</th><th class="px-5 py-3 font-medium">Owner</th><th class="px-5 py-3 font-medium">Description</th><th class="px-5 py-3 font-medium">Privacy</th><th class="px-5 py-3 font-medium">System default</th><th class="px-5 py-3 text-right font-medium">Actions</th></tr></thead>
            <tbody>
              {#each overview.dashboards as dashboard (dashboard.id)}
                <tr class="border-b border-border-soft last:border-0 hover:bg-surface-2">
                  <td class="px-5 py-3"><a href="#/{dashboard.slug}" class="flex items-center gap-2 font-medium text-text hover:text-primary">{#if dashboard.icon}<Icon icon={dashboard.icon} iconDark={dashboard.iconDark} size={20} />{/if}<span>{dashboard.name}</span></a></td>
                  <td class="px-5 py-3 text-text-muted">{dashboard.owner?.username || '—'}</td>
                  <td class="max-w-md px-5 py-3 text-text-muted">{dashboard.description || '—'}</td>
                  <td class="px-5 py-3 capitalize text-text-muted">{dashboard.privacy}</td>
                  <td class="px-5 py-3">{#if dashboard.isMain}<span class="text-primary">Current</span>{:else}<button type="button" disabled={dashboard.privacy === 'private'} title={dashboard.privacy === 'private' ? 'Private dashboards cannot be system defaults' : 'Set as system default'} class="rounded-btn border border-border px-2.5 py-1.5 text-xs hover:bg-surface disabled:cursor-not-allowed disabled:opacity-40" onclick={() => setSystemDefault(dashboard)}>Set default</button>{/if}</td>
                  <td class="px-5 py-3 text-right"><button type="button" class="rounded-btn px-2.5 py-1.5 text-xs text-danger hover:bg-danger-soft" onclick={() => ask(`Delete dashboard “${dashboard.name}”?`, () => deleteDashboard(dashboard))}>Delete</button></td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-card border border-border-soft bg-surface">
        <div class="border-b border-border-soft px-5 py-4"><h2 class="text-sm font-semibold text-text">Environment flags</h2><p class="mt-1 text-xs text-text-muted">Read-only effective values loaded from the process environment or .env file. Secrets are never exposed.</p></div>
        <div class="overflow-x-auto">
          <table class="w-full min-w-[42rem] text-left text-sm">
            <thead class="text-xs text-text-muted"><tr class="border-b border-border-soft"><th class="px-5 py-3 font-medium">Key</th><th class="px-5 py-3 font-medium">Value</th><th class="px-5 py-3 font-medium">Description</th></tr></thead>
            <tbody>
              {#each Object.entries(overview.flags) as [name, enabled]}
                <tr class="border-b border-border-soft last:border-0 hover:bg-surface-2">
                  <td class="px-5 py-3"><code class="text-xs text-text-muted">{name}</code></td>
                  <td class="px-5 py-3"><span class="inline-block rounded-full px-2 py-0.5 text-xs {enabled ? 'bg-primary/15 text-primary' : 'bg-surface-2 text-text-subtle'}">{enabled ? 'true' : 'false'}</span></td>
                  <td class="px-5 py-3 text-text-muted">{flagDescriptions[name] ?? '—'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      </section>
    {/if}
  </div>
</AppLayout>

<Modal bind:open={editOpen} title="Edit user">
  <form class="space-y-3" onsubmit={saveUser}>
    <label class="block"><span class="mb-1 block text-xs text-text-muted">Login</span><input class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm" bind:value={editUsername} maxlength="64" required /></label>
    <label class="block"><span class="mb-1 block text-xs text-text-muted">New password</span><input type="password" class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm" bind:value={editPassword} minlength="6" placeholder="Leave empty to keep current" /></label>
    {#if editingUser?.hasOIDC}
      <div class="rounded-lg bg-danger-soft p-3 text-sm text-danger">This account is linked to OIDC. Saving changes will remove that link. The user will need a password or a new OIDC registration to sign in again.</div>
      <label class="flex items-start gap-2 text-sm text-text"><input type="checkbox" class="mt-1" bind:checked={resetOIDC} /><span>I understand that the OIDC link will be reset.</span></label>
    {/if}
    <div class="flex justify-end gap-2 pt-2"><button type="button" class="rounded-btn px-3 py-2 text-sm text-text-muted" onclick={() => (editOpen = false)}>Cancel</button><button type="submit" disabled={savingUser} class="rounded-btn bg-primary px-3 py-2 text-sm font-medium text-primary-fg disabled:opacity-60">{savingUser ? 'Saving…' : 'Save'}</button></div>
  </form>
</Modal>

<ConfirmModal bind:open={confirmOpen} title="Confirm" message={confirmMessage} onConfirm={async () => confirmAction?.()} />
