<script lang="ts">
  import { onMount } from 'svelte';
  import { api, setToken, type AuthConfig } from '../lib/api';
  import { user } from '../lib/stores';
  import { push } from 'svelte-spa-router';
  import AuthLayout from '../layouts/AuthLayout.svelte';
  import ThemeFab from '../components/ThemeFab.svelte';
  import logoUrl from '../assets/dashlit.svg';
  import { toastError } from '../lib/toasts';

  let username = $state('');
  let password = $state('');
  let mode = $state<'login' | 'register'>('login');
  let error = $state('');
  let loading = $state(false);
  let touched = $state({ username: false, password: false });
  let authConfig = $state<AuthConfig | null>(null);
  let configLoading = $state(true);
  let oidcLoginURL = $state('/api/auth/oidc/login');

  const usernameError = $derived(touched.username && !username.trim() ? 'Required' : '');
  const passwordError = $derived(touched.password && password.length < 6 ? 'Min 6 characters' : '');
  const passwordFormEnabled = $derived(!!authConfig && (mode === 'register' ? authConfig.passwordRegistrationEnabled : authConfig.passwordLoginEnabled));

  onMount(async () => {
    oidcLoginURL = `/api/auth/oidc/login?return_to=${encodeURIComponent(window.location.origin + '/')}`;
    const query = window.location.hash.split('?', 2)[1];
    const oidcError = query ? new URLSearchParams(query).get('oidc_error') : null;
    if (oidcError) {
      error = oidcError;
      history.replaceState(null, '', `${window.location.pathname}${window.location.search}#/login`);
    }
    try {
      authConfig = await api.authConfig();
      if (!authConfig.passwordRegistrationEnabled) mode = 'login';
    } catch (err: unknown) {
      error = err instanceof Error ? err.message : 'Could not load authentication settings';
    } finally {
      configLoading = false;
    }
  });

  async function submit(e: Event) {
    e.preventDefault();
    if (!passwordFormEnabled) return;
    touched = { username: true, password: true };
    if (!username.trim() || password.length < 6) return;
    error = '';
    loading = true;
    try {
      const res = mode === 'login' ? await api.login(username, password) : await api.register(username, password);
      setToken(res.token);
      user.set(res.user);
      push('/');
    } catch (err: unknown) {
      toastError(err, mode === 'login' ? 'Sign in failed' : 'Registration failed');
    } finally {
      loading = false;
    }
  }
</script>

<AuthLayout>
  <form class="rounded-card border border-border-soft bg-surface p-6 shadow-xl" onsubmit={submit}>
    <div class="mb-6 text-center">
      <img src={logoUrl} alt="" class="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-xl" />
      <h1 class="text-xl font-semibold tracking-tight">DashLit</h1>
      <p class="mt-1 text-sm text-text-muted">
        {mode === 'login' ? 'Sign in to your dashboard' : 'Create an account (first user is admin)'}
      </p>
    </div>

    {#if error}
      <p class="mb-3 rounded-lg bg-danger-soft px-3 py-2 text-sm text-danger">{error}</p>
    {/if}

    {#if configLoading}
      <div class="py-8 text-center text-sm text-text-muted">Loading…</div>
    {:else if authConfig}
      {#if passwordFormEnabled}
        <label class="mb-3 block">
          <span class="mb-1 block text-xs font-medium text-text-muted">Username</span>
          <input
            class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2.5 text-sm outline-none focus:border-primary {usernameError ? 'field-error' : ''}"
            bind:value={username}
            onblur={() => (touched.username = true)}
            autocomplete="username"
            required
          />
          {#if usernameError}<span class="mt-1 block text-xs text-danger">{usernameError}</span>{/if}
        </label>

        <label class="mb-4 block">
          <span class="mb-1 block text-xs font-medium text-text-muted">Password</span>
          <input
            type="password"
            class="w-full rounded-btn border border-border bg-bg-elevated px-3 py-2.5 text-sm outline-none focus:border-primary {passwordError ? 'field-error' : ''}"
            bind:value={password}
            onblur={() => (touched.password = true)}
            autocomplete={mode === 'login' ? 'current-password' : 'new-password'}
            required
            minlength="6"
          />
          {#if passwordError}<span class="mt-1 block text-xs text-danger">{passwordError}</span>{/if}
        </label>

        <button type="submit" disabled={loading} class="w-full rounded-btn bg-primary py-2.5 text-sm font-medium text-white hover:bg-primary-hover disabled:opacity-60">
          {loading ? '…' : mode === 'login' ? 'Sign in' : 'Create account'}
        </button>
      {/if}

      {#if authConfig.oidcEnabled && mode === 'login'}
        {#if authConfig.passwordLoginEnabled}
          <div class="my-4 flex items-center gap-3 text-xs text-text-subtle before:h-px before:flex-1 before:bg-border after:h-px after:flex-1 after:bg-border">or</div>
        {/if}
        <a href={oidcLoginURL} class="flex w-full items-center justify-center gap-2 rounded-btn border border-border bg-bg-elevated py-2.5 text-sm font-medium text-text hover:bg-surface-2">
          <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"
            ><path d="M12 3a9 9 0 1 0 9 9" /><path d="M12 7a5 5 0 1 0 5 5" /><circle cx="12" cy="12" r="1" fill="currentColor" /></svg
          >
          {authConfig.oidcButtonTitle}
        </a>
      {/if}

      {#if (authConfig.passwordRegistrationEnabled && mode === 'login') || mode === 'register'}
        <button
          type="button"
          class="mt-3 w-full py-2 text-center text-sm text-text-muted hover:text-text"
          onclick={() => {
            mode = mode === 'login' ? 'register' : 'login';
            error = '';
          }}
        >
          {mode === 'login' ? 'Need an account? Register' : 'Already have an account? Sign in'}
        </button>
      {/if}

      {#if !authConfig.passwordLoginEnabled && !authConfig.oidcEnabled && !authConfig.passwordRegistrationEnabled}
        <p class="text-center text-sm text-text-muted">No login method is available.</p>
      {/if}
    {/if}
  </form>
</AuthLayout>
<ThemeFab />
