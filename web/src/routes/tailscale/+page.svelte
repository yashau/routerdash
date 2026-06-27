<script lang="ts">
	import { onMount } from 'svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { getJSON, type TailscaleStatus } from '$lib/api';

	let status: TailscaleStatus | undefined = $state();
	let page = $state(1);
	const pageSize = 10;

	function loadStatus(signal?: AbortSignal) {
		void getJSON<TailscaleStatus>(`/api/tailscale?page=${page}&pageSize=${pageSize}`, signal).then((value) => (status = value));
	}

	function goToPage(nextPage: number) {
		page = nextPage;
		loadStatus();
	}

	onMount(() => {
		const controller = new AbortController();
		loadStatus(controller.signal);
		return () => controller.abort();
	});
</script>

<div class="grid gap-4 lg:grid-cols-3">
	<Card.Root class="rounded-lg">
		<Card.Header>
			<div class="flex items-center justify-between">
				<Card.Title>Status</Card.Title>
				{#if status?.available}
					<span class="text-sm font-medium">{status.backendState ?? 'Running'}</span>
				{:else}
					<StatusBadge {status} />
				{/if}
			</div>
		</Card.Header>
		<Card.Content class="space-y-2 text-sm">
			<p>Accepting routes: <strong>{status?.acceptingRoutes ? 'yes' : 'no'}</strong></p>
			<p>Uptime: <strong>{status?.uptime ?? '-'}</strong></p>
			<p>IPs: <strong>{status?.selfIps?.join(', ') || '-'}</strong></p>
		</Card.Content>
	</Card.Root>
	<Card.Root class="rounded-lg lg:col-span-2">
		<Card.Header>
			<Card.Title>Advertised Routes</Card.Title>
		</Card.Header>
		<Card.Content class="space-y-2">
			{#if status?.advertisedRouteStates?.length}
				{#each status.advertisedRouteStates as route}
					<div class="flex flex-col gap-1 rounded-md border p-2 text-sm sm:flex-row sm:items-center sm:justify-between">
						<span class="break-words font-medium [overflow-wrap:anywhere]">{route.route}</span>
						<Badge
							variant={route.approved ? 'outline' : 'secondary'}
							class={route.approved ? 'border-emerald-300 text-emerald-700 dark:border-emerald-800 dark:text-emerald-300' : ''}
						>
							{route.status}
						</Badge>
					</div>
				{/each}
			{:else}
				<p class="text-sm">No advertised routes reported.</p>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

{#if status?.available}
	<Card.Root class="mt-6 rounded-lg">
		<Card.Header>
			<Card.Title>Peers</Card.Title>
			<Card.Description>Received routes are shown per peer when Tailscale reports them.</Card.Description>
		</Card.Header>
		<Card.Content>
			<div class="grid gap-3 md:hidden">
				{#each status.peers ?? [] as peer}
					<div class="rounded-lg border p-3">
						<div class="flex items-center justify-between gap-3">
							<p class="font-medium">{peer.name}</p>
							<span class="text-sm">{peer.online ? 'online' : 'offline'}</span>
						</div>
						<div class="mt-3 space-y-2 text-sm">
							<div>
								<p class="text-muted-foreground">Received routes</p>
								<p class="break-words [overflow-wrap:anywhere]">{peer.receivedRoutes?.join(', ') || '-'}</p>
							</div>
							<div>
								<p class="text-muted-foreground">Allowed IPs</p>
								<p class="break-words [overflow-wrap:anywhere]">{peer.allowedIps?.join(', ') || '-'}</p>
							</div>
						</div>
					</div>
				{/each}
			</div>
			<div class="hidden md:block">
				<Table.Root class="table-fixed">
					<Table.Header>
						<Table.Row>
							<Table.Head class="w-40">Peer</Table.Head>
							<Table.Head class="w-20">State</Table.Head>
							<Table.Head>Received Routes</Table.Head>
							<Table.Head>Allowed IPs</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each status.peers ?? [] as peer}
							<Table.Row>
								<Table.Cell class="whitespace-normal break-words [overflow-wrap:anywhere]">{peer.name}</Table.Cell>
								<Table.Cell>{peer.online ? 'online' : 'offline'}</Table.Cell>
								<Table.Cell class="whitespace-normal break-words [overflow-wrap:anywhere]">{peer.receivedRoutes?.join(', ') || '-'}</Table.Cell>
								<Table.Cell class="whitespace-normal break-words [overflow-wrap:anywhere]">{peer.allowedIps?.join(', ') || '-'}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
			{#if status.page && status.page.totalPages > 1}
				{@const peerPage = status.page}
				<div class="mt-4 flex flex-col gap-2 text-sm sm:flex-row sm:items-center sm:justify-between">
					<p class="text-muted-foreground">Page {peerPage.page} of {peerPage.totalPages} ({peerPage.total} peers)</p>
					<div class="flex gap-2">
						<Button variant="outline" disabled={!peerPage.hasPrev} onclick={() => goToPage(peerPage.page - 1)}>Previous</Button>
						<Button variant="outline" disabled={!peerPage.hasNext} onclick={() => goToPage(peerPage.page + 1)}>Next</Button>
					</div>
				</div>
			{/if}
		</Card.Content>
	</Card.Root>
{/if}
