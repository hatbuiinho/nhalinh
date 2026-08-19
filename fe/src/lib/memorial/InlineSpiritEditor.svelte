<script lang="ts">
	import SpiritImageUploader from './SpiritImageUploader.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import { uploadSpiritImage } from '$lib/uploads/api';
	import type { EditableSpiritInput } from './api';
	import { emptyInlineSpirit, parseSpiritSheet } from './sheet-parser';

	let {
		items = $bindable(),
		allowExistingUploads = false,
		requireFirst = true,
		readOnly = false,
		onbusychange = () => {}
	}: {
		items: EditableSpiritInput[];
		allowExistingUploads?: boolean;
		requireFirst?: boolean;
		readOnly?: boolean;
		onbusychange?: (busy: boolean) => void;
	} = $props();
	let sheetPaste = $state('');
	let uploadingIndex = $state<number | null>(null);

	function addRow() {
		items = [...items, emptyInlineSpirit()];
	}
	function removeRow(index: number) {
		items = items.filter((_, itemIndex) => itemIndex !== index);
		if (items.length === 0) addRow();
	}
	function importSheet(raw: string) {
		try {
			const imported = parseSpiritSheet(raw);
			const current = items.filter((row) => Object.values(row).some((value) => value.trim()));
			items = [...current, ...imported];
			sheetPaste = '';
			toastStore.success(`Đã nhập ${imported.length} Hương linh`);
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể nhập dữ liệu');
		}
	}
	async function selectImage(index: number, file: File) {
		uploadingIndex = index;
		onbusychange(true);
		try {
			items[index].image_url = await uploadSpiritImage(file);
			toastStore.success('Đã tải ảnh Hương linh');
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể tải ảnh');
		} finally {
			uploadingIndex = null;
			onbusychange(false);
		}
	}
</script>

<div class="flex min-h-0 flex-1 flex-col">
	{#if !readOnly}<div class="shrink-0 border-b border-[var(--color-border)] pb-3">
			<div class="flex items-center gap-2">
				<label class="min-w-0 flex-1"
					><span class="mb-1 block text-xs font-medium">Dán từ Excel / Google Sheets</span><textarea
						bind:value={sheetPaste}
						onpaste={(event) => {
							event.preventDefault();
							importSheet(event.clipboardData?.getData('text') ?? '');
						}}
						rows="2"
						placeholder="Tên | Pháp danh | Năm sinh | Năm mất | Tuổi | Hình URL | Nơi an táng | Người gửi | Tháng gửi | Ghi chú"
						class="w-full rounded-md border-[var(--color-border-strong)] text-xs"></textarea></label
				>
				{#if sheetPaste.trim()}<button
						type="button"
						onclick={() => importSheet(sheetPaste)}
						class="h-9 shrink-0 rounded-md bg-[var(--color-primary-soft)] px-3 text-xs font-semibold text-[var(--color-primary-dark)]"
						>Nhập dữ liệu</button
					>{/if}<button
					type="button"
					onclick={addRow}
					class="h-9 shrink-0 rounded-md border border-[var(--color-primary)] px-3 text-xs font-semibold text-[var(--color-primary-dark)]"
					>Thêm dòng</button
				>
			</div>
		</div>{/if}
	<div class="min-h-0 flex-1 overflow-y-auto pt-3 pr-1">
		<div class="space-y-2">
			{#each items as spirit, index (index)}<div
					class="rounded-md border border-[var(--color-border)] bg-[var(--color-surface-muted)] p-2"
				>
					<div class="mb-1 flex items-center justify-between">
						<span class="text-xs font-semibold"
							>#{index + 1} · {spirit.full_name || 'Hương linh mới'}</span
						>{#if !readOnly}<button
								type="button"
								onclick={() => removeRow(index)}
								aria-label={`Xoá dòng ${index + 1}`}
								class="grid h-7 w-7 place-items-center text-[var(--color-danger)]"
								><span class="icon-[lucide--trash-2] h-3.5 w-3.5"></span></button
							>{/if}
					</div>
					<div
						class={allowExistingUploads && spirit.id ? 'grid gap-2 lg:grid-cols-[auto_1fr]' : ''}
					>
						{#if allowExistingUploads && spirit.id}<SpiritImageUploader
								imageUrl={spirit.image_url}
								displayName={spirit.full_name}
								uploading={uploadingIndex === index}
								compact
								{readOnly}
								onselect={(file) => selectImage(index, file)}
							/>{/if}
						<div class="grid gap-1.5 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
							{@render field(
								'Tên *',
								spirit,
								'full_name',
								(requireFirst && index === 0) || Object.values(spirit).some((value) => value.trim())
							)}
							{@render field('Pháp danh', spirit, 'dharma_name')}
							{@render field('Năm sinh', spirit, 'birth_year')}
							{@render field('Năm mất', spirit, 'death_year')}
							{@render field('Tuổi', spirit, 'age')}
							{#if !allowExistingUploads || !spirit.id}{@render field(
									'Hình URL',
									spirit,
									'image_url'
								)}{/if}
							{@render field('Nơi an táng', spirit, 'burial_place')}
							{@render field('Người gửi', spirit, 'sender')}
							{@render field('Tháng gửi', spirit, 'sent_month')}
							{@render field('Ghi chú', spirit, 'notes')}
						</div>
					</div>
				</div>{/each}
		</div>
	</div>
</div>

{#snippet field(
	label: string,
	spirit: EditableSpiritInput,
	key: keyof Omit<EditableSpiritInput, 'id'>,
	required = false
)}<label class="relative block">
		<input
			value={spirit[key]}
			oninput={(event) => (spirit[key] = event.currentTarget.value)}
			placeholder=" "
			required={required && !readOnly}
			readonly={readOnly}
			class="inline-spirit-input peer h-9 w-full rounded-md border-[var(--color-border-strong)] px-2 pt-2 text-xs"
		/><span
			class="inline-spirit-label pointer-events-none absolute top-0 left-2 -translate-y-1/2 rounded-sm bg-[var(--color-surface-muted)] px-1 text-[10px] leading-none text-[var(--color-text-secondary)] transition-all peer-placeholder-shown:top-1/2 peer-placeholder-shown:text-xs peer-focus:top-0 peer-focus:text-[10px] peer-focus:text-[var(--color-primary-dark)]"
			>{label}</span
		>
	</label>{/snippet}

<style>
	.inline-spirit-input:placeholder-shown:not(:focus) + .inline-spirit-label {
		background-color: transparent;
	}
</style>
