<script lang="ts">
	import { enhance } from '$app/forms';
	import toast from 'svelte-5-french-toast';
	import { page_title } from '$lib';
	import { Card, Button, Label, Input, Helper, DarkMode } from 'flowbite-svelte';
</script>

<svelte:head>
	<title>{page_title}</title>
</svelte:head>

<div
	class="flex h-screen items-center justify-center bg-gradient-to-r from-blue-50 via-indigo-50 to-sky-50 p-4 dark:from-slate-950 dark:via-gray-950 dark:to-zinc-950"
>
	<DarkMode class="absolute top-4 right-4 cursor-pointer" />
	<Card class="p-4 sm:p-6 md:p-8">
		<form
			method="POST"
			use:enhance={() => {
				const toastId = toast.loading('Checking...');
				return async ({ result, update }) => {
					await update();
					if (result.type === 'failure') {
						const message = (result.data?.error as string) ?? 'An error occurred';
						toast.error(message, { id: toastId });
					} else {
						toast.success('Welcome back!', { icon: '👋', id: toastId });
					}
				};
			}}
			action="?/login"
			class="flex flex-col space-y-6"
		>
			<Label class="space-y-2">
				<span>Your password</span>
				<Input type="password" name="password" placeholder="•••••" required />
			</Label>
			<Button type="submit" class="w-full">Login</Button>
		</form>
	</Card>
</div>
