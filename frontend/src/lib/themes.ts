export type Theme = 'system' | 'crema' | 'latte' | 'frappe' | 'macchiato' | 'mocha';
export type ResolvedTheme = Exclude<Theme, 'system'>;

export type ThemeOption = {
  value: Theme;
  label: string;
  mode: 'system' | 'light' | 'dark';
  swatch: string;
};

export const themeOptions: ThemeOption[] = [
  { value: 'system', label: 'System', mode: 'system', swatch: 'linear-gradient(135deg, #f3f1ea 50%, #303446 50%)' },
  { value: 'crema', label: 'Crema', mode: 'light', swatch: '#f3f1ea' },
  { value: 'latte', label: 'Latte', mode: 'light', swatch: '#eff1f5' },
  { value: 'frappe', label: 'Frappé', mode: 'dark', swatch: '#303446' },
  { value: 'macchiato', label: 'Macchiato', mode: 'dark', swatch: '#24273a' },
  { value: 'mocha', label: 'Mocha', mode: 'dark', swatch: '#1e1e2e' },
];

export function normalizeTheme(value: string | null): Theme {
  if (themeOptions.some((option) => option.value === value)) return value as Theme;
  // Migrate values saved by the old light/dark toggle.
  if (value === 'light') return 'crema';
  if (value === 'dark') return 'frappe';
  return 'system';
}

export function themeLabel(value: Theme): string {
  return themeOptions.find((option) => option.value === value)?.label ?? value;
}
