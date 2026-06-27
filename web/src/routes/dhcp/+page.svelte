<script lang="ts">
	import { onMount } from 'svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import { getJSON, type DHCPLeases } from '$lib/api';

	let leases: DHCPLeases | undefined = $state();
	let page = $state(1);
	const pageSize = 50;

	function loadLeases(signal?: AbortSignal) {
		void getJSON<DHCPLeases>(`/api/dhcp?page=${page}&pageSize=${pageSize}`, signal).then((value) => (leases = value));
	}

	function goToPage(nextPage: number) {
		page = nextPage;
		loadLeases();
	}

	function expires(value?: string) {
		if (!value) return '-';
		if (value === 'never') return 'never';
		return new Date(value).toLocaleString();
	}

	onMount(() => {
		const controller = new AbortController();
		loadLeases(controller.signal);
		return () => controller.abort();
	});
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<div class="flex items-center justify-between gap-3">
			<div>
				<Card.Title>DHCP Leases</Card.Title>
				<Card.Description>{leases?.path ? `dnsmasq leases from ${leases.path}` : 'dnsmasq DHCP lease list.'}</Card.Description>
			</div>
			<StatusBadge status={leases} showLoading />
		</div>
	</Card.Header>
	<Card.Content class="space-y-4">
		{#if leases?.available}
			<div class="grid gap-3 md:hidden">
				{#each leases.leases ?? [] as lease}
					<div class="rounded-lg border p-3">
						<div class="flex items-center justify-between gap-3">
							<p class="font-medium">{lease.hostname || lease.ip}</p>
							<span class="text-muted-foreground text-sm">{lease.remaining || '-'}</span>
						</div>
						<div class="mt-3 grid gap-2 text-sm">
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">IP</span><span>{lease.ip}</span></div>
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">MAC</span><span class="break-all text-right">{lease.mac}</span></div>
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">Client ID</span><span class="break-all text-right">{lease.clientId || '-'}</span></div>
							<div class="flex justify-between gap-3"><span class="text-muted-foreground">Expires</span><span class="text-right">{expires(lease.expiresAt)}</span></div>
						</div>
					</div>
				{/each}
			</div>

			<div class="hidden overflow-x-auto md:block">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Hostname</Table.Head>
							<Table.Head>IP</Table.Head>
							<Table.Head>MAC</Table.Head>
							<Table.Head>Client ID</Table.Head>
							<Table.Head>Expires</Table.Head>
							<Table.Head>Remaining</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each leases.leases ?? [] as lease}
							<Table.Row>
								<Table.Cell>{lease.hostname || '-'}</Table.Cell>
								<Table.Cell>{lease.ip}</Table.Cell>
								<Table.Cell>{lease.mac}</Table.Cell>
								<Table.Cell class="max-w-64 truncate" title={lease.clientId}>{lease.clientId || '-'}</Table.Cell>
								<Table.Cell>{expires(lease.expiresAt)}</Table.Cell>
								<Table.Cell>{lease.remaining || '-'}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>

			{#if leases.leases?.length === 0}
				<p class="text-muted-foreground text-sm">No active leases reported.</p>
			{/if}

			{#if leases.page && leases.page.totalPages > 1}
				{@const leasePage = leases.page}
				<div class="flex flex-col gap-2 text-sm sm:flex-row sm:items-center sm:justify-between">
					<p class="text-muted-foreground">Page {leasePage.page} of {leasePage.totalPages} ({leasePage.total} leases)</p>
					<div class="flex gap-2">
						<Button variant="outline" disabled={!leasePage.hasPrev} onclick={() => goToPage(leasePage.page - 1)}>Previous</Button>
						<Button variant="outline" disabled={!leasePage.hasNext} onclick={() => goToPage(leasePage.page + 1)}>Next</Button>
					</div>
				</div>
			{/if}
		{:else if leases}
			<p class="text-destructive text-sm">{leases.error}</p>
		{:else}
			<p class="text-muted-foreground text-sm">Loading leases...</p>
		{/if}
	</Card.Content>
</Card.Root>
