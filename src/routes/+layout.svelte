<script lang="ts">
	import { Toaster } from 'svelte-5-french-toast';
	import { page } from '$app/state';
	import '../app.css';
	let { children, data } = $props();
</script>

<Toaster />
{#if data.envOrigin !== page.url.origin}
	<div class="w-full bg-red-600 p-2 text-center text-white">
		<p>
			Invalid ORIGIN found. Please specify ORIGIN={page.url.origin}. Check the
			<a
				target="_blank"
				class="underline"
				href="https://dashlit.cwec.dev/guide/getting-started.html">Docs</a
			>.
		</p>
	</div>
{/if}
{@render children()}

<aside class="upgrade-notice" aria-label="New DashLit version available">
	<span>A new version of DashLit is available.</span>
	<a
		href="https://codewec.github.io/dashlit/guide/migration"
		target="_blank"
		rel="noreferrer">View the upgrade guide</a
	>
</aside>
