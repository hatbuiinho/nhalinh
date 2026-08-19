<script lang="ts">
	import AvatarCropDialog from '$lib/ui/AvatarCropDialog.svelte';
	import Lightbox from '$lib/ui/Lightbox.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	let {
		imageUrl = '',
		displayName,
		uploading = false,
		compact = false,
		readOnly = false,
		onselect
	}: {
		imageUrl?: string;
		displayName: string;
		uploading?: boolean;
		compact?: boolean;
		readOnly?: boolean;
		onselect: (file: File) => void | Promise<void>;
	} = $props();
	let input = $state<HTMLInputElement>(),
		cropOpen = $state(false),
		cropFile = $state<File | null>(null),
		lightboxOpen = $state(false);
	function choose(event: Event) {
		const target = event.currentTarget as HTMLInputElement;
		const file = target.files?.[0];
		target.value = '';
		if (!file) return;
		if (!['image/jpeg', 'image/png', 'image/webp'].includes(file.type)) {
			toastStore.error('Chỉ hỗ trợ ảnh JPEG, PNG hoặc WebP');
			return;
		}
		if (file.size > 15 * 1024 * 1024) {
			toastStore.error('Ảnh gốc không được lớn hơn 15 MB');
			return;
		}
		cropFile = file;
		cropOpen = true;
	}
	async function cropped(file: File) {
		await onselect(file);
		cropOpen = false;
		cropFile = null;
	}
</script>

<div class={compact ? 'flex items-center gap-2' : 'flex items-start gap-4'}>
	<button
		type="button"
		onclick={() => imageUrl && (lightboxOpen = true)}
		class={[
			'relative grid aspect-[3/4] shrink-0 place-items-center overflow-hidden rounded-md border border-[var(--color-border-strong)] bg-[var(--color-primary-soft)] font-semibold text-[var(--color-primary-dark)]',
			compact ? 'w-12 text-sm' : 'w-28 text-2xl'
		]}
		>{#if imageUrl}<img
				src={imageUrl}
				alt={displayName}
				class="h-full w-full object-cover"
			/>{:else}{displayName.trim().slice(0, 1).toUpperCase() || '?'}{/if}{#if uploading}<span
				class="absolute inset-0 grid place-items-center bg-black/45 text-white"
				><span class="icon-[lucide--loader-circle] h-6 w-6 animate-spin"></span></span
			>{/if}</button
	>
	{#if !readOnly}<div class={compact ? 'shrink-0' : ''}>
			{#if !compact}<p class="text-sm font-semibold">Ảnh Hương linh</p>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
					Ảnh được cắt theo tỉ lệ chân dung 3:4.
				</p>{/if}
			<button
				type="button"
				disabled={uploading}
				onclick={() => input?.click()}
				class={[
					'rounded-md border border-[var(--color-primary)] font-semibold text-[var(--color-primary-dark)] disabled:opacity-50',
					compact ? 'h-8 px-2 text-xs' : 'mt-3 h-10 px-3 text-sm'
				]}
				><span class="mr-1 icon-[lucide--image-up] inline-block h-4 w-4 align-text-bottom"
				></span>{compact ? 'Ảnh' : 'Chọn ảnh'}</button
			>
		</div>{/if}
	<input
		bind:this={input}
		type="file"
		accept="image/jpeg,image/png,image/webp"
		class="hidden"
		onchange={choose}
	/>
</div>
<Lightbox src={imageUrl} alt={displayName} bind:open={lightboxOpen} />
<AvatarCropDialog
	open={cropOpen}
	file={cropFile}
	busy={uploading}
	aspectRatio={3 / 4}
	outputWidth={900}
	outputHeight={1200}
	title="Cắt ảnh Hương linh"
	description="Ảnh sẽ được lưu theo tỉ lệ chân dung 3:4"
	onclose={() => {
		if (!uploading) {
			cropOpen = false;
			cropFile = null;
		}
	}}
	oncrop={cropped}
/>
