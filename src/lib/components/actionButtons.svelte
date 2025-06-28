<script lang="ts">
	import { ActionType, type DeletionEntity } from '$lib/types';
	import { Button } from 'flowbite-svelte';
	import { EditOutline, CloseCircleOutline } from 'flowbite-svelte-icons';

	const {
		id,
		handleHover,
		handleClick
	}: {
		id: string;
		handleHover: (id: string | undefined) => void;
		handleClick: (type: ActionType) => void;
	} = $props();
</script>

<div class="inline-flex gap-1">
	<Button
		pill={true}
		color="light"
		onmouseenter={() => {
			handleHover(id);
		}}
		onmouseleave={() => {
			handleHover(undefined);
		}}
		onclick={() => {
			handleClick(ActionType.EDIT);
		}}
		class="h-4 w-4 cursor-pointer p-3!"><EditOutline class="h-4 w-4" color="gray" /></Button
	>
	<Button
		pill={true}
		color="light"
		onmouseenter={() => {
			handleHover(id);
		}}
		onmouseleave={() => {
			handleHover(undefined);
		}}
		onclick={() => {
			handleClick(ActionType.DELETE);
		}}
		class="h-4 w-4 cursor-pointer p-3!"><CloseCircleOutline class="h-4 w-4" color="gray" /></Button
	>
</div>
