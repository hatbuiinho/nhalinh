<script lang="ts">
	let {
		imageUrl = '',
		name,
		sizeClass = '',
		compact = false
	}: {
		imageUrl?: string;
		name: string;
		sizeClass?: string;
		compact?: boolean;
	} = $props();
	let failed = $state(false);
	let usableImage = $derived(Boolean(imageUrl) && !failed);
	let words = $derived(name.trim().split(/\s+/).filter(Boolean));
</script>

<div class={['spirit-portrait-frame', sizeClass, compact && 'tablet-portrait-compact']}>
	{#if usableImage}
		<img src={imageUrl} alt={`Ảnh bài vị ${name}`} onerror={() => (failed = true)} />
		<span class={[
			'absolute top-[6%] right-[6%] left-[6%] z-[4] truncate rounded-sm bg-black/55 px-1 py-0.5 text-center font-semibold text-white',
			compact ? 'text-[8px] text-amber-200' : 'text-[10px]'
		]} title={name}>
			{name || 'Bài vị'}
		</span>
	{:else}
		<div class="spirit-portrait-empty overflow-hidden">
			<span class={[
				'flex h-full w-full flex-col items-center justify-evenly leading-[1.05] font-semibold',
				compact ? 'text-xs text-amber-700' : 'text-xl'
			]}>
				{#each words as word}
					<span>{word}</span>
				{:else}
					<span>?</span>
				{/each}
			</span>
		</div>
	{/if}
</div>

<style>
	.tablet-portrait-compact {
		margin-bottom: 0;
		padding: 2px;
		border-color: #6f4b3b;
		background: #6f4b3b;
		box-shadow: inset 0 0 0 1px #b38c73;
	}

	.tablet-portrait-compact::before,
	.tablet-portrait-compact::after {
		display: none;
	}
</style>
