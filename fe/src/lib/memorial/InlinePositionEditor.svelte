<script lang="ts">
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import {
		emptyPositionRow,
		parsePositionSheet,
		type EditablePositionRow
	} from './position-sheet-parser';

	let {
		items = $bindable(),
		areaCode
	}: {
		items: EditablePositionRow[];
		areaCode: string;
	} = $props();
	let sheetPaste = $state('');

	function addRow() {
		items = [...items, emptyPositionRow()];
	}
	function removeRow(index: number) {
		items = items.filter((_, itemIndex) => itemIndex !== index);
		if (items.length === 0) addRow();
	}
	function importSheet(raw: string) {
		try {
			const imported = parsePositionSheet(raw);
			const current = items.filter((row) => Object.values(row).some((value) => value.trim()));
			items = [...current, ...imported];
			sheetPaste = '';
			toastStore.success(`Đã nhập ${imported.length} vị trí`);
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể nhập dữ liệu');
		}
	}
	function displayName(position: EditablePositionRow) {
		return position.row_number && position.column_number
			? `${position.row_number}${areaCode}-${position.column_number}`
			: 'Vị trí mới';
	}
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<div class="shrink-0 border-b border-[var(--color-border)] pb-3">
		<div class="flex items-center gap-2">
			<label class="min-w-0 flex-1"
				><span class="mb-1 block text-xs font-medium">Dán từ Excel / Google Sheets</span><textarea
					bind:value={sheetPaste}
					onpaste={(event) => {
						event.preventDefault();
						importSheet(event.clipboardData?.getData('text') ?? '');
					}}
					rows="2"
					placeholder="Hàng | Cột | Ghi chú"
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
	</div>
	<div class="min-h-0 flex-1 overflow-y-auto pt-3 pr-1">
		<div class="space-y-2">
			{#each items as position, index (index)}{@const required =
					index === 0 || Object.values(position).some((value) => value.trim())}
				<div
					class="rounded-md border border-[var(--color-border)] bg-[var(--color-surface-muted)] p-2"
				>
					<div class="mb-2 flex items-center justify-between">
						<span class="text-xs font-semibold">#{index + 1} · {displayName(position)}</span>
						<button
							type="button"
							onclick={() => removeRow(index)}
							aria-label={`Xoá dòng ${index + 1}`}
							class="grid h-7 w-7 place-items-center text-[var(--color-danger)]"
							><span class="icon-[lucide--trash-2] h-3.5 w-3.5"></span></button
						>
					</div>
					<div class="grid gap-2 sm:grid-cols-3">
						{@render field('Hàng *', position, 'row_number', required, 'number')}
						{@render field('Cột *', position, 'column_number', required, 'number')}
						{@render field('Ghi chú', position, 'notes')}
					</div>
				</div>{/each}
		</div>
	</div>
</div>

{#snippet field(
	label: string,
	position: EditablePositionRow,
	key: keyof EditablePositionRow,
	required = false,
	type = 'text'
)}<label class="relative block">
		<input
			value={position[key]}
			oninput={(event) => (position[key] = event.currentTarget.value)}
			placeholder=" "
			{type}
			min={type === 'number' ? 1 : undefined}
			{required}
			class="inline-position-input peer h-10 w-full rounded-md border-[var(--color-border-strong)] px-2 pt-2 text-xs"
		/><span
			class="inline-position-label pointer-events-none absolute top-0 left-2 -translate-y-1/2 rounded-sm bg-[var(--color-surface-muted)] px-1 text-[10px] leading-none text-[var(--color-text-secondary)] transition-all peer-placeholder-shown:top-1/2 peer-placeholder-shown:text-xs peer-focus:top-0 peer-focus:text-[10px] peer-focus:text-[var(--color-primary-dark)]"
			>{label}</span
		>
	</label>{/snippet}

<style>
	.inline-position-input:placeholder-shown:not(:focus) + .inline-position-label {
		background-color: transparent;
	}
</style>
