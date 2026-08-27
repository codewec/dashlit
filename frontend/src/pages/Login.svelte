<script lang="ts">
  import { api, setToken } from '../lib/api';
  import { user } from '../lib/stores';
  import { push } from 'svelte-spa-router';
  import AuthLayout from '../layouts/AuthLayout.svelte';
  import ThemeFab from '../components/ThemeFab.svelte';

  let username = $state('');
  let password = $state('');
  let mode = $state<'login' | 'register'>('login');
  let error = $state('');
  let loading = $state(false);
  let touched = $state({ username: false, password: false });

  const usernameError = $derived(touched.username && !username.trim() ? 'Required' : '');
  const passwordError = $derived(touched.password && password.length < 6 ? 'Min 6 characters' : '');

  async function submit(e: Event) {
    e.preventDefault();
    touched = { username: true, password: true };
    if (!username.trim() || password.length < 6) return;
    error = '';
    loading = true;
    try {
      const res = mode === 'login' ? await api.login(username, password) : await api.register(username, password);
      setToken(res.token);
      user.set(res.user);
      push('/');
    } catch (err: any) {
      error = err.message || 'Failed';
    } finally {
      loading = false;
    }
  }
</script>

<AuthLayout>
  <form class="rounded-card border border-border-soft bg-surface p-6 shadow-xl" onsubmit={submit}>
    <div class="mb-6 text-center">
      <div class="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-xl bg-primary text-sm font-bold text-white">B</div>
      <h1 class="text-xl font-semibold tracking-tight">DashLit</h1>
      <p class="mt-1 text-sm text-text-muted">
        {mode === 'login' ? 'Sign in to your dashboard' : 'Create an account (first user is admin)'}
      </p>
    </div>

    {#if error}
      <p class="mb-3 rounded-lg bg-danger-soft px-3 py-2 text-sm text-danger">{error}</p>
    {/if}

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
  </form>
</AuthLayout>
<ThemeFab />
