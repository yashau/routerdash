<script lang="ts">
	import { onMount } from 'svelte';
	import './layout.css';
	import routerIcon from '$lib/assets/router.svg';
	import Shell from '$lib/components/Shell.svelte';
	import { applySystemTheme } from '$lib/theme';

	let { children } = $props();

	onMount(() => applySystemTheme());
</script>

<svelte:head>
	<link rel="icon" href="/favicon.ico?v=router-colors" sizes="any" />
	<link rel="icon" href={routerIcon} type="image/svg+xml" />
</svelte:head>
<Shell>
	{@render children()}
</Shell>
