<script lang="ts">
	import { onMount } from 'svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import OccupancyStatistics from '$lib/memorial/OccupancyStatistics.svelte';
	import { listHouses, type House } from '$lib/memorial/api';
	import { toastStore } from '$lib/ui/toast-store.svelte';

	let houses = $state<House[]>([]),
		houseId = $state(''),
		loading = $state(true);

	onMount(() => void initialize());
	async function initialize() {
		loading = true;
		try {
			houses = await listHouses();
			houseId = houses[0]?.id ?? '';
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
			{#if houses.length > 1}<label class="block min-w-56">
					<span class="mb-1 block text-xs font-medium text-[var(--color-text-secondary)]"
						>Nhà Linh</span
					>
					<select
						bind:value={houseId}
						class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
					>
						{#each houses as house (house.id)}<option value={house.id}>{house.name}</option>{/each}
					</select>
				</label>{/if}
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
