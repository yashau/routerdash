<script lang="ts">
	import { onMount } from 'svelte';
	import CodeBlock from '$lib/components/CodeBlock.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { getJSON, type RatholeStatus } from '$lib/api';

	let status: RatholeStatus | undefined = $state();

	onMount(() => {
		const controller = new AbortController();
		void getJSON<RatholeStatus>('/api/rathole', controller.signal).then((value) => (status = value));
		return () => controller.abort();
	});
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<div class="flex items-center justify-between">
			<Card.Title>Rathole Client</Card.Title>
			{#if status?.available}
				<span class="text-sm font-medium">{status.state ?? 'active'}</span>
			{:else}
				<StatusBadge {status} />
			{/if}
		</div>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="grid gap-3">
			{#each status?.units ?? [] as unit}
				<div class="rounded-lg border p-3">
					<div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
						<p class="font-medium">{unit.name}</p>
						<span class="text-muted-foreground text-sm">{unit.active ?? 'unknown'}{unit.sub ? ` (${unit.sub})` : ''}</span>
					</div>
					{#if unit.description}
						<p class="text-muted-foreground mt-1 text-sm">{unit.description}</p>
					{/if}
				</div>
			{/each}
		</div>
		<CodeBlock value={status?.output} />
	</Card.Content>
</Card.Root>
