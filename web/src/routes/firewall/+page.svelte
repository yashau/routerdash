<script lang="ts">
	import { onMount } from 'svelte';
	import CodeBlock from '$lib/components/CodeBlock.svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { getJSON, type FirewallStatus } from '$lib/api';

	let status: FirewallStatus | undefined = $state();
	let activeFirewall = $derived(
		status?.nftables.available
			? status.nftables.value
			: status?.iptables.available
				? status.iptables.value
				: undefined
	);

	onMount(() => {
		const controller = new AbortController();
		void getJSON<FirewallStatus>('/api/firewall', controller.signal).then((value) => (status = value));
		return () => controller.abort();
	});
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<Card.Title>Firewall Configuration</Card.Title>
	</Card.Header>
	<Card.Content class="space-y-4">
		{#if activeFirewall}
			<CodeBlock value={activeFirewall} />
		{:else}
			<p class="text-muted-foreground text-sm">No firewall command output is available.</p>
		{/if}
	</Card.Content>
</Card.Root>
