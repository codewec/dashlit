import { writable } from 'svelte/store';
import type { User, Dashboard } from './api';

export const user = writable<User | null>(null);
export const editMode = writable(false);
export const theme = writable<'light' | 'dark' | 'system'>('dark');
export const currentDashboard = writable<Dashboard | null>(null);
export const searchQuery = writable('');

export function applyTheme(mode: 'light' | 'dark' | 'system') {
  const root = document.documentElement;
  let resolved: 'light' | 'dark' = mode === 'system'
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : mode;
  root.setAttribute('data-theme', resolved);
  localStorage.setItem('bd_theme', mode);
}
