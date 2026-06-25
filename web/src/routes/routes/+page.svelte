<script lang="ts">
	import { onMount } from 'svelte';
	import CodeBlock from '$lib/components/CodeBlock.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { getJSON, type RouteTables } from '$lib/api';

	let routes: RouteTables | undefined = $state();
	let page = $state(1);
	const pageSize = 50;

	function loadRoutes(signal?: AbortSignal) {
		void getJSON<RouteTables>(`/api/routes?page=${page}&pageSize=${pageSize}`, signal).then((value) => (routes = value));
	}

	function goToPage(nextPage: number) {
		page = nextPage;
		loadRoutes();
	}

	onMount(() => {
		const controller = new AbortController();
		loadRoutes(controller.signal);
		return () => controller.abort();
	});
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<div class="flex items-center justify-between">
			<div>
				<Card.Title>Routes</Card.Title>
				<Card.Description>All installed route tables from `ip route show table all`.</Card.Description>
			</div>
			<StatusBadge status={routes} />
		</div>
	</Card.Header>
	<Card.Content class="space-y-4">
		<CodeBlock value={routes?.output} />
		{#if routes?.page && routes.page.totalPages > 1}
			{@const routePage = routes.page}
			<div class="flex flex-col gap-2 text-sm sm:flex-row sm:items-center sm:justify-between">
				<p class="text-muted-foreground">Page {routePage.page} of {routePage.totalPages} ({routePage.total} routes)</p>
				<div class="flex gap-2">
					<Button variant="outline" disabled={!routePage.hasPrev} onclick={() => goToPage(routePage.page - 1)}>Previous</Button>
					<Button variant="outline" disabled={!routePage.hasNext} onclick={() => goToPage(routePage.page + 1)}>Next</Button>
				</div>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
