<script lang="ts">
	import { tick } from 'svelte';
	import { tooltip } from '$lib/actions/tooltip';
	import CardBaiVi from './CardBaiVi.svelte';
	import type { Position } from './api';

	let {
		positions,
		areaCode,
		fullscreen = false,
		onposition,
		onemptyposition,
		onfillcoordinates
	}: {
		positions: Position[];
		areaCode: string;
		fullscreen?: boolean;
		onposition: (position: Position) => void;
		onemptyposition?: (coordinate: { rowNumber: number; columnNumber: number }) => void;
		onfillcoordinates?: (bounds: { maxRow: number; maxColumn: number }) => void;
	} = $props();

	type HeatLevel = 'empty' | 'low' | 'medium' | 'high' | 'very-high';
	type HeatFilter = 'all' | HeatLevel;
	type TabletStatus = 'pending' | 'enshrined' | 'moving' | 'moved' | 'archived';
	type TabletStatusFilter = 'all' | TabletStatus;

	const minZoom = 0.2;
	const maxZoom = 1.0;
	const zoomStep = 0.05;
	const baseLabelWidth = 56;
	const baseCellWidth = 70;

	let heatFilter = $state<HeatFilter>('all');
	let tabletStatusFilter = $state<TabletStatusFilter>('all');
	let zoom = $state(1);
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
	let maxSpiritCount = $derived(Math.max(0, ...positions.map((position) => position.spirit_count)));
	let rows = $derived(Array.from({ length: maxRow }, (_, index) => index + 1));
	let columns = $derived(Array.from({ length: maxColumn }, (_, index) => index + 1));
	let positionByCoordinate = $derived(
		new Map(
			positions.map((position) => [`${position.row_number}:${position.column_number}`, position])
		)
	);
	let labelWidth = $derived(Math.max(42, Math.round(baseLabelWidth * zoom)));
	let cellWidth = $derived(Math.round(baseCellWidth * zoom));
	let cellHeight = $derived(Math.round(100 * zoom));
	let gridGap = $derived(Math.max(3, Math.round(12 * zoom)));
	let headerHeight = $derived(Math.max(28, Math.round(40 * zoom)));
	let cellPadding = $derived(Math.max(6, Math.round(12 * zoom)));
	let cellRadius = $derived(Math.max(3, Math.round(6 * zoom)));
	let axisPadding = $derived(Math.max(8, Math.round(8 * zoom)));
	let titleFontSize = $derived(Math.max(9, Math.round(14 * zoom)));
	let metricFontSize = $derived(Math.max(8, Math.round(12 * zoom)));
	let spiritNameFontSize = $derived(Math.max(7, Math.round(11 * zoom)));
	let headerFontSize = $derived(Math.max(10, Math.round(12 * zoom)));
	let titleGap = $derived(Math.max(4, Math.round(8 * zoom)));
	let metricGap = $derived(Math.max(2, Math.round(4 * zoom)));
	let showPositionLabel = $derived(zoom >= 0.3);
	let showSummary = $derived(zoom >= 0.45);
	let showSpiritNames = $derived(zoom >= 0.5);
	let zoomPercent = $derived(`${Math.round(zoom * 100)}%`);
	let dragCursor = $derived(panning ? 'cursor-grabbing' : 'cursor-grab');

	function clampZoom(value: number) {
		return Math.min(maxZoom, Math.max(minZoom, Number(value.toFixed(2))));
	}

	function heatLevel(position: Position): HeatLevel {
		if (position.spirit_count === 0 || maxSpiritCount === 0) return 'empty';
		const ratio = position.spirit_count / maxSpiritCount;
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
		const matchesHeat = heatFilter === 'all' || heatLevel(position) === heatFilter;
		const matchesStatus =
			tabletStatusFilter === 'all' || position.tablet_statuses.includes(tabletStatusFilter);
		return matchesHeat && matchesStatus;
	}

	function tabletStatusLabel(status: TabletStatus) {
		return {
			pending: 'Chưa an vị',
			enshrined: 'Đã an vị',
			moving: 'Đang di dời',
			moved: 'Đã di dời / lưu kho',
			archived: 'Lưu trữ'
		}[status];
	}

	function positionStatusSummary(position: Position) {
		return position.tablet_statuses.map(tabletStatusLabel).join(' · ');
	}

	function positionStatusBadge(position: Position) {
		return position.tablet_statuses.length === 1
			? tabletStatusLabel(position.tablet_statuses[0])
			: `${position.tablet_statuses.length} trạng thái`;
	}

	function compactSpiritName(name: string) {
		const words = name.trim().split(/\s+/).filter(Boolean);
		return words.length > 3 ? `${words.slice(0, 3).join(' ')}…` : words.join(' ');
	}

	function lifeYears(position: Position) {
		const birth = position.single_spirit_birth_year || '?';
		const death = position.single_spirit_death_year || '?';
		return `${birth}–${death}`;
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
		fullscreen ? 'min-h-0 flex-1' : 'min-h-[620px] lg:min-h-[680px]'
	]}
>
	<div
		class="z-10 flex shrink-0 flex-wrap items-center justify-between gap-3 border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-3"
	>
		<div>
			<h3 class="font-semibold">Sơ đồ Khu {areaCode}</h3>
			<p class="mt-1 text-xs text-[var(--color-text-secondary)]">
				Màu thể hiện mật độ Hương linh tương đối trong khu. Kéo để di chuyển, Ctrl/Cmd + lăn chuột để
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
			aria-label="Lọc mật độ Hương linh"
				class="h-10 rounded-md border-[var(--color-border-strong)] text-sm"
			>
				<option value="all">Tất cả mật độ</option>
				<option value="empty">Vị trí trống</option>
				<option value="low">Mật độ thấp</option>
				<option value="medium">Mật độ vừa</option>
				<option value="high">Mật độ cao</option>
				<option value="very-high">Mật độ rất cao</option>
			</select>
			<select
				bind:value={tabletStatusFilter}
				aria-label="Lọc trạng thái Bài vị"
				class="h-10 rounded-md border-[var(--color-border-strong)] text-sm"
			>
				<option value="all">Tất cả trạng thái</option>
				<option value="pending">Chưa an vị</option>
				<option value="enshrined">Đã an vị</option>
				<option value="moving">Đang di dời</option>
				<option value="moved">Đã di dời / lưu kho</option>
				<option value="archived">Lưu trữ</option>
			</select>
			{#if onfillcoordinates && maxRow > 0 && maxColumn > 0}<button
				type="button"
				onclick={() => onfillcoordinates({ maxRow, maxColumn })}
				class="h-10 rounded-md border border-[var(--color-primary)] bg-[var(--color-primary-soft)] px-3 text-sm font-semibold text-[var(--color-primary-dark)]"
			><span class="mr-1 icon-[lucide--grid-3x3] inline-block h-4 w-4 align-text-bottom"></span>Tạo mã cho ô trống</button>{/if}
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
				style={`grid-template-columns: ${labelWidth}px repeat(${maxColumn}, ${cellWidth}px); grid-template-rows: ${headerHeight}px repeat(${maxRow}, ${cellHeight}px); gap: ${gridGap}px;`}
			>
				<div
					class="sticky top-0 left-0 z-30 bg-[var(--color-surface)] shadow-[1px_1px_0_var(--color-border)]"
					style={`height: ${headerHeight}px;`}
				></div>
				{#each columns as column (column)}
					<div
						class="sticky top-0 z-20 bg-[var(--color-surface)] text-center font-semibold text-[var(--color-text-secondary)] shadow-[0_1px_0_var(--color-border)]"
						style={`height: ${headerHeight}px; padding: ${axisPadding}px; font-size: ${headerFontSize}px; line-height: 1;`}
						aria-label={`Cột ${column}`}
						title={`Cột ${column}`}
					>
						{column}
					</div>
				{/each}
				{#each rows as row (row)}
					<div
						class="sticky left-0 z-10 grid place-items-center bg-[var(--color-surface)] font-semibold text-[var(--color-text-secondary)] shadow-[1px_0_0_var(--color-border)]"
						style={`height: ${cellHeight}px; padding: ${axisPadding}px; font-size: ${headerFontSize}px; line-height: 1;`}
						aria-label={`Hàng ${row}`}
						title={`Hàng ${row}`}
					>
						{row}
					</div>
					{#each columns as column (`${row}:${column}`)}
						{@const position = positionByCoordinate.get(`${row}:${column}`)}
						{#if position}
							{@const spiritNames = position.spirit_names ?? []}
							{@const visibleSpiritNames = spiritNames.slice(0, 3)}
							<button
								type="button"
								use:tooltip={spiritNames.join('\n')}
								onclick={() => onposition(position)}
								disabled={!matches(position)}
								class={[
									'flex w-full min-w-0 flex-col overflow-hidden text-left transition enabled:hover:-translate-y-0.5 enabled:hover:shadow-md disabled:cursor-default disabled:opacity-15',
									heatTone(position)
								]}
								style={`height: ${cellHeight}px; border-radius: ${cellRadius}px;`}
							>
								{#if position.tablet_count > 0}<CardBaiVi
									code={position.name}
									spiritCount={position.spirit_count}
									spiritNames={spiritNames}
									birthYear={position.single_spirit_birth_year}
									deathYear={position.single_spirit_death_year}
									status={position.tablet_statuses[0] ?? 'enshrined'}
									tabletType={position.tablet_types[0] ?? 'spirit'}
									fontSize={metricFontSize}
									maxNameWords={4}
								/>{:else if showPositionLabel}
									<strong class="m-2 text-left" style={`font-size: ${titleFontSize}px;`}>{position.name}</strong>
									{#if showSummary}<span class="m-2 mt-0 text-left leading-tight font-semibold opacity-80" style={`font-size: ${metricFontSize}px;`}>+ Tạo bài vị</span>{/if}
								{:else if showSummary}
									<span
										class="grid h-full place-items-center leading-tight font-semibold opacity-80"
										style={`font-size: ${metricFontSize}px;`}
									>
										+ Tạo bài vị
									</span>
								{/if}
							</button>
						{:else if onemptyposition}
							<button
								type="button"
								onclick={() => onemptyposition({ rowNumber: row, columnNumber: column })}
								class="grid w-full place-items-center rounded-[inherit] bg-[color-mix(in_srgb,var(--color-surface)_94%,var(--color-primary)_6%)] text-center font-semibold text-[var(--color-primary-dark)] transition hover:bg-[var(--color-primary-soft)]"
								style={`height: ${cellHeight}px; font-size: ${metricFontSize}px;`}
								aria-label={`Thêm vị trí ${column}${areaCode}-${row}`}
								title={`Thêm vị trí ${column}${areaCode}-${row}`}
							>
								{#if zoom >= 0.5}<span>+ Thêm vị trí</span>{/if}
							</button>
						{:else}
							<div
								class="bg-[color-mix(in_srgb,var(--color-surface)_94%,var(--color-primary)_6%)]"
								style={`height: ${cellHeight}px; border-radius: ${cellRadius}px;`}
							></div>
						{/if}
					{/each}
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.position-tablet {
		border: 1px solid #6f4b3b;
		box-shadow:
			inset 0 0 0 1px color-mix(in srgb, #ead1ad 65%, transparent),
			inset 0 0 0 2px color-mix(in srgb, #6f4b3b 45%, transparent),
			0 1px 2px rgb(44 28 20 / 18%);
	}
</style>

{#snippet legend(label: string, tone: string)}
	<span class="inline-flex items-center gap-1.5">
		<span class={`h-2.5 w-2.5 rounded-full ${tone}`}></span>{label}
	</span>
{/snippet}
