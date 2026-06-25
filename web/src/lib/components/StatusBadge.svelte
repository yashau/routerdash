<script lang="ts">
	import { Badge } from '$lib/components/ui/badge/index.js';
	import type { Availability } from '$lib/api';

	let {
		status,
		good = 'Available',
		bad = 'Unavailable',
		showOk = false,
		showLoading = false
	}: {
		status?: Availability;
		good?: string;
		bad?: string;
		showOk?: boolean;
		showLoading?: boolean;
	} = $props();
</script>

{#if !status && showLoading}
	<Badge variant="outline">Loading</Badge>
{:else if status?.available && showOk}
	<Badge
		class="border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-800 dark:bg-emerald-950 dark:text-emerald-300"
	>{good}</Badge>
{:else if status && !status.available}
	<Badge variant="destructive" title={status.error}>{bad}</Badge>
{/if}
