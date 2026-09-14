<script lang="ts">
  import { onMount } from 'svelte'
  import { api, setToken } from './lib/api'
  import { user, hydrateThemeFromUser, systemInfo } from './lib/stores'
  import { routePath, startRouter } from './lib/router'
  import Login from './pages/Login.svelte'
  import DashboardView from './pages/DashboardView.svelte'
  import { Toaster } from 'svelte-french-toast'
  import CustomThemeModal from './components/CustomThemeModal.svelte'

  let ready = $state(false)
  const dashboardSlug = $derived(/^\/[^/]+$/.test($routePath) ? $routePath.slice(1) : null)

  onMount(() => {
    const stopRouter = startRouter()
    void (async () => {
      hydrateThemeFromUser(null)

      void api
        .systemInfo()
        .then((info) => systemInfo.set(info))
        .catch(() => systemInfo.set(null))

      // An OIDC callback creates an HttpOnly cookie session. Drop any older
      // local bearer token so it cannot mask the new cookie during /auth/me.
      if (new URLSearchParams(window.location.search).get('oidc') === '1') {
        setToken(null)
        history.replaceState(null, '', window.location.pathname)
      }

      try {
        const me = await api.me()
        user.set(me)
        hydrateThemeFromUser(me)
      } catch {
        setToken(null)
        user.set(null)
        hydrateThemeFromUser(null)
      }
      ready = true
    })()
    return stopRouter
  })
</script>

{#if ready}
  {#if $routePath === '/login'}
    <Login />
  {:else if $routePath === '/profile'}
    {#await import('./pages/Profile.svelte')}
      <div class="flex min-h-dvh items-center justify-center text-sm text-text-subtle">Loading…</div>
    {:then module}
      {@const Profile = module.default}
      <Profile />
    {:catch}
      <div class="flex min-h-dvh flex-col items-center justify-center gap-3 px-4 text-center">
        <p class="text-sm text-text-muted">Could not load profile</p>
        <button type="button" class="text-sm text-primary hover:underline" onclick={() => window.location.reload()}>Reload page</button>
      </div>
    {/await}
  {:else if $routePath === '/admin'}
    {#await import('./pages/Admin.svelte')}
      <div class="flex min-h-dvh items-center justify-center text-sm text-text-subtle">Loading…</div>
    {:then module}
      {@const Admin = module.default}
      <Admin />
    {:catch}
      <div class="flex min-h-dvh flex-col items-center justify-center gap-3 px-4 text-center">
        <p class="text-sm text-text-muted">Could not load administration page</p>
        <button type="button" class="text-sm text-primary hover:underline" onclick={() => window.location.reload()}>Reload page</button>
      </div>
    {/await}
  {:else if $routePath === '/'}
    <DashboardView />
  {:else if dashboardSlug}
    <DashboardView params={{ slug: dashboardSlug }} />
  {:else}
    <div class="flex min-h-dvh flex-col items-center justify-center gap-3 px-4 text-center">
      <p class="text-5xl font-semibold tracking-tight text-text">404</p>
      <p class="text-sm text-text-muted">Page not found</p>
      <a href="/" class="mt-2 text-sm text-primary hover:underline">Go to home</a>
    </div>
  {/if}
{/if}

<CustomThemeModal />

<Toaster
  position="top-right"
  toastOptions={{
    duration: 4000,
    style:
      'background: var(--color-surface); color: var(--color-text); border: 1px solid var(--color-border); font: 400 0.875rem/1.25rem var(--font-sans);',
  }}
/>
