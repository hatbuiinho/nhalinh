<script lang="ts">
	import { onMount } from 'svelte';
	import { searchUnplacedSpirits, type Spirit } from './api';
	import { toastStore } from '$lib/ui/toast-store.svelte';

	let {
		houseId,
		selected = $bindable()
	}: {
		houseId: string;
		selected: Spirit[];
	} = $props();

	let root: HTMLDivElement;
	let query = $state('');
	let results = $state<Spirit[]>([]);
	let loading = $state(false);
	let open = $state(false);
	let debounceTimer: ReturnType<typeof setTimeout> | undefined;
	let requestVersion = 0;
	let availableResults = $derived(
		results.filter((result) => !selected.some((spirit) => spirit.id === result.id))
	);

	onMount(() => {
		function dismiss(event: PointerEvent) {
			if (root && !root.contains(event.target as Node)) open = false;
		}
		document.addEventListener('pointerdown', dismiss);
		return () => {
			document.removeEventListener('pointerdown', dismiss);
			if (debounceTimer) clearTimeout(debounceTimer);
		};
	});

	function queueSearch() {
		if (debounceTimer) clearTimeout(debounceTimer);
		const value = query.trim();
		if (!value) {
			requestVersion += 1;
			results = [];
			loading = false;
			open = false;
			return;
		}
		open = true;
		debounceTimer = setTimeout(() => void search(value), 250);
	}

	async function search(value: string) {
		const version = ++requestVersion;
		loading = true;
		try {
			const items = await searchUnplacedSpirits(houseId, value);
			if (version === requestVersion) results = items;
		} catch (error) {
			if (version === requestVersion) {
				toastStore.error(error instanceof Error ? error.message : 'Không thể tìm Hương linh');
				open = false;
			}
		} finally {
			if (version === requestVersion) loading = false;
		}
	}

	function selectSpirit(spirit: Spirit) {
		requestVersion += 1;
		selected = [...selected, spirit];
		query = '';
		results = [];
		open = false;
	}

	function removeSpirit(id: string) {
		selected = selected.filter((spirit) => spirit.id !== id);
	}
</script>

<div bind:this={root} class="relative">
	<label class="block">
		<span class="mb-1 flex items-center justify-between gap-2 text-sm">
			<span>Chọn Hương linh chưa có bài vị</span>
			{#if selected.length}<span
					class="rounded-full bg-[var(--color-primary-soft)] px-2 py-0.5 text-xs font-semibold text-[var(--color-primary-dark)]"
					>Đã chọn {selected.length}</span
				>{/if}
		</span>
		<div class="relative">
			<span
				class="pointer-events-none absolute top-1/2 left-3 icon-[lucide--search] h-4 w-4 -translate-y-1/2 text-[var(--color-text-secondary)]"
			></span>
			<input
				bind:value={query}
				oninput={queueSearch}
				onfocus={() => (open = Boolean(query.trim()))}
				placeholder="Nhập tên, pháp danh, năm sinh..."
				class="h-10 w-full rounded-md border-[var(--color-border-strong)] pr-3 pl-9 text-sm"
			/>
		</div>
	</label>

	{#if open}<div
			class="absolute right-0 left-0 z-30 mt-1 max-h-64 overflow-y-auto rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-1 shadow-lg"
		>
			{#if loading}<div class="px-3 py-4 text-center text-sm text-[var(--color-text-secondary)]">
					Đang tìm...
				</div>{:else if availableResults.length === 0}<div
					class="px-3 py-4 text-center text-sm text-[var(--color-text-secondary)]"
				>
					Không tìm thấy Hương linh chưa xếp phù hợp
				</div>{:else}{#each availableResults as spirit (spirit.id)}<button
						type="button"
						onclick={() => selectSpirit(spirit)}
						class="flex w-full items-center gap-3 rounded-md px-3 py-2 text-left hover:bg-[var(--color-primary-soft)]"
					>
						{#if spirit.image_url}<img
								src={spirit.image_url}
								alt=""
								class="h-10 w-8 shrink-0 rounded object-cover"
							/>{:else}<span
								class="grid h-10 w-8 shrink-0 place-items-center rounded bg-[var(--color-surface-muted)]"
								><span class="icon-[lucide--user] h-4 w-4 opacity-50"></span></span
							>{/if}
						<span class="min-w-0 flex-1">
							<span class="block truncate text-sm font-semibold">{spirit.full_name}</span>
							<span class="block truncate text-xs text-[var(--color-text-secondary)]">
								{[spirit.dharma_name, spirit.birth_year, spirit.death_year]
									.filter(Boolean)
									.join(' · ') || 'Chưa có thông tin bổ sung'}
							</span>
						</span>
						<span class="icon-[lucide--plus] h-4 w-4 shrink-0"></span>
					</button>{/each}{/if}
		</div>{/if}

	{#if selected.length}<div class="mt-2 flex max-h-20 flex-wrap gap-1.5 overflow-y-auto">
			{#each selected as spirit (spirit.id)}<span
					class="inline-flex max-w-full items-center gap-1 rounded-md bg-[var(--color-surface-muted)] px-2 py-1 text-xs"
				>
					<span class="truncate font-medium">{spirit.full_name}</span>
					<button
						type="button"
						onclick={() => removeSpirit(spirit.id)}
						aria-label={`Bỏ chọn ${spirit.full_name}`}
						class="grid h-5 w-5 shrink-0 place-items-center rounded hover:bg-black/5"
						><span class="icon-[lucide--x] h-3.5 w-3.5"></span></button
					>
				</span>{/each}
		</div>{/if}
</div>
