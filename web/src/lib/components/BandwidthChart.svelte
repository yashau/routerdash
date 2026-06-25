<script lang="ts">
	import { rate } from '$lib/format';

	let { points }: { points: { at: number; rx: number; tx: number }[] } = $props();
	let width = 560;
	let height = 180;
	let max = $derived(Math.max(1, ...points.flatMap((point) => [point.rx, point.tx])));

	function pathFor(key: 'rx' | 'tx') {
		if (points.length < 2) return '';
		return points
			.map((point, index) => {
				const x = (index / (points.length - 1)) * width;
				const y = height - (point[key] / max) * (height - 12) - 6;
				return `${index === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`;
			})
			.join(' ');
	}
</script>

<div class="bg-card rounded-lg border p-3">
	<div class="mb-2 flex items-center justify-between gap-3 text-sm">
		<span class="font-medium">Bandwidth</span>
		<span class="text-muted-foreground">
			{points.length ? `${rate(points.at(-1)?.rx ?? 0)} down / ${rate(points.at(-1)?.tx ?? 0)} up` : 'Collecting'}
		</span>
	</div>
	<svg viewBox={`0 0 ${width} ${height}`} class="h-44 w-full">
		<path d={pathFor('rx')} fill="none" stroke="oklch(0.54 0.14 154)" stroke-width="3" />
		<path d={pathFor('tx')} fill="none" stroke="oklch(0.55 0.16 235)" stroke-width="3" />
	</svg>
</div>
