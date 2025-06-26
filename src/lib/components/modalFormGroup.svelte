<script lang="ts">
	import type { Group } from '$lib/types';
	import { Button, Input, Label, Modal } from 'flowbite-svelte';
	import { slide } from 'svelte/transition';

	const {
		isOpen,
		group,
		handleClose
	}: { isOpen: boolean; group: Group; handleClose: (group: Group | undefined) => void } = $props();

	const form = $derived(group);
</script>

<Modal open={isOpen} onclose={() => handleClose(undefined)} transition={slide} size="xs">
	<form
		class="flex flex-col space-y-6 pt-4"
		onsubmit={(e: SubmitEvent) => {
			e.preventDefault();
			handleClose(form);
		}}
	>
		<Label class="space-y-2">
			<span>Title</span>
			<Input bind:value={form.title} type="text" name="title" placeholder="Title" required />
		</Label>
		<Label class="space-y-2">
			<span>Description</span>
			<Input
				bind:value={form.description}
				type="text"
				name="description"
				placeholder="Description"
			/>
		</Label>
		<Button type="submit" class="w-full">Save</Button>
	</form>
</Modal>
