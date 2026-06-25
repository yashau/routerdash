<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import StatusBadge from './StatusBadge.svelte';
	import type { Availability } from '$lib/api';

	let {
		label,
		value,
		status,
		detail
	}: {
		label: string;
		value?: string;
		status?: Availability;
		detail?: string;
	} = $props();
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<div class="flex items-start justify-between gap-3">
			<Card.Title>{label}</Card.Title>
			<StatusBadge {status} />
		</div>
	</Card.Header>
	<Card.Content>
		<p class="text-2xl font-semibold tracking-normal">{value ?? 'Loading'}</p>
		{#if detail}
			<p class="text-muted-foreground mt-2 text-sm">{detail}</p>
		{/if}
	</Card.Content>
</Card.Root>
