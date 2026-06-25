<script lang="ts">
	import { Activity, Route, Shield, Stethoscope, Waypoints } from '@lucide/svelte';
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getJSON, type FirewallStatus, type Summary, type VersionInfo } from '$lib/api';
	import routerIcon from '$lib/assets/router.svg?raw';
	import { cn } from '$lib/utils';

	let { children } = $props();
	let summary: Summary | undefined = $state();
	let firewall: FirewallStatus | undefined = $state();
	let version = $state('');

	const nav = [
		{ href: '/', label: 'Dashboard', short: 'Dash', icon: Activity },
		{ href: '/tailscale', label: 'Tailscale', short: 'TS', icon: Route },
		{ href: '/firewall', label: 'Firewall', short: 'FW', icon: Shield },
		{ href: '/routes', label: 'Routes', short: 'Rte', icon: Route },
		{ href: '/frr', label: 'FRR', short: 'FRR', icon: Waypoints },
		{ href: '/diagnostics', label: 'Diagnostics', short: 'Diag', icon: Stethoscope }
	];

	function active(pathname: string, href: string) {
		return href === '/' ? pathname === '/' : pathname.startsWith(href);
	}

	function trimHostname(value: string) {
		const max = 28;
		return value.length > max ? `${value.slice(0, max - 3)}...` : value;
	}

	let firewallLabel = $derived(firewall?.nftables.available ? 'nftables' : firewall?.iptables.available ? 'iptables' : 'Firewall');
	let firewallShort = $derived(firewall?.nftables.available ? 'NFT' : firewall?.iptables.available ? 'IPT' : 'FW');
	let pageName = $derived(
		nav.find((item) => active(page.url.pathname, item.href))?.href === '/firewall'
			? firewallLabel
			: (nav.find((item) => active(page.url.pathname, item.href))?.label ?? 'Dashboard')
	);
	let hostname = $derived(summary?.hostname || 'Router');
	let displayHostname = $derived(trimHostname(hostname));
	let title = $derived(`${displayHostname} - ${pageName}`);

	onMount(() => {
		const controller = new AbortController();
		void getJSON<Summary>('/api/summary', controller.signal).then((value) => (summary = value));
		void getJSON<FirewallStatus>('/api/firewall', controller.signal).then((value) => (firewall = value));
		void getJSON<VersionInfo>('/api/version', controller.signal).then((value) => (version = value.version));
		return () => controller.abort();
	});
</script>

<svelte:head>
	<title>{title}</title>
</svelte:head>

<div
	class="min-h-screen bg-[linear-gradient(180deg,oklch(0.98_0.015_180),oklch(0.99_0.004_106)_320px)] dark:bg-[linear-gradient(180deg,oklch(0.2_0.03_190),oklch(0.16_0.012_250)_320px)]"
>
	<header class="border-b bg-background/85 backdrop-blur">
		<div class="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-4 lg:flex-row lg:items-center lg:justify-between">
			<a href="/" class="flex items-center gap-3">
				<div
					class="[&>svg]:size-11 flex shrink-0 text-foreground"
					style="--router-accent: var(--primary); --router-foreground: var(--foreground);"
				>
					{@html routerIcon}
				</div>
				<div>
					<p class="max-w-64 truncate text-lg font-semibold" title={hostname}>{displayHostname}</p>
					<p class="text-muted-foreground text-xs">Read-only router status and diagnostics{version ? ` - ${version}` : ''}</p>
				</div>
			</a>
			<nav class="grid w-full grid-cols-4 gap-1 sm:flex sm:min-h-8 lg:w-auto">
				{#each nav as item}
					{@const Icon = item.icon}
					<a
						href={item.href}
						aria-current={active(page.url.pathname, item.href) ? 'page' : undefined}
						class={cn(
							'inline-flex h-9 min-w-0 items-center justify-center gap-1 rounded-md border border-transparent px-1.5 text-xs font-medium transition-colors outline-none select-none sm:h-8 sm:shrink-0 sm:gap-1.5 sm:px-2.5 sm:text-sm',
							'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-3',
							active(page.url.pathname, item.href)
								? 'bg-primary text-primary-foreground'
								: 'text-muted-foreground hover:bg-muted hover:text-foreground'
						)}
					>
						<Icon size={16} class="shrink-0" />
						<span class="min-w-0 truncate sm:hidden">{item.href === '/firewall' ? firewallShort : item.short}</span>
						<span class="hidden min-w-0 truncate sm:inline">{item.href === '/firewall' ? firewallLabel : item.label}</span>
					</a>
				{/each}
			</nav>
		</div>
	</header>
	<main class="mx-auto min-w-0 max-w-7xl px-4 py-6">
		{@render children()}
	</main>
</div>
