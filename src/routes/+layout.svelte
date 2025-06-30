<script lang="ts">
	import { Toaster } from 'svelte-5-french-toast';
	import '../app.css';
	let { children, data } = $props();
</script>

<Toaster />
{#if data.originWarning.isWarning}
	<div class="w-full bg-red-600 p-2 text-center text-white">
		<p>
			Invalid origin found. Please specify ORIGIN={data.originWarning.currentOrigin}. Check
			<a
				target="_blank"
				class="underline"
				href="https://dashlit.cwec.dev/guide/getting-started.html">Docs</a
			>.
		</p>
	</div>
{/if}
{@render children()}
