<script lang="ts">
	import { Check, CircleAlert, CircleX } from '@lucide/svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { Probe } from '$lib/api';
	import { cn } from '$lib/utils';

	let { probe }: { probe: Probe } = $props();
	let latency = $derived(typeof probe.ms === 'number' ? probe.ms : undefined);
	let fast = $derived(probe.ok && latency !== undefined && latency < 50);
	let reachableUnknownLatency = $derived(probe.ok && latency === undefined);
	let label = $derived(probe.ok ? (latency === undefined ? 'Reachable' : `${latency} ms`) : 'Unreachable');
</script>

{#if probe.ok}
	<Badge
		class={cn(
			'gap-1 border',
			fast || reachableUnknownLatency
				? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300'
				: 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-800 dark:bg-amber-950 dark:text-amber-300'
		)}
	>
		{#if fast || reachableUnknownLatency}
			<Check size={12} />
		{:else}
			<CircleAlert size={12} />
		{/if}
		{label}
	</Badge>
{:else}
	<Badge
		class="gap-1 border border-red-200 bg-red-50 text-red-700 dark:border-red-800 dark:bg-red-950 dark:text-red-300"
		title={probe.error}
	>
		<CircleX size={12} />
		{label}
	</Badge>
{/if}
