<script lang="ts">
	import { onMount } from 'svelte';
	import CodeBlock from '$lib/components/CodeBlock.svelte';
	import StatusBadge from '$lib/components/StatusBadge.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { getJSON, type FRRStatus } from '$lib/api';

	let frr: FRRStatus | undefined = $state();
	let selected = $state<'config' | 'ospf' | 'bgp'>('config');

	let choices = $derived([
		{ value: 'config' as const, label: 'Running Config', status: frr?.runningConfig },
		...(!frr || frr.ospf.available ? [{ value: 'ospf' as const, label: 'OSPF', status: frr?.ospf }] : []),
		...(!frr || frr.bgp.available ? [{ value: 'bgp' as const, label: 'BGP', status: frr?.bgp }] : [])
	]);

	let output = $derived(
		selected === 'ospf' ? frr?.ospf.value : selected === 'bgp' ? frr?.bgp.value : frr?.runningConfig.value
	);

	onMount(() => {
		const controller = new AbortController();
		void getJSON<FRRStatus>('/api/frr', controller.signal).then((value) => {
			frr = value;
			if (selected === 'ospf' && !value.ospf.available) selected = 'config';
			if (selected === 'bgp' && !value.bgp.available) selected = 'config';
		});
		return () => controller.abort();
	});
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<Card.Title>FRR</Card.Title>
		<Card.Description>OSPF/BGP summaries and formatted running configuration.</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="bg-muted inline-flex h-8 w-fit items-center rounded-lg p-[3px]">
			{#each choices as choice}
				<Button
					type="button"
					size="sm"
					variant={selected === choice.value ? 'default' : 'ghost'}
					class="h-6 active:translate-y-0"
					onclick={() => (selected = choice.value)}
				>
					{choice.label}
					<StatusBadge status={choice.status} />
				</Button>
			{/each}
		</div>
		<CodeBlock value={output} />
	</Card.Content>
</Card.Root>
