<script lang="ts">
	import Button from 'flowbite-svelte/Button.svelte';
	const { handleClick }: { handleClick: () => void } = $props();
</script>

<Button
	onclick={handleClick}
	color="light"
	class="cursor-pointer rounded-lg border border-dashed border-gray-300 bg-transparent p-3 text-center text-sm text-gray-500 ring-0 focus-within:ring-0 hover:border-gray-500"
	>Add item</Button
>
