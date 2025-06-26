<script lang="ts">
	import type { DeletionEntity } from '$lib/types';
	import { Button } from 'flowbite-svelte';
	import { ExclamationCircleOutline } from 'flowbite-svelte-icons';
	import Modal from 'flowbite-svelte/Modal.svelte';
	import { slide } from 'svelte/transition';

	const {
		entity,
		handleClose,
		handleConfirm
	}: {
		entity: DeletionEntity | undefined;
		handleClose: () => void;
		handleConfirm: () => void;
	} = $props();
</script>

<Modal open={entity !== undefined} onclose={() => handleClose()} transition={slide} size="xs">
	{#if entity}
		<div class="text-center">
			<ExclamationCircleOutline class="mx-auto mb-4 h-12 w-12 text-gray-400 dark:text-gray-200" />
			<h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
				Are you sure you want to delete "{entity.element.title}"?
			</h3>
			<Button color="red" onclick={() => handleConfirm()} class="me-2">Yes, I'm sure</Button>
			<Button color="alternative" onclick={() => handleClose()}>No, cancel</Button>
		</div>
	{/if}
</Modal>
