<script lang="ts">
	import type { Item } from '$lib/types';
	import { Button, Helper, Input, Label, Modal } from 'flowbite-svelte';
	import { slide } from 'svelte/transition';

	const {
		isOpen,
		item,
		handleClose
	}: { isOpen: boolean; item: Item; handleClose: (item: Item | undefined) => void } = $props();

	const form = $derived(item);
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
		<Label class="space-y-2">
			<span>Icon</span>
			<Input bind:value={form.icon} type="text" name="icon" placeholder="Icon name" />
			<Helper class="text-sm">
				Icon name from <a
					target="_blank"
					href="https://icon-sets.iconify.design/"
					class="text-primary-600 dark:text-primary-500 font-medium hover:underline">iconify</a
				>.
			</Helper>
		</Label>
		<Label class="space-y-2">
			<span>Url</span>
			<Input bind:value={form.url} type="text" name="url" placeholder="Url" required />
		</Label>
		<Button type="submit" class="w-full">Save</Button>
	</form>
</Modal>
