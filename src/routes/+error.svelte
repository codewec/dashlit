<script>
	import { page } from '$app/state';
	import { DarkMode } from 'flowbite-svelte';
</script>

<svelte:head>
	<title>{page.status} - error</title>
</svelte:head>

<DarkMode class="absolute top-4 right-4 cursor-pointer" />
<div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 transform text-center">
	<h1 class="mb-4 text-xl font-extrabold">
		{page.status} - {page.error?.message ?? 'Ошибка'}
	</h1>
	<a
		href="/"
		onclick={(e) => {
			e.preventDefault();
			window.history.back();
		}}
		class="text-blue-500 hover:underline">Go to Back</a
	>
</div>
