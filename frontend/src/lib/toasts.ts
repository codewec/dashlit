import baseToast, { type ToastOptions } from 'svelte-french-toast'

const iconTheme = {
  primary: 'var(--color-primary)',
  secondary: 'var(--color-primary-fg)',
}

function themedOptions(options?: ToastOptions): ToastOptions {
  return { ...options, iconTheme: options?.iconTheme ?? iconTheme }
}

type ToastMessage = Parameters<typeof baseToast.success>[0]

export const toast = {
  success(message: ToastMessage, options?: ToastOptions) {
    return baseToast.success(message, themedOptions(options))
  },
  error(message: ToastMessage, options?: ToastOptions) {
    return baseToast.error(message, themedOptions(options))
  },
  loading(message: ToastMessage, options?: ToastOptions) {
    return baseToast.loading(message, themedOptions(options))
  },
  dismiss: baseToast.dismiss,
  remove: baseToast.remove,
}

export function errorMessage(error: unknown, fallback = 'Something went wrong'): string {
  return error instanceof Error && error.message ? error.message : fallback
}

export function toastError(error: unknown, fallback?: string): void {
  toast.error(errorMessage(error, fallback))
}
