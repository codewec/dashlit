<script lang="ts">
	import { ButtonGroup, GradientButton, Button } from 'flowbite-svelte';

	let {
		editMode,
		canLogout,
		handleSave,
		handleEdit
	}: { editMode: boolean; canLogout: boolean; handleSave: () => void; handleEdit: () => void } =
		$props();
</script>

<div class="mb-8 flex justify-between gap-2">
	<h1 class="text-2xl font-bold text-gray-900">DashLit</h1>
	<ButtonGroup size="sm" class="*:ring-0!">
		{#if editMode}
			<GradientButton color="purpleToBlue" class="hover:text-white" onclick={() => handleSave()}
				>Save</GradientButton
			>
		{:else}
			<Button onclick={() => handleEdit()}>Edit</Button>
		{/if}
		{#if canLogout}
			<Button href="/logout">Logout</Button>
		{/if}
	</ButtonGroup>
</div>
