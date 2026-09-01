<script lang="ts">
  import { onMount } from 'svelte'
  import { replace } from 'svelte-spa-router'
  import { api, type AuthConfig, type Dashboard } from '../lib/api'
  import { user } from '../lib/stores'
  import { toast, toastError } from '../lib/toasts'
  import { toDashList, type DashListItem } from '../lib/dashboard-helpers'
  import AppLayout from '../layouts/AppLayout.svelte'
  import Icon from '../components/Icon.svelte'

  let dashboards = $state<Dashboard[]>([])
  let dashList = $state<DashListItem[]>([])
  let username = $state('')
  let newPassword = $state('')
  let confirmPassword = $state('')
  let loading = $state(true)
  let savingUsername = $state(false)
  let savingPassword = $state(false)
  let pageError = $state('')
  let authConfig = $state<AuthConfig | null>(null)

  const ownDashboards = $derived(dashboards.filter((dashboard) => dashboard.ownerId === $user?.id))

  onMount(async () => {
    if (!$user) {
      replace('/login')
      return
    }
    username = $user.username
    try {
      const [availableDashboards, config] = await Promise.all([api.listDashboards(), api.authConfig()])
      dashboards = availableDashboards ?? []
      dashList = toDashList(dashboards)
      authConfig = config
    } catch (error: unknown) {
      pageError = error instanceof Error ? error.message : 'Could not load dashboards'
    } finally {
      loading = false
    }
  })

  async function saveUsername(event: SubmitEvent) {
    event.preventDefault()
    if ($user?.authMethod === 'oidc') return
    if (!username.trim()) return
    savingUsername = true
    try {
      const updated = await api.updateProfile({ username: username.trim() })
      user.set(updated)
      username = updated.username
      toast.success('Username updated')
    } catch (error: unknown) {
      toastError(error, 'Could not update username')
    } finally {
      savingUsername = false
    }
  }

  async function savePassword(event: SubmitEvent) {
    event.preventDefault()
    if (newPassword.length < 6) {
      toast.error('New password must be at least 6 characters')
      return
    }
    if (newPassword !== confirmPassword) {
      toast.error('Passwords do not match')
      return
    }
    savingPassword = true
    try {
      const updated = await api.updateProfile({
        username: $user?.username ?? username.trim(),
        newPassword,
      })
      user.set(updated)
      newPassword = ''
      confirmPassword = ''
      toast.success('Password updated')
    } catch (error: unknown) {
      toastError(error, 'Could not update password')
    } finally {
      savingPassword = false
    }
  }
</script>

<AppLayout dashboards={dashList} currentSlug="" showSearch={false}>
  <div class="space-y-6">
    <div>
      <h1 class="text-xl font-semibold tracking-tight text-text">Profile</h1>
      <p class="mt-1 text-sm text-text-muted">Manage your account and dashboards.</p>
      {#if $user}
        <p class="mt-2 inline-flex items-center gap-1.5 rounded-full border border-border bg-surface px-2.5 py-1 text-xs text-text-muted">
          Signed in with <span class="font-medium text-text">{$user.authMethod === 'oidc' ? 'OIDC' : 'password'}</span>
        </p>
      {/if}
    </div>

    <div class="grid gap-4 {authConfig?.passwordLoginEnabled ? 'lg:grid-cols-2' : ''}">
      <form class="rounded-card border border-border-soft bg-surface p-5" onsubmit={saveUsername}>
        <h2 class="text-sm font-semibold text-text">Username</h2>
        <label class="mt-4 block">
          <span class="mb-1 block text-xs text-text-muted">Login</span>
          <input
            class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary disabled:cursor-not-allowed disabled:opacity-60"
            bind:value={username}
            maxlength="64"
            required
            autocomplete="username"
            disabled={$user?.authMethod === 'oidc'}
          />
          {#if $user?.authMethod === 'oidc'}
            <span class="mt-1 block text-[11px] text-text-subtle">Username is managed by the OIDC provider.</span>
          {/if}
        </label>
        {#if $user?.authMethod !== 'oidc'}
          <div class="mt-4 flex justify-end">
            <button
              type="submit"
              disabled={savingUsername}
              class="rounded-btn bg-primary px-4 py-2 text-sm font-medium text-primary-fg hover:bg-primary-hover disabled:opacity-60"
            >
              {savingUsername ? 'Saving…' : 'Save username'}
            </button>
          </div>
        {/if}
      </form>

      {#if authConfig?.passwordLoginEnabled}
        <form class="rounded-card border border-border-soft bg-surface p-5" onsubmit={savePassword}>
          <h2 class="text-sm font-semibold text-text">Password</h2>
          <div class="mt-4 space-y-3">
            <label class="block">
              <span class="mb-1 block text-xs text-text-muted">New password</span>
              <input
                type="password"
                class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary"
                bind:value={newPassword}
                minlength="6"
                required
                autocomplete="new-password"
              />
            </label>
            <label class="block">
              <span class="mb-1 block text-xs text-text-muted">Confirm new password</span>
              <input
                type="password"
                class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2 text-sm outline-none focus:border-primary"
                bind:value={confirmPassword}
                minlength="6"
                required
                autocomplete="new-password"
              />
            </label>
          </div>
          <div class="mt-4 flex justify-end">
            <button
              type="submit"
              disabled={savingPassword}
              class="rounded-btn bg-primary px-4 py-2 text-sm font-medium text-primary-fg hover:bg-primary-hover disabled:opacity-60"
            >
              {savingPassword ? 'Saving…' : 'Change password'}
            </button>
          </div>
        </form>
      {/if}
    </div>

    <section class="rounded-card border border-border-soft bg-surface">
      <div class="border-b border-border-soft px-5 py-4">
        <h2 class="text-sm font-semibold text-text">My dashboards</h2>
      </div>
      {#if loading}
        <p class="px-5 py-8 text-center text-sm text-text-muted">Loading…</p>
      {:else if pageError}
        <p class="px-5 py-8 text-center text-sm text-danger">{pageError}</p>
      {:else if ownDashboards.length === 0}
        <p class="px-5 py-8 text-center text-sm text-text-muted">You do not have any dashboards yet.</p>
      {:else}
        <div class="overflow-x-auto">
          <table class="w-full min-w-3xl text-left text-sm">
            <thead class="text-xs text-text-muted">
              <tr class="border-b border-border-soft">
                <th class="px-5 py-3 font-medium">Name</th>
                <th class="px-5 py-3 font-medium">Description</th>
                <th class="px-5 py-3 font-medium">Slug</th>
                <th class="px-5 py-3 font-medium">Privacy</th>
                <th class="px-5 py-3 font-medium">Default</th>
              </tr>
            </thead>
            <tbody>
              {#each ownDashboards as dashboard (dashboard.id)}
                <tr class="border-b border-border-soft last:border-0 hover:bg-surface-2">
                  <td class="px-5 py-3 font-medium">
                    <a class="flex items-center gap-2 text-text hover:text-primary" href="#/{dashboard.slug}">
                      {#if dashboard.icon}<Icon icon={dashboard.icon} iconDark={dashboard.iconDark} size={20} class="shrink-0 rounded" />{/if}
                      <span>{dashboard.name}</span>
                    </a>
                  </td>
                  <td class="max-w-md px-5 py-3 text-text-muted">{dashboard.description || '—'}</td>
                  <td class="px-5 py-3 text-text-muted">{dashboard.slug}</td>
                  <td class="px-5 py-3 capitalize text-text-muted">{dashboard.privacy}</td>
                  <td class="px-5 py-3 text-text-muted">{dashboard.isDefault ? 'Yes' : '—'}</td>
                </tr>
              {/each}
            </tbody>
          </table>
        </div>
      {/if}
    </section>
  </div>
</AppLayout>
