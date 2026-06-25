<script lang="ts">
	import { onMount } from 'svelte';
	import BandwidthChart from '$lib/components/BandwidthChart.svelte';
	import ConnectivityBadge from '$lib/components/ConnectivityBadge.svelte';
	import StatCard from '$lib/components/StatCard.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { getJSON, type Metrics, type Summary } from '$lib/api';
	import { pct, rate } from '$lib/format';

	let summary: Summary | undefined = $state();
	let metrics: Metrics | undefined = $state();
	let error = $state('');
	let points: { at: number; rx: number; tx: number }[] = $state([]);

	function interfaceAddress(name: string, fallback?: string) {
		return fallback || summary?.lan.addresses.find((address) => address.interface === name)?.cidr || '';
	}

	onMount(() => {
		const controller = new AbortController();
		void getJSON<Summary>('/api/summary', controller.signal)
			.then((value) => (summary = value))
			.catch((err: Error) => (error = err.message));
		const tick = () => {
			void getJSON<Metrics>('/api/metrics', controller.signal)
				.then((value) => {
					metrics = value;
					const rx = value.interfaces.reduce((total, iface) => total + iface.rxBps, 0);
					const tx = value.interfaces.reduce((total, iface) => total + iface.txBps, 0);
					points = [...points.slice(-39), { at: Date.now(), rx, tx }];
				})
				.catch((err: Error) => (error = err.message));
		};
		tick();
		const timer = window.setInterval(tick, 2500);
		return () => {
			controller.abort();
			window.clearInterval(timer);
		};
	});
</script>

<section class="grid gap-4 lg:grid-cols-4">
	<StatCard label="Uptime" value={summary?.uptime.value} status={summary?.uptime} />
	<StatCard label="WAN IP" value={summary?.wanIp.value} status={summary?.wanIp} />
	<StatCard label="CPU" value={metrics ? pct(metrics.cpuPercent) : undefined} status={metrics} />
	<StatCard label="Memory" value={metrics ? pct(metrics.memory.usedPct) : undefined} status={metrics} />
</section>

<section class="mt-6 grid gap-4 lg:grid-cols-[2fr_1fr]">
	<BandwidthChart {points} />
	<Card.Root class="rounded-lg">
		<Card.Header>
			<Card.Title>Core Services</Card.Title>
			<Card.Description>Only the highest-signal router state is shown here.</Card.Description>
		</Card.Header>
		<Card.Content class="space-y-4">
			<div class="flex items-center justify-between">
				<span>Tailscale</span>
				{#if summary?.tailscale.available}
					<span class="text-sm font-medium">{summary.tailscale.backendState ?? 'Running'}</span>
				{:else}
					<StatusBadge status={summary?.tailscale} />
				{/if}
			</div>
			<div class="flex items-center justify-between">
				<span>Rathole</span>
				{#if summary?.rathole.available}
					<span class="text-sm font-medium">{summary.rathole.state ?? 'active'}</span>
				{:else}
					<StatusBadge status={summary?.rathole} />
				{/if}
			</div>
			{#each summary?.connectivity ?? [] as probe}
				<div class="flex items-center justify-between">
					<span>{probe.name}</span>
					<ConnectivityBadge {probe} />
				</div>
			{/each}
		</Card.Content>
	</Card.Root>
</section>

<section class="mt-6 grid gap-4 lg:grid-cols-3">
	{#each metrics?.interfaces ?? [] as iface}
		<Card.Root class="rounded-lg">
			<Card.Header>
				<Card.Title>{iface.name}</Card.Title>
				<Card.Description>{interfaceAddress(iface.name, iface.addressCidr) || iface.operState || 'Interface'}</Card.Description>
			</Card.Header>
			<Card.Content class="grid gap-2 text-sm">
				<div class="flex justify-between"><span>Down</span><span>{rate(iface.rxBps)}</span></div>
				<div class="flex justify-between"><span>Up</span><span>{rate(iface.txBps)}</span></div>
			</Card.Content>
		</Card.Root>
	{/each}
</section>

{#if error}
	<p class="text-destructive mt-4 text-sm">{error}</p>
{/if}
