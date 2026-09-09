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
	let matrix = $state({
		columnFrom: '1',
		columnTo: '',
		rowFrom: '1',
		rowTo: '',
		notes: ''
	});

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
	function createMatrix() {
		const columnFrom = Number(matrix.columnFrom);
		const columnTo = Number(matrix.columnTo);
		const rowFrom = Number(matrix.rowFrom);
		const rowTo = Number(matrix.rowTo);
		if (![columnFrom, columnTo, rowFrom, rowTo].every(Number.isInteger)) {
			toastStore.error('Cột và hàng phải là số nguyên lớn hơn 0');
			return;
		}
		if (columnFrom < 1 || rowFrom < 1 || columnTo < columnFrom || rowTo < rowFrom) {
			toastStore.error('Khoảng cột hoặc hàng không hợp lệ');
			return;
		}

		const total = (columnTo - columnFrom + 1) * (rowTo - rowFrom + 1);
		if (total > 500) {
			toastStore.error('Mỗi lần chỉ có thể tạo tối đa 500 vị trí');
			return;
		}

		const current = items.filter((row) => Object.values(row).some((value) => value.trim()));
		const coordinates = new Set(current.map((row) => `${row.column_number}:${row.row_number}`));
		const generated: EditablePositionRow[] = [];
		for (let column = columnFrom; column <= columnTo; column += 1) {
			for (let row = rowFrom; row <= rowTo; row += 1) {
				const coordinate = `${column}:${row}`;
				if (coordinates.has(coordinate)) continue;
				coordinates.add(coordinate);
				generated.push({
					column_number: String(column),
					row_number: String(row),
					notes: matrix.notes
				});
			}
		}
		if (current.length + generated.length > 500) {
			toastStore.error('Danh sách vị trí chỉ có thể chứa tối đa 500 dòng mỗi lần lưu');
			return;
		}
		items = [...current, ...generated];
		toastStore.success(`Đã thêm ${generated.length} vị trí vào danh sách`);
	}
	function displayName(position: EditablePositionRow) {
		return position.row_number && position.column_number
			? `${position.column_number}${areaCode}-${position.row_number}`
			: 'Vị trí mới';
	}
</script>

<div class="flex min-h-0 flex-1 flex-col">
	<div class="shrink-0 border-b border-[var(--color-border)] pb-3">
		<div
			class="mb-3 rounded-md border border-[var(--color-border)] bg-[var(--color-surface-muted)] p-3"
		>
			<div class="mb-2 flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
				<span class="text-sm font-semibold">Tạo vị trí theo ma trận</span>
				<span class="text-xs text-[var(--color-text-secondary)]"
					>Ví dụ: cột 1–10, hàng 1–5 sẽ tạo 50 vị trí</span
				>
			</div>
			<div class="grid gap-2 sm:grid-cols-5">
				<label
					><span class="mb-1 block text-xs font-medium">Cột từ *</span><input
						bind:value={matrix.columnFrom}
						type="number"
						min="1"
						class="h-9 w-full rounded-md border-[var(--color-border-strong)] text-xs"
					/></label
				>
				<label
					><span class="mb-1 block text-xs font-medium">Cột đến *</span><input
						bind:value={matrix.columnTo}
						type="number"
						min="1"
						placeholder="10"
						class="h-9 w-full rounded-md border-[var(--color-border-strong)] text-xs"
					/></label
				>
				<label
					><span class="mb-1 block text-xs font-medium">Hàng từ *</span><input
						bind:value={matrix.rowFrom}
						type="number"
						min="1"
						class="h-9 w-full rounded-md border-[var(--color-border-strong)] text-xs"
					/></label
				>
				<label
					><span class="mb-1 block text-xs font-medium">Hàng đến *</span><input
						bind:value={matrix.rowTo}
						type="number"
						min="1"
						placeholder="5"
						class="h-9 w-full rounded-md border-[var(--color-border-strong)] text-xs"
					/></label
				>
				<label
					><span class="mb-1 block text-xs font-medium">Ghi chú chung</span><input
						bind:value={matrix.notes}
						class="h-9 w-full rounded-md border-[var(--color-border-strong)] text-xs"
					/></label
				>
			</div>
			<button
				type="button"
				onclick={createMatrix}
				class="mt-3 h-9 rounded-md border border-[var(--color-primary)] px-3 text-xs font-semibold text-[var(--color-primary-dark)]"
				>Tạo các vị trí</button
			>
		</div>
		<div class="flex items-center gap-2">
			<label class="min-w-0 flex-1"
				><span class="mb-1 block text-xs font-medium">Dán từ Excel / Google Sheets</span><textarea
					bind:value={sheetPaste}
					onpaste={(event) => {
						event.preventDefault();
						importSheet(event.clipboardData?.getData('text') ?? '');
					}}
					rows="2"
					placeholder="Cột | Hàng | Ghi chú"
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
						{@render field('Cột *', position, 'column_number', required, 'number')}
						{@render field('Hàng *', position, 'row_number', required, 'number')}
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
