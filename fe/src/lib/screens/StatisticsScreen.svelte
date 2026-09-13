<script lang="ts">
	import { onMount } from 'svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import OccupancyStatistics from '$lib/memorial/OccupancyStatistics.svelte';
	import { listHouses, type House } from '$lib/memorial/api';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import { houseFilter } from '$lib/memorial/house-filter.svelte';

	let houses = $state<House[]>([]),
		houseId = $state(''),
		loading = $state(true);

	onMount(() => {
		const onHouseSelect = (event: Event) => { houseId = (event as CustomEvent<string>).detail || houses[0]?.id || ''; };
		window.addEventListener('memorial-house-select', onHouseSelect);
		void initialize();
		return () => window.removeEventListener('memorial-house-select', onHouseSelect);
	});
	async function initialize() {
		loading = true;
		try {
			houses = await listHouses();
			houseId = houseFilter.id || houses[0]?.id || '';
			if (!houseFilter.id) houseFilter.id = houseId;
		} catch (error) {
			toastStore.error(error instanceof Error ? error.message : 'Không thể tải thống kê');
		} finally {
			loading = false;
		}
	}
</script>

<section class="h-full overflow-y-auto px-4 py-4 md:px-6 lg:px-8">
	<div class="mx-auto max-w-[1320px]">
		<div class="mb-5 flex flex-wrap items-end justify-between gap-3">
			<div>
				<h1 class="text-lg font-semibold">Thống kê phân bổ</h1>
				<p class="mt-1 text-sm text-[var(--color-text-secondary)]">
					Tổng quan vị trí, bài vị và Hương linh theo khu vực.
				</p>
			</div>
		</div>
		{#if loading}<div class="py-20"><LoadingIndicator label="Đang tải Nhà Linh..." /></div>
		{:else if houses.length === 0}<div
				class="rounded-md border border-dashed py-16 text-center text-sm text-[var(--color-text-secondary)]"
			>
				Chưa có Nhà Linh để thống kê
			</div>
		{:else}{#key houseId}<OccupancyStatistics {houseId} />{/key}{/if}
	</div>
</section>
