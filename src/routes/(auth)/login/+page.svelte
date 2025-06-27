<script lang="ts">
	import { enhance } from '$app/forms';
	import { Card, Button, Label, Input, Helper } from 'flowbite-svelte';

	let errorMessage: string | undefined = $state(undefined);
</script>

<div
	class="flex h-screen items-center justify-center bg-gradient-to-r from-indigo-200 via-purple-200 to-pink-200 p-4"
>
	<Card class="p-4 sm:p-6 md:p-8">
		<form
			method="POST"
			use:enhance={() => {
				errorMessage = undefined;
				return async ({ result, update }) => {
					await update();
					if (result.type === 'failure') {
						errorMessage = (result.data?.error as string) ?? 'An error occurred';
					}
				};
			}}
			action="?/login"
			class="flex flex-col space-y-6"
		>
			<Label class="space-y-2">
				<span>Your password</span>
				<Input type="password" name="password" placeholder="•••••" required />
				{#if errorMessage}
					<Helper class="mt-2" color="red">
						<span class="font-medium">Error: </span>
						{errorMessage}
					</Helper>
				{/if}
			</Label>
			<Button type="submit" class="w-full">Login</Button>
		</form>
	</Card>
</div>
