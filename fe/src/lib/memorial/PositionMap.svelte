<script lang="ts">
	import type { Position } from './api';

	let {
		positions,
		areaCode,
		fullscreen = false,
		onposition
	}: {
		positions: Position[];
		areaCode: string;
		fullscreen?: boolean;
		onposition: (position: Position) => void;
	} = $props();
	type HeatLevel = 'empty' | 'low' | 'medium' | 'high' | 'very-high';
	type HeatFilter = 'all' | HeatLevel;
	let heatFilter = $state<HeatFilter>('all');
	let maxRow = $derived(Math.max(0, ...positions.map((position) => position.row_number)));
	let maxColumn = $derived(Math.max(0, ...positions.map((position) => position.column_number)));
	let maxTabletCount = $derived(Math.max(0, ...positions.map((position) => position.tablet_count)));
	let rows = $derived(Array.from({ length: maxRow }, (_, index) => index + 1));
	let columns = $derived(Array.from({ length: maxColumn }, (_, index) => index + 1));
	let positionByCoordinate = $derived(
		new Map(
			positions.map((position) => [`${position.row_number}:${position.column_number}`, position])
		)
	);

	function heatLevel(position: Position): HeatLevel {
		if (position.tablet_count === 0 || maxTabletCount === 0) return 'empty';
		const ratio = position.tablet_count / maxTabletCount;
		if (ratio <= 0.25) return 'low';
		if (ratio <= 0.5) return 'medium';
		if (ratio <= 0.75) return 'high';
		return 'very-high';
	}
	function heatTone(position: Position) {
		const readableText = 'text-slate-950 dark:text-slate-50';
		switch (heatLevel(position)) {
			case 'empty':
				return `${readableText} border-slate-300 bg-slate-50 dark:border-slate-700 dark:bg-slate-900/40`;
			case 'low':
				return `${readableText} border-emerald-300 bg-emerald-50 dark:border-emerald-800 dark:bg-emerald-950/40`;
			case 'medium':
				return `${readableText} border-sky-300 bg-sky-50 dark:border-sky-800 dark:bg-sky-950/40`;
			case 'high':
				return `${readableText} border-amber-300 bg-amber-50 dark:border-amber-800 dark:bg-amber-950/40`;
			case 'very-high':
				return `${readableText} border-red-300 bg-red-50 dark:border-red-800 dark:bg-red-950/40`;
		}
	}
	function matches(position: Position) {
		return heatFilter === 'all' || heatLevel(position) === heatFilter;
	}
</script>

<div
	class={[
		'flex flex-col overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface)]',
		fullscreen ? 'min-h-0 flex-1' : 'max-h-[calc(100dvh-13rem)] min-h-[420px] lg:min-h-0 lg:flex-1'
	]}
>
	<div
		class="z-10 flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-3"
	>
		<div>
			<h3 class="font-semibold">Sơ đồ Khu {areaCode}</h3>
			<p class="mt-1 text-xs text-[var(--color-text-secondary)]">
				Màu thể hiện mật độ bài vị tương đối trong khu. Click vị trí để xem chi tiết.
			</p>
		</div>
		<div class="flex flex-wrap items-center justify-end gap-3">
			<div class="flex flex-wrap gap-3 text-xs">
				{@render legend('Trống', 'bg-slate-400')}
				{@render legend('Thấp', 'bg-emerald-400')}
				{@render legend('Vừa', 'bg-sky-400')}
				{@render legend('Cao', 'bg-amber-400')}
				{@render legend('Rất cao', 'bg-red-400')}
			</div>
			<select
				bind:value={heatFilter}
				aria-label="Lọc mật độ bài vị"
				class="h-10 rounded-md border-[var(--color-border-strong)] text-sm"
			>
				<option value="all">Tất cả mật độ</option>
				<option value="empty">Vị trí trống</option>
				<option value="low">Mật độ thấp</option>
				<option value="medium">Mật độ vừa</option>
				<option value="high">Mật độ cao</option>
				<option value="very-high">Mật độ rất cao</option>
			</select>
		</div>
	</div>
	{#if positions.length === 0}<p
			class="px-4 py-16 text-center text-sm text-[var(--color-text-secondary)]"
		>
			Chưa có vị trí trong khu vực này
		</p>
	{:else}<div class="min-h-0 flex-1 overflow-auto">
			<div
				class="grid min-w-max gap-2"
				style={`grid-template-columns: 56px repeat(${maxColumn}, minmax(132px, 1fr));`}
			>
				<div
					class="sticky top-0 left-0 z-30 bg-[var(--color-surface)] shadow-[1px_1px_0_var(--color-border)]"
				></div>
				{#each columns as column (column)}<div
						class="sticky top-0 z-20 bg-[var(--color-surface)] px-2 py-2 text-center text-xs font-semibold text-[var(--color-text-secondary)] shadow-[0_1px_0_var(--color-border)]"
					>
						Cột {column}
					</div>{/each}
				{#each rows as row (row)}<div
						class="sticky left-0 z-10 grid place-items-center bg-[var(--color-surface)] text-xs font-semibold text-[var(--color-text-secondary)] shadow-[1px_0_0_var(--color-border)]"
					>
						Hàng {row}
					</div>
					{#each columns as column (`${row}:${column}`)}{@const position = positionByCoordinate.get(
							`${row}:${column}`
						)}{#if position}<button
								type="button"
								onclick={() => onposition(position)}
								disabled={!matches(position)}
								class={[
									'min-h-24 rounded-lg border p-3 text-left transition enabled:hover:-translate-y-0.5 enabled:hover:shadow-md disabled:cursor-default disabled:opacity-15',
									heatTone(position)
								]}
								><strong class="block text-sm">{position.name}</strong><span
									class="mt-2 block text-xs font-semibold">{position.tablet_count} bài vị</span
								><span class="mt-1 block text-xs opacity-75"
									>{position.spirit_count} Hương linh</span
								></button
							>{:else}<div
								class="min-h-24 rounded-lg border border-dashed border-[var(--color-border)] bg-[var(--color-surface-muted)]/40"
							></div>{/if}{/each}{/each}
			</div>
		</div>{/if}
</div>

{#snippet legend(label: string, tone: string)}<span class="inline-flex items-center gap-1.5"
		><span class={`h-2.5 w-2.5 rounded-full ${tone}`}></span>{label}</span
	>{/snippet}
