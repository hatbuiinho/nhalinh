<script lang="ts">
	import { onMount } from 'svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import { getOccupancy, type Occupancy, type OccupancyArea } from './api';

	let { houseId }: { houseId: string } = $props();
	type SortKey = keyof Pick<
		OccupancyArea,
		'code' | 'position_count' | 'empty_position_count' | 'tablet_count' | 'spirit_count'
	>;
	let data = $state<Occupancy | null>(null),
		loading = $state(true),
		error = $state(''),
		sortKey = $state<SortKey>('code'),
		sortDirection = $state<'asc' | 'desc'>('asc');
	let sortedAreas = $derived.by(() => {
		const direction = sortDirection === 'asc' ? 1 : -1;
		return [...(data?.areas ?? [])].sort((left, right) => {
			const leftValue = left[sortKey];
			const rightValue = right[sortKey];
			const compared =
				typeof leftValue === 'number' && typeof rightValue === 'number'
					? leftValue - rightValue
					: String(leftValue).localeCompare(String(rightValue), 'vi', {
							numeric: true,
							sensitivity: 'base'
						});
			return compared * direction;
		});
	});

	onMount(() => void load());
	async function load() {
		loading = true;
		error = '';
		try {
			data = await getOccupancy(houseId);
		} catch (cause) {
			error = cause instanceof Error ? cause.message : 'Không thể tải thống kê';
		} finally {
			loading = false;
		}
	}
	function sort(key: SortKey) {
		if (sortKey === key) sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		else {
			sortKey = key;
			sortDirection = 'asc';
		}
	}
	function sortAria(key: SortKey): 'ascending' | 'descending' | 'none' {
		if (sortKey !== key) return 'none';
		return sortDirection === 'asc' ? 'ascending' : 'descending';
	}
</script>

{#if loading}<div class="py-20"><LoadingIndicator label="Đang tổng hợp thống kê..." /></div>
{:else if error}<div
		class="rounded-lg border border-[var(--color-danger)]/40 bg-[var(--color-surface)] px-4 py-10 text-center"
	>
		<p class="text-sm text-[var(--color-danger)]">{error}</p>
		<button
			type="button"
			onclick={() => void load()}
			class="mt-4 h-10 rounded-md border border-[var(--color-border-strong)] px-4 text-sm font-semibold"
			>Thử lại</button
		>
	</div>
{:else if data}<div class="space-y-5">
		<div class="grid grid-cols-2 gap-3 md:grid-cols-3 xl:grid-cols-6">
			{@render metric('Khu vực', data.summary.area_count, 'icon-[lucide--layout-grid]')}
			{@render metric('Vị trí', data.summary.position_count, 'icon-[lucide--map-pin]')}
			{@render metric(
				'Vị trí trống',
				data.summary.empty_position_count,
				'icon-[lucide--circle-dashed]'
			)}
			{@render metric(
				'Đang sử dụng',
				data.summary.used_position_count,
				'icon-[lucide--circle-dot]'
			)}
			{@render metric('Bài vị', data.summary.tablet_count, 'icon-[lucide--landmark]')}
			{@render metric('Hương linh', data.summary.spirit_count, 'icon-[lucide--users]')}
			{@render metric(
				'Chưa xếp vị trí',
				data.summary.unplaced_spirit_count,
				'icon-[lucide--user-round-search]',
				data.summary.unplaced_spirit_count > 0
			)}
		</div>

		<div
			class="overflow-x-auto rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)]"
		>
			<div class="border-b border-[var(--color-border)] px-4 py-3">
				<h2 class="font-semibold">Phân bổ theo khu vực</h2>
				<p class="mt-1 text-xs text-[var(--color-text-secondary)]">
					So sánh số lượng vị trí, bài vị và Hương linh giữa các khu vực.
				</p>
			</div>
			<table class="w-full min-w-[820px] text-left text-sm">
				<thead class="bg-[var(--color-surface-muted)] text-xs text-[var(--color-text-secondary)]">
					<tr>
						<th class="px-4 py-3" aria-sort={sortAria('code')}
							>{@render sortHeader('Khu vực', 'code')}</th
						>
						<th class="px-4 py-3" aria-sort={sortAria('position_count')}
							>{@render sortHeader('Vị trí', 'position_count', true)}</th
						>
						<th class="px-4 py-3" aria-sort={sortAria('empty_position_count')}
							>{@render sortHeader('Vị trí trống', 'empty_position_count', true)}</th
						>
						<th class="px-4 py-3" aria-sort={sortAria('tablet_count')}
							>{@render sortHeader('Bài vị', 'tablet_count', true)}</th
						>
						<th class="px-4 py-3" aria-sort={sortAria('spirit_count')}
							>{@render sortHeader('Hương linh', 'spirit_count', true)}</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-[var(--color-border)]">
					{#each sortedAreas as area (area.id)}<tr class="hover:bg-[var(--color-surface-muted)]">
							<td class="px-4 py-3"
								><strong>Khu {area.code}</strong>{#if area.name}<span
										class="ml-1 text-[var(--color-text-secondary)]">· {area.name}</span
									>{/if}</td
							>
							<td class="px-4 py-3 text-right">{area.position_count}</td>
							<td class="px-4 py-3 text-right">{area.empty_position_count}</td>
							<td class="px-4 py-3 text-right">{area.tablet_count}</td>
							<td class="px-4 py-3 text-right">{area.spirit_count}</td>
						</tr>{/each}
				</tbody>
			</table>
		</div>
	</div>{/if}

{#snippet metric(label: string, value: number, icon: string, alert = false)}<article
		class={[
			'rounded-lg border bg-[var(--color-surface)] p-4',
			alert ? 'border-amber-300 dark:border-amber-800' : 'border-[var(--color-border)]'
		]}
	>
		<div class="flex items-center justify-between gap-2">
			<p class="text-xs font-medium text-[var(--color-text-secondary)]">{label}</p>
			<span class={`${icon} h-4 w-4 text-[var(--color-primary)]`} aria-hidden="true"></span>
		</div>
		<p class="mt-2 text-2xl font-bold">{value}</p>
	</article>{/snippet}

{#snippet sortHeader(label: string, key: SortKey, right = false)}<button
		type="button"
		onclick={() => sort(key)}
		class={['flex w-full items-center gap-1 font-semibold', right && 'justify-end']}
		>{label}<span
			class={[
				'h-3.5 w-3.5',
				sortKey !== key
					? 'icon-[lucide--arrow-up-down] opacity-40'
					: sortDirection === 'asc'
						? 'icon-[lucide--arrow-up] text-[var(--color-primary-dark)]'
						: 'icon-[lucide--arrow-down] text-[var(--color-primary-dark)]'
			]}
		></span></button
	>{/snippet}
