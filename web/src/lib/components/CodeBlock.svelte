<script lang="ts">
	let { value, placeholder = 'Waiting for output...' }: { value?: string; placeholder?: string } = $props();
</script>

<pre
	class="bg-muted/70 text-foreground min-h-40 overflow-auto rounded-lg border p-4 text-xs leading-relaxed whitespace-pre-wrap shadow-inner dark:bg-black/35"
>{value || placeholder}</pre>
