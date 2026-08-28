import toast from 'svelte-french-toast';

export function errorMessage(error: unknown, fallback = 'Something went wrong'): string {
  return error instanceof Error && error.message ? error.message : fallback;
}

export function toastError(error: unknown, fallback?: string): void {
  toast.error(errorMessage(error, fallback));
}

export { toast };
