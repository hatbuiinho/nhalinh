<script lang="ts">
	let {
		imageUrl = '',
		alt,
		sizeClass = '',
		showErrorAlt = false,
		bare = false,
		compact = false
	}: {
		imageUrl?: string;
		alt: string;
		sizeClass?: string;
		showErrorAlt?: boolean;
		bare?: boolean;
		compact?: boolean;
	} = $props();
	let failed = $state(false);
	let usableImage = $derived(Boolean(imageUrl) && !failed);
</script>

{#if bare}
	{#if usableImage}
		<img src={imageUrl} {alt} onerror={() => (failed = true)} class={`${sizeClass} object-cover`} />
	{/if}
{:else}
	<div class={['spirit-portrait-frame', sizeClass, compact && 'spirit-portrait-compact']}>
		{#if usableImage}
			<img src={imageUrl} {alt} onerror={() => (failed = true)} />
		{:else}<span class="spirit-portrait-empty" class:spirit-portrait-error={failed}>
			{#if failed && showErrorAlt}<span class="line-clamp-3 px-1 text-center text-[9px] leading-tight">Ảnh lỗi: {alt}</span>
			{:else}{alt.trim().slice(0, 1).toUpperCase() || '?'}{/if}
		</span>{/if}
	</div>
{/if}

<style>
	.spirit-portrait-compact {
		margin-bottom: 0;
		padding: 2px;
		border-color: #6f4b3b;
		background: #6f4b3b;
		box-shadow: inset 0 0 0 1px #b38c73;
	}

	.spirit-portrait-compact::before,
	.spirit-portrait-compact::after {
		display: none;
	}
</style>
