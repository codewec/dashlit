<script lang="ts">
	import { Toaster } from 'svelte-5-french-toast';
	import '../app.css';
	let { children } = $props();
</script>

<Toaster />
{@render children()}
