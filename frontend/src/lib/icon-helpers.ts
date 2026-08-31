import type { IconSearchResult } from './api';

export type IconVariant = 'light' | 'dark';
export type IconSourceTab = 'search' | 'url' | 'upload';

export function detectIconSourceTab(value: string): IconSourceTab {
  if (value.startsWith('local:')) return 'upload';
  if (value.startsWith('http://') || value.startsWith('https://') || value.startsWith('/')) return 'url';
  return 'search';
}

export function iconSrc(icon: string): string {
  if (!icon) return '';
  if (icon.startsWith('http') || icon.startsWith('/')) return icon;
  if (icon.startsWith('local:')) return `/api/icons/${icon.slice(6)}`;
  if (icon.startsWith('selfhst-icon:')) {
    return `/api/icons/selfhst/${encodeURIComponent(icon.slice(13))}`;
  }
  if (icon.includes(':')) {
    const [prefix, name] = icon.split(':');
    return `/api/icons/iconify/${prefix}/${name}`;
  }
  return icon;
}

export function resolveIcon(icon: string, iconDark: string | undefined, isDark: boolean): string {
  return isDark && iconDark ? iconDark : icon || '';
}

export function isIconifyIcon(icon: string): boolean {
  return !!icon
    && !icon.startsWith('http')
    && !icon.startsWith('/')
    && !icon.startsWith('local:')
    && !icon.startsWith('selfhst-icon:')
    && icon.includes(':');
}

export function iconForVariant(result: IconSearchResult, variant: IconVariant): string {
  return variant === 'dark' ? result.iconDark || result.icon : result.icon;
}

export function iconSearchTitle(result: IconSearchResult): string {
  const variant = result.variant === 'color'
    ? 'Color'
    : result.variant === 'monochrome' ? 'Monochrome' : '';
  return [result.name, variant, result.source].filter(Boolean).join(' · ');
}

export function iconPreviewClass(variant: IconVariant): string {
  return variant === 'dark' ? 'bg-[#181825]' : 'bg-white';
}

export function applyIconSelection(
  result: IconSearchResult,
  variant: IconVariant,
  pairedValue: string,
): { value: string; pairedValue: string } {
  if (result.variant === 'color') {
    return {
      value: result.icon,
      pairedValue: variant === 'light' ? '' : pairedValue,
    };
  }
  return {
    value: iconForVariant(result, variant),
    pairedValue: pairedValue || (variant === 'dark' ? result.icon : result.iconDark || ''),
  };
}

export function combineIconResults(
  selfhst: IconSearchResult[],
  iconify: IconSearchResult[],
): IconSearchResult[] {
  return [...selfhst, ...iconify];
}
