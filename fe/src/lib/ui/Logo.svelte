<script lang="ts">
	let { compact = false }: { compact?: boolean } = $props();
</script>

<div class="inline-flex items-center gap-2 text-[var(--color-text)]" aria-label="Quản lý Nhà Linh">
	<img src="/icons/icon.svg" alt="" class="h-9 w-9 shrink-0" />
	{#if !compact}
		<span class="text-base font-semibold">Nhà Linh</span>
	{/if}
</div>
