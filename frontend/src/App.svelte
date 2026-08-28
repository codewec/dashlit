<script lang="ts">
  import { onMount } from 'svelte';
  import Router from 'svelte-spa-router';
  import { api, setToken } from './lib/api';
  import { user, theme, applyTheme } from './lib/stores';
  import Login from './pages/Login.svelte';
  import DashboardView from './pages/DashboardView.svelte';
  import Profile from './pages/Profile.svelte';
  import { Toaster } from 'svelte-french-toast';
  import { normalizeTheme } from './lib/themes';

  const routes = {
    '/login': Login,
    '/profile': Profile,
    '/': DashboardView,
    '/:slug': DashboardView,
  };

  let ready = $state(false);

  onMount(async () => {
    const saved = normalizeTheme(localStorage.getItem('bd_theme'));
    theme.set(saved);
    applyTheme(saved);

    // An OIDC callback creates an HttpOnly cookie session. Drop any older
    // local bearer token so it cannot mask the new cookie during /auth/me.
    if (new URLSearchParams(window.location.search).get('oidc') === '1') {
      setToken(null);
      history.replaceState(null, '', `${window.location.pathname}${window.location.hash}`);
    }

    try {
      user.set(await api.me());
    } catch {
      setToken(null);
      user.set(null);
    }
    ready = true;
  });
</script>

{#if ready}
  <Router {routes} />
{/if}

<Toaster
  position="top-right"
  toastOptions={{
    duration: 4000,
    style: 'background: var(--color-surface); color: var(--color-text); border: 1px solid var(--color-border);',
  }}
/>
