<script lang="ts">
	let {
		imageUrl = '',
		alt,
		sizeClass = '',
		showErrorAlt = false
	}: {
		imageUrl?: string;
		alt: string;
		sizeClass?: string;
		showErrorAlt?: boolean;
	} = $props();
	let failed = $state(false);
	let usableImage = $derived(Boolean(imageUrl) && !failed);
</script>

<div class={`spirit-portrait-frame ${sizeClass}`}>
	{#if usableImage}
		<img src={imageUrl} {alt} onerror={() => (failed = true)} />
	{:else}<span class="spirit-portrait-empty" class:spirit-portrait-error={failed}>
		{#if failed && showErrorAlt}<span class="line-clamp-3 px-1 text-center text-[9px] leading-tight">Ảnh lỗi: {alt}</span>
		{:else}{alt.trim().slice(0, 1).toUpperCase() || '?'}{/if}
	</span>{/if}
</div>
