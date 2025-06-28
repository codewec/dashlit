<script lang="ts">
	import Button from 'flowbite-svelte/Button.svelte';
	const {
		id,
		handleHover,
		handleClick
	}: {
		id: string;
		handleHover: (id: string | undefined) => void;
		handleClick: () => void;
	} = $props();
</script>

<Button
	onmouseenter={() => {
		handleHover(id);
	}}
	onmouseleave={() => {
		handleHover(undefined);
	}}
	onclick={handleClick}
	color="light"
	class="cursor-pointer rounded-lg border border-dashed border-gray-300 bg-transparent p-3 text-center text-sm text-gray-500 ring-0 focus-within:ring-0 hover:border-gray-500"
	>Add item</Button
>
