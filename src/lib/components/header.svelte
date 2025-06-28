<script lang="ts">
	import { ButtonGroup, GradientButton, Button, DarkMode } from 'flowbite-svelte';

	let {
		editMode,
		canLogout,
		handleSave,
		handleEdit
	}: { editMode: boolean; canLogout: boolean; handleSave: () => void; handleEdit: () => void } =
		$props();
</script>

<div class="mb-8 flex justify-between gap-2">
	<h1 class="text-2xl font-bold">DashLit</h1>

	<div class="inline-flex items-center">
		<DarkMode size="sm" class="cursor-pointer p-2" />
		<ButtonGroup size="sm" class="*:ring-0!">
			{#if editMode}
				<GradientButton
					color="purpleToBlue"
					class="cursor-pointer hover:text-white"
					onclick={() => handleSave()}>Save</GradientButton
				>
			{:else}
				<Button class="cursor-pointer" onclick={() => handleEdit()}>Edit</Button>
			{/if}
			{#if canLogout}
				<Button data-sveltekit-preload-data="off" href="/logout">Logout</Button>
			{/if}
		</ButtonGroup>
	</div>
</div>
