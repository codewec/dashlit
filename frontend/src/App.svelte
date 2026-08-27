<script lang="ts">
  import { onMount } from 'svelte';
  import Router from 'svelte-spa-router';
  import { api, setToken, getToken } from './lib/api';
  import { user, theme, applyTheme } from './lib/stores';
  import Login from './pages/Login.svelte';
  import DashboardView from './pages/DashboardView.svelte';

  const routes = {
    '/login': Login,
    '/': DashboardView,
    '/:slug': DashboardView,
  };

  let ready = $state(false);

  onMount(async () => {
    const saved = (localStorage.getItem('bd_theme') as 'light' | 'dark' | 'system') || 'dark';
    theme.set(saved);
    applyTheme(saved);

    if (getToken()) {
      try {
        user.set(await api.me());
      } catch {
        setToken(null);
        user.set(null);
      }
    }
    ready = true;
  });
</script>

{#if ready}
  <Router {routes} />
{/if}
