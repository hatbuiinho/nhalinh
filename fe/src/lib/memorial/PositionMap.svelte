<script lang="ts">
	import { tick } from 'svelte';
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

	const minZoom = 0.2;
	const maxZoom = 1.0;
	const zoomStep = 0.05;
	const baseLabelWidth = 56;
	const baseCellWidth = 132;
	const baseCellHeight = 96;

	let heatFilter = $state<HeatFilter>('all');
	let zoom = $state(0.5);
	let viewport = $state<HTMLDivElement | null>(null);
	let panning = $state(false);
	let panPointerId = $state<number | null>(null);
	let panStartX = 0;
	let panStartY = 0;
	let panStartLeft = 0;
	let panStartTop = 0;
	let lastPointerX = $state<number | null>(null);
	let lastPointerY = $state<number | null>(null);

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
	let labelWidth = $derived(Math.max(42, Math.round(baseLabelWidth * zoom)));
	let cellWidth = $derived(Math.round(baseCellWidth * zoom));
	let cellHeight = $derived(Math.round(baseCellHeight * zoom));
	let gridGap = $derived(Math.max(2, Math.round(8 * zoom)));
	let headerHeight = $derived(Math.max(28, Math.round(40 * zoom)));
	let cellPadding = $derived(Math.max(6, Math.round(12 * zoom)));
	let cellRadius = $derived(Math.max(8, Math.round(12 * zoom)));
	let axisPadding = $derived(Math.max(8, Math.round(8 * zoom)));
	let titleFontSize = $derived(Math.max(10, Math.round(14 * zoom)));
	let metricFontSize = $derived(Math.max(9, Math.round(12 * zoom)));
	let headerFontSize = $derived(Math.max(10, Math.round(12 * zoom)));
	let titleGap = $derived(Math.max(4, Math.round(8 * zoom)));
	let metricGap = $derived(Math.max(2, Math.round(4 * zoom)));
	let zoomPercent = $derived(`${Math.round(zoom * 100)}%`);
	let dragCursor = $derived(panning ? 'cursor-grabbing' : 'cursor-grab');

	function clampZoom(value: number) {
		return Math.min(maxZoom, Math.max(minZoom, Number(value.toFixed(2))));
	}

	function heatLevel(position: Position): HeatLevel {
		if (position.tablet_count === 0 || maxTabletCount === 0) return 'empty';
		const ratio = position.tablet_count / maxTabletCount;
		if (ratio <= 0.25) return 'low';
		if (ratio <= 0.5) return 'medium';
		if (ratio <= 0.75) return 'high';
		return 'very-high';
	}

	function heatTone(position: Position) {
		const readableText = 'text-[var(--color-primary-dark)]';
		switch (heatLevel(position)) {
			case 'empty':
				return `${readableText} bg-[color-mix(in_srgb,var(--color-surface)_92%,var(--color-primary)_8%)]`;
			case 'low':
				return `${readableText} bg-[color-mix(in_srgb,var(--color-surface)_78%,var(--color-primary)_22%)]`;
			case 'medium':
				return `${readableText} bg-[color-mix(in_srgb,var(--color-surface)_62%,var(--color-primary)_38%)]`;
			case 'high':
				return `text-white bg-[color-mix(in_srgb,var(--color-surface)_42%,var(--color-primary)_58%)]`;
			case 'very-high':
				return `text-white bg-[color-mix(in_srgb,var(--color-surface)_18%,var(--color-primary)_82%)]`;
		}
	}

	function matches(position: Position) {
		return heatFilter === 'all' || heatLevel(position) === heatFilter;
	}

	function pointerAnchor() {
		if (
			!viewport ||
			lastPointerX === null ||
			lastPointerY === null ||
			lastPointerX < 0 ||
			lastPointerY < 0 ||
			lastPointerX > viewport.clientWidth ||
			lastPointerY > viewport.clientHeight
		) {
			return {
				x: viewport ? viewport.clientWidth / 2 : 0,
				y: viewport ? viewport.clientHeight / 2 : 0
			};
		}
		return { x: lastPointerX, y: lastPointerY };
	}

	async function applyZoom(nextZoom: number, anchor = pointerAnchor()) {
		const targetZoom = clampZoom(nextZoom);
		if (!viewport || targetZoom === zoom) {
			zoom = targetZoom;
			return;
		}
		const currentZoom = zoom;
		const contentX = viewport.scrollLeft + anchor.x;
		const contentY = viewport.scrollTop + anchor.y;
		zoom = targetZoom;
		await tick();
		const ratio = targetZoom / currentZoom;
		viewport.scrollLeft = Math.max(0, contentX * ratio - anchor.x);
		viewport.scrollTop = Math.max(0, contentY * ratio - anchor.y);
	}

	function zoomIn() {
		void applyZoom(zoom + zoomStep);
	}

	function zoomOut() {
		void applyZoom(zoom - zoomStep);
	}

	function resetView() {
		void applyZoom(0.5);
		if (viewport) {
			viewport.scrollLeft = 0;
			viewport.scrollTop = 0;
		}
	}

	function handleWheel(event: WheelEvent) {
		if (!event.ctrlKey && !event.metaKey) return;
		event.preventDefault();
		if (viewport) {
			const bounds = viewport.getBoundingClientRect();
			lastPointerX = event.clientX - bounds.left;
			lastPointerY = event.clientY - bounds.top;
		}
		void applyZoom(zoom + (event.deltaY < 0 ? zoomStep : -zoomStep));
	}

	function startPan(event: PointerEvent) {
		if (!viewport || event.button !== 0) return;
		const bounds = viewport.getBoundingClientRect();
		lastPointerX = event.clientX - bounds.left;
		lastPointerY = event.clientY - bounds.top;
		if (event.target instanceof HTMLElement) {
			const interactive = event.target.closest('button, select, option, input, textarea, a');
			if (interactive && interactive !== viewport) return;
		}
		panning = true;
		panPointerId = event.pointerId;
		panStartX = event.clientX;
		panStartY = event.clientY;
		panStartLeft = viewport.scrollLeft;
		panStartTop = viewport.scrollTop;
		viewport.setPointerCapture(event.pointerId);
	}

	function movePan(event: PointerEvent) {
		if (!viewport) return;
		const bounds = viewport.getBoundingClientRect();
		lastPointerX = event.clientX - bounds.left;
		lastPointerY = event.clientY - bounds.top;
		if (!panning || panPointerId !== event.pointerId) return;
		viewport.scrollLeft = panStartLeft - (event.clientX - panStartX);
		viewport.scrollTop = panStartTop - (event.clientY - panStartY);
	}

	function stopPan(event?: PointerEvent) {
		if (viewport && event) {
			const bounds = viewport.getBoundingClientRect();
			lastPointerX = event.clientX - bounds.left;
			lastPointerY = event.clientY - bounds.top;
		}
		if (
			viewport &&
			event &&
			panPointerId === event.pointerId &&
			viewport.hasPointerCapture(event.pointerId)
		) {
			viewport.releasePointerCapture(event.pointerId);
		}
		panning = false;
		panPointerId = null;
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
				Màu thể hiện mật độ bài vị tương đối trong khu. Kéo để di chuyển, Ctrl/Cmd + lăn chuột để
				zoom.
			</p>
		</div>
		<div class="flex flex-wrap items-center justify-end gap-3">
			<div class="flex flex-wrap gap-3 text-xs">
				{@render legend(
					'Trống',
					'bg-[color-mix(in_srgb,var(--color-surface)_92%,var(--color-primary)_8%)]'
				)}
				{@render legend(
					'Thấp',
					'bg-[color-mix(in_srgb,var(--color-surface)_78%,var(--color-primary)_22%)]'
				)}
				{@render legend(
					'Vừa',
					'bg-[color-mix(in_srgb,var(--color-surface)_62%,var(--color-primary)_38%)]'
				)}
				{@render legend(
					'Cao',
					'bg-[color-mix(in_srgb,var(--color-surface)_42%,var(--color-primary)_58%)]'
				)}
				{@render legend(
					'Rất cao',
					'bg-[color-mix(in_srgb,var(--color-surface)_18%,var(--color-primary)_82%)]'
				)}
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
			<div
				class="flex items-center rounded-md border border-[var(--color-border)] bg-[var(--color-surface)]"
			>
				<button
					type="button"
					onclick={zoomOut}
					class="grid h-10 w-10 place-items-center text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
					aria-label="Thu nhỏ sơ đồ"
				>
					<span class="icon-[lucide--minus] h-4 w-4" aria-hidden="true"></span>
				</button>
				<span class="min-w-16 px-2 text-center text-sm font-semibold">{zoomPercent}</span>
				<button
					type="button"
					onclick={zoomIn}
					class="grid h-10 w-10 place-items-center text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
					aria-label="Phóng to sơ đồ"
				>
					<span class="icon-[lucide--plus] h-4 w-4" aria-hidden="true"></span>
				</button>
				<button
					type="button"
					onclick={resetView}
					class="border-l border-[var(--color-border)] px-3 text-sm font-medium text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
				>
					Reset
				</button>
			</div>
		</div>
	</div>
	{#if positions.length === 0}
		<p class="px-4 py-16 text-center text-sm text-[var(--color-text-secondary)]">
			Chưa có vị trí trong khu vực này
		</p>
	{:else}
		<div
			bind:this={viewport}
			role="region"
			aria-label={`Sơ đồ vị trí khu ${areaCode || ''}`}
			class={[
				'min-h-0 flex-1 overflow-auto overscroll-contain bg-[var(--color-surface-muted)]/30',
				dragCursor
			]}
			onwheel={handleWheel}
			onpointerdown={startPan}
			onpointermove={movePan}
			onpointerup={stopPan}
			onpointercancel={stopPan}
			onpointerleave={(event) => {
				if (panning) stopPan(event);
				else {
					lastPointerX = null;
					lastPointerY = null;
				}
			}}
		>
			<div
				class="grid min-w-max"
				style={`grid-template-columns: ${labelWidth}px repeat(${maxColumn}, minmax(${cellWidth}px, 1fr)); gap: ${gridGap}px;`}
			>
				<div
					class="sticky top-0 left-0 z-30 bg-[var(--color-surface)] shadow-[1px_1px_0_var(--color-border)]"
					style={`min-height: ${headerHeight}px;`}
				></div>
				{#each columns as column (column)}
					<div
						class="sticky top-0 z-20 bg-[var(--color-surface)] text-center font-semibold text-[var(--color-text-secondary)] shadow-[0_1px_0_var(--color-border)]"
						style={`min-height: ${headerHeight}px; padding: ${axisPadding}px; font-size: ${headerFontSize}px; line-height: 1;`}
						aria-label={`Cột ${column}`}
						title={`Cột ${column}`}
					>
						{column}
					</div>
				{/each}
				{#each rows as row (row)}
					<div
						class="sticky left-0 z-10 grid place-items-center bg-[var(--color-surface)] font-semibold text-[var(--color-text-secondary)] shadow-[1px_0_0_var(--color-border)]"
						style={`min-height: ${cellHeight}px; padding: ${axisPadding}px; font-size: ${headerFontSize}px; line-height: 1;`}
						aria-label={`Hàng ${row}`}
						title={`Hàng ${row}`}
					>
						{row}
					</div>
					{#each columns as column (`${row}:${column}`)}
						{@const position = positionByCoordinate.get(`${row}:${column}`)}
						{#if position}
							<button
								type="button"
								onclick={() => onposition(position)}
								disabled={!matches(position)}
								class={[
									'text-left transition enabled:hover:-translate-y-0.5 enabled:hover:shadow-md disabled:cursor-default disabled:opacity-15',
									heatTone(position)
								]}
								style={`min-height: ${cellHeight}px; padding: ${cellPadding}px; border-radius: ${cellRadius}px;`}
							>
								<strong class="block leading-tight" style={`font-size: ${titleFontSize}px;`}>
									{position.name}
								</strong>
								<span
									class="block leading-tight font-semibold"
									style={`margin-top: ${titleGap}px; font-size: ${metricFontSize}px;`}
								>
									{position.tablet_count} BV
								</span>
								<span
									class="block leading-tight opacity-75"
									style={`margin-top: ${metricGap}px; font-size: ${metricFontSize}px;`}
								>
									{position.spirit_count} HL
								</span>
							</button>
						{:else}
							<div
								class="bg-[color-mix(in_srgb,var(--color-surface)_94%,var(--color-primary)_6%)]"
								style={`min-height: ${cellHeight}px; border-radius: ${cellRadius}px;`}
							></div>
						{/if}
					{/each}
				{/each}
			</div>
		</div>
	{/if}
</div>

{#snippet legend(label: string, tone: string)}
	<span class="inline-flex items-center gap-1.5">
		<span class={`h-2.5 w-2.5 rounded-full ${tone}`}></span>{label}
	</span>
{/snippet}
