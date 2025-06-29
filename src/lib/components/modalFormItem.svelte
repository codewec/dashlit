<script lang="ts">
	import type { Item } from '$lib/types';
	import { Button, ButtonGroup, Helper, Input, InputAddon, Label, Modal } from 'flowbite-svelte';
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
		<div>
			<Label for="input-addon" class="mb-2">Icon</Label>
			<ButtonGroup class="inline-flex w-full items-stretch">
				<Input bind:value={form.icon} type="text" name="icon" placeholder="URL or Icon name" />
				<span class="color-picker">
					<Input
						bind:value={form.iconColor}
						defaultValue="#808080"
						class="h-full w-20 !rounded-tl-none !rounded-bl-none border-l-0"
						type="color"
						name="iconColor"
						placeholder="URL or Icon name"
					/>
				</span>
			</ButtonGroup>
			<Helper class="text-sm">
				URL or Icon name from <a
					target="_blank"
					href="https://icon-sets.iconify.design/"
					class="text-primary-600 dark:text-primary-500 font-medium hover:underline">iconify</a
				>. The color is applied only to the icon from the iconify.
			</Helper>
		</div>
		<Label class="space-y-2">
			<span>Url</span>
			<Input bind:value={form.url} type="text" name="url" placeholder="Url" required />
		</Label>
		<Button type="submit" class="w-full">Save</Button>
	</form>
</Modal>
