<script lang="ts">
	import { Loader2, Play } from '@lucide/svelte';
	import CodeBlock from '$lib/components/CodeBlock.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { postDiagnostic, type DiagnosticResult } from '$lib/api';

	let tool = $state('ping');
	let target = $state('1.1.1.1');
	let loading = $state(false);
	let result: DiagnosticResult | undefined = $state();
	let requestError = $state('');
	let output = $derived(result ? (result.output ?? '').trim() || 'No output returned.' : undefined);

	async function run() {
		loading = true;
		result = undefined;
		requestError = '';
		try {
			result = await postDiagnostic(tool, target);
		} catch (error) {
			requestError = error instanceof Error ? error.message : 'Diagnostic request failed.';
		} finally {
			loading = false;
		}
	}
</script>

<Card.Root class="rounded-lg">
	<Card.Header>
		<Card.Title>Diagnostics</Card.Title>
		<Card.Description>Run bounded ping or mtr checks from the router.</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="bg-muted inline-flex h-8 w-fit items-center rounded-lg p-[3px]">
			<Button type="button" size="sm" variant={tool === 'ping' ? 'default' : 'ghost'} class="h-6 active:translate-y-0" onclick={() => (tool = 'ping')}>
				Ping
			</Button>
			<Button type="button" size="sm" variant={tool === 'mtr' ? 'default' : 'ghost'} class="h-6 active:translate-y-0" onclick={() => (tool = 'mtr')}>
				MTR
			</Button>
		</div>
		<form class="flex flex-col gap-3 sm:flex-row" onsubmit={(event) => { event.preventDefault(); void run(); }}>
			<Input bind:value={target} placeholder="Hostname or IP" />
			<Button type="submit" disabled={loading}>
				{#if loading}
					<Loader2 class="animate-spin" size={16} />
				{:else}
					<Play size={16} />
				{/if}
				Run
			</Button>
		</form>
		{#if requestError || (result && !result.available)}
			<p class="text-destructive text-sm">{requestError || result?.error || 'Diagnostic failed.'}</p>
		{/if}
		<CodeBlock value={output} placeholder={loading ? 'Waiting for output...' : 'Run a diagnostic to see output.'} />
	</Card.Content>
</Card.Root>
