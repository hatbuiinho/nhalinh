<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import Lightbox from '$lib/ui/Lightbox.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import InlineSpiritEditor from '$lib/memorial/InlineSpiritEditor.svelte';
	import SpiritImageUploader from '$lib/memorial/SpiritImageUploader.svelte';
	import { uploadSpiritImage } from '$lib/uploads/api';
	import {
		createTablet,
		createSpirits,
		deleteSpirit,
		downloadSpiritImportTemplate,
		exportSpiritsExcel,
		importSpiritsFromExcel,
		listAreas,
		listHouses,
		listPositions,
		listSpirits,
		listTablets,
		patchSpirit,
		previewSpiritImport,
		searchPositions,
		updateSpirit,
		type Area,
		type EditableSpiritInput,
		type House,
		type InlineSpiritInput,
		type Position,
		type Spirit,
		type SpiritImportPreview,
		type SpiritImportResult,
		type SpiritInput,
		type Tablet
	} from '$lib/memorial/api';
	import { emptyInlineSpirit, toInlineSpirit } from '$lib/memorial/sheet-parser';

	type DesktopView = 'list' | 'table';
	type ImportStep = 'guide' | 'upload' | 'preview';
	type ExportScope = 'current' | 'all';
	type SpiritSortKey = keyof Pick<
		Spirit,
		| 'image_url'
		| 'full_name'
		| 'dharma_name'
		| 'birth_year'
		| 'death_year'
		| 'age'
		| 'house_name'
		| 'area_code'
		| 'position_name'
		| 'tablet_name'
		| 'burial_place'
		| 'sender'
		| 'sent_month'
		| 'notes'
		| 'created_at'
		| 'updated_at'
	>;
	type SpiritColumn = { key: SpiritSortKey; label: string; width: string };
	type EditablePatchKey = Extract<
		SpiritSortKey,
		| 'full_name'
		| 'dharma_name'
		| 'birth_year'
		| 'death_year'
		| 'age'
		| 'burial_place'
		| 'sender'
		| 'sent_month'
		| 'notes'
	>;
	const spiritColumns: SpiritColumn[] = [
		{ key: 'image_url', label: 'Ảnh', width: 'w-20' },
		{ key: 'full_name', label: 'Họ tên', width: 'w-52' },
		{ key: 'dharma_name', label: 'Pháp danh', width: 'w-44' },
		{ key: 'birth_year', label: 'Năm sinh', width: 'w-28' },
		{ key: 'death_year', label: 'Năm mất', width: 'w-28' },
		{ key: 'age', label: 'Tuổi', width: 'w-24' },
		{ key: 'house_name', label: 'Nhà Linh', width: 'w-44' },
		{ key: 'area_code', label: 'Khu vực', width: 'w-28' },
		{ key: 'position_name', label: 'Vị trí', width: 'w-32' },
		{ key: 'tablet_name', label: 'Bài vị', width: 'w-44' },
		{ key: 'burial_place', label: 'Nơi an táng', width: 'w-52' },
		{ key: 'sender', label: 'Người gửi', width: 'w-44' },
		{ key: 'sent_month', label: 'Tháng gửi', width: 'w-32' },
		{ key: 'notes', label: 'Ghi chú', width: 'w-56' },
		{ key: 'created_at', label: 'Ngày tạo', width: 'w-36' },
		{ key: 'updated_at', label: 'Cập nhật', width: 'w-36' }
	];
	const spiritSortKeys = new Set<SpiritSortKey>(spiritColumns.map((column) => column.key));
	const vietnameseCollator = new Intl.Collator('vi', { numeric: true, sensitivity: 'base' });
	const editablePatchKeys = new Set<SpiritSortKey>([
		'full_name',
		'dharma_name',
		'birth_year',
		'death_year',
		'age',
		'burial_place',
		'sender',
		'sent_month',
		'notes'
	]);

	let houses = $state<House[]>([]),
		areas = $state<Area[]>([]),
		positions = $state<Position[]>([]),
		tablets = $state<Tablet[]>([]),
		spirits = $state<Spirit[]>([]);
	let houseId = $state(''),
		areaId = $state(''),
		query = $state(''),
		total = $state(0),
		loading = $state(true),
		saving = $state(false),
		imageUploading = $state(false),
		positionLoading = $state(false),
		tabletLoading = $state(false),
		desktopView = $state<DesktopView>('table'),
		contentFullscreen = $state(false),
		spiritSortKey = $state<SpiritSortKey>('full_name'),
		spiritSortDirection = $state<'asc' | 'desc'>('asc'),
		lightboxOpen = $state(false),
		lightboxSrc = $state(''),
		lightboxAlt = $state(''),
		importPopupOpen = $state(false),
		importStep = $state<ImportStep>('guide'),
		importHouseId = $state(''),
		importFile = $state<File | null>(null),
		importPreview = $state<SpiritImportPreview | null>(null),
		importBusy = $state(false),
		exportPopupOpen = $state(false),
		exportScope = $state<ExportScope>('current'),
		exportBusy = $state(false),
		inlineEditorBusy = $state(false),
		patchSaving = $state(false),
		formOpen = $state(false),
		editing = $state<Spirit | null>(null),
		formHouseId = $state(''),
		formPositionQuery = $state(''),
		formPositions = $state<Position[]>([]),
		formTablets = $state<Tablet[]>([]),
		selectedFormPosition = $state<Pick<Position, 'id' | 'name'> | null>(null),
		quickCreateTablet = $state(false),
		importInputKey = $state(0),
		timer: ReturnType<typeof setTimeout> | undefined,
		positionTimer: ReturnType<typeof setTimeout> | undefined,
		positionRequest = 0,
		tabletRequest = 0;
	let form = $state<SpiritInput>(emptyForm());
	let newSpirits = $state<EditableSpiritInput[]>([emptyInlineSpirit()]);
	let cellEditRoot = $state<HTMLFormElement>();
	let cellEditField = $state<HTMLInputElement | HTMLTextAreaElement>();
	let cellEdit = $state<{ id: string; key: EditablePatchKey; label: string; value: string } | null>(
		null
	);
	let selectedHouse = $derived(houses.find((v) => v.id === houseId));
	let canWrite = $derived(
		authStore.user?.role === 'admin' || selectedHouse?.access_role === 'editor'
	);
	let sortedSpirits = $derived.by(() => {
		const direction = spiritSortDirection === 'asc' ? 1 : -1;
		return [...spirits].sort((left, right) => {
			const result = vietnameseCollator.compare(
				String(left[spiritSortKey] ?? ''),
				String(right[spiritSortKey] ?? '')
			);
			return (result || vietnameseCollator.compare(left.full_name, right.full_name)) * direction;
		});
	});

	onMount(() => {
		restoreTablePreferences();
		document.addEventListener('pointerdown', dismissCellEditOnOutsideClick);
		void initialize();
		return () => {
			document.removeEventListener('pointerdown', dismissCellEditOnOutsideClick);
			if (timer) clearTimeout(timer);
			if (positionTimer) clearTimeout(positionTimer);
		};
	});
	function dismissCellEditOnOutsideClick(event: PointerEvent) {
		if (cellEdit && event.target instanceof Node && !cellEditRoot?.contains(event.target)) {
			cellEdit = null;
		}
	}
	function restoreTablePreferences() {
		try {
			const savedView = localStorage.getItem('nhalinh:spirit-view');
			const savedSortKey = localStorage.getItem('nhalinh:spirit-sort-key');
			const savedSortDirection = localStorage.getItem('nhalinh:spirit-sort-direction');
			if (savedView === 'list' || savedView === 'table') desktopView = savedView;
			if (savedSortKey && spiritSortKeys.has(savedSortKey as SpiritSortKey)) {
				spiritSortKey = savedSortKey as SpiritSortKey;
			}
			if (savedSortDirection === 'asc' || savedSortDirection === 'desc') {
				spiritSortDirection = savedSortDirection;
			}
		} catch {
			// Keep defaults when browser storage is unavailable.
		}
	}
	function setDesktopView(value: DesktopView) {
		desktopView = value;
		try {
			localStorage.setItem('nhalinh:spirit-view', value);
		} catch {
			// The view still changes without browser storage.
		}
	}
	function sortSpirits(key: SpiritSortKey) {
		if (spiritSortKey === key) {
			spiritSortDirection = spiritSortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			spiritSortKey = key;
			spiritSortDirection = 'asc';
		}
		try {
			localStorage.setItem('nhalinh:spirit-sort-key', spiritSortKey);
			localStorage.setItem('nhalinh:spirit-sort-direction', spiritSortDirection);
		} catch {
			// Sorting still works without browser storage.
		}
	}
	function openSpiritImage(item: Spirit) {
		if (!item.image_url) return;
		lightboxSrc = item.image_url;
		lightboxAlt = item.full_name;
		lightboxOpen = true;
	}
	function formatTimestamp(value: string) {
		return value ? new Intl.DateTimeFormat('vi-VN').format(new Date(value)) : '—';
	}
	function beginCellEdit(item: Spirit, column: SpiritColumn) {
		if (!canWrite || !editablePatchKeys.has(column.key)) return;
		cellEdit = {
			id: item.id,
			key: column.key as EditablePatchKey,
			label: column.label,
			value: item[column.key]
		};
		void focusCellEditField();
	}
	async function focusCellEditField() {
		await tick();
		if (!cellEditField) return;
		cellEditField.focus();
		const caretPosition = cellEditField.value.length;
		cellEditField.setSelectionRange(caretPosition, caretPosition);
	}
	async function saveCellPatch(event: SubmitEvent) {
		event.preventDefault();
		if (!cellEdit) return;
		patchSaving = true;
		try {
			const updated = await patchSpirit(cellEdit.id, cellEdit.key, cellEdit.value);
			spirits = spirits.map((item) => (item.id === updated.id ? updated : item));
			cellEdit = null;
			toastStore.success('Đã cập nhật Hương linh');
		} catch (error) {
			toastStore.error(message(error));
		} finally {
			patchSaving = false;
		}
	}
	async function initialize() {
		loading = true;
		try {
			houses = await listHouses();
			houseId = houses[0]?.id ?? '';
			await changeHouse();
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			loading = false;
		}
	}
	async function changeHouse() {
		areaId = '';
		areas = houseId ? await listAreas(houseId) : [];
		positions = [];
		tablets = [];
		await load();
	}
	async function changeArea() {
		positions = areaId ? await listPositions(areaId) : [];
		tablets = (await Promise.all(positions.map((position) => listTablets(position.id)))).flat();
		await load();
	}
	async function load() {
		const page = await listSpirits(query, houseId, areaId);
		spirits = page.spirits;
		total = page.total;
	}
	function search() {
		if (timer) clearTimeout(timer);
		timer = setTimeout(() => void load().catch((e) => toastStore.error(message(e))), 300);
	}
	function suggestedImportHouse() {
		if (houseId && houses.some((house) => house.id === houseId)) return houseId;
		if (houses.length === 1) return houses[0].id;
		return houses[0]?.id ?? '';
	}
	function openImportPopup() {
		importPopupOpen = true;
		importStep = 'guide';
		importHouseId = suggestedImportHouse();
		importFile = null;
		importPreview = null;
		importInputKey++;
	}
	function closeImportPopup() {
		if (importBusy) return;
		importPopupOpen = false;
		importStep = 'guide';
		importFile = null;
		importPreview = null;
		importInputKey++;
	}
	function continueImportGuide() {
		if (!importHouseId) {
			toastStore.error('Vui lòng chọn Nhà Linh để import');
			return;
		}
		importStep = 'upload';
	}
	function selectImportFile(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		importFile = input.files?.[0] ?? null;
		importPreview = null;
	}
	async function downloadImportTemplateFile() {
		importBusy = true;
		try {
			const { blob, filename } = await downloadSpiritImportTemplate();
			downloadBlob(blob, extractFilename(filename) || 'huong-linh-import-template.xlsx');
		} catch (error) {
			toastStore.error(message(error));
		} finally {
			importBusy = false;
		}
	}
	async function requestSpiritImportPreview() {
		if (!importHouseId) {
			toastStore.error('Vui lòng chọn Nhà Linh để import');
			return;
		}
		if (!importFile) {
			toastStore.error('Vui lòng chọn file Excel .xlsx');
			return;
		}
		importBusy = true;
		try {
			importPreview = await previewSpiritImport(importFile, importHouseId);
			importStep = 'preview';
		} catch (error) {
			toastStore.error(message(error));
		} finally {
			importBusy = false;
		}
	}
	async function commitSpiritImport() {
		if (!importHouseId || !importFile || !importPreview) return;
		importBusy = true;
		try {
			const result = await importSpiritsFromExcel(importFile, importHouseId);
			importBusy = false;
			toastStore.success(importSummary(result));
			closeImportPopup();
			await changeHouseAwareRefresh();
		} catch (error) {
			toastStore.error(message(error));
		} finally {
			importBusy = false;
		}
	}
	async function changeHouseAwareRefresh() {
		await load();
	}
	function importSummary(result: SpiritImportResult) {
		const segments = [
			`Đã import ${result.created_spirit_count} Hương linh`,
			result.created_area_count > 0 ? `tạo ${result.created_area_count} khu vực` : '',
			result.created_position_count > 0 ? `tạo ${result.created_position_count} vị trí` : '',
			result.created_tablet_count > 0 ? `tạo ${result.created_tablet_count} bài vị` : ''
		].filter(Boolean);
		return segments.join(' · ');
	}
	function openExportPopup() {
		exportScope = 'current';
		exportPopupOpen = true;
	}
	function closeExportPopup() {
		if (exportBusy) return;
		exportPopupOpen = false;
	}
	async function exportSpiritWorkbook() {
		exportBusy = true;
		try {
			const { blob, filename } = await exportSpiritsExcel(exportScope, { query, houseId, areaId });
			downloadBlob(blob, extractFilename(filename) || `huong-linh-${exportScope}.xlsx`);
			closeExportPopup();
			toastStore.success(
				exportScope === 'all'
					? 'Đã tải file Excel toàn bộ Hương linh'
					: 'Đã tải file Excel theo dữ liệu đang lọc'
			);
		} catch (error) {
			toastStore.error(message(error));
		} finally {
			exportBusy = false;
		}
	}
	function extractFilename(contentDisposition: string | null) {
		if (!contentDisposition) return '';
		const match = contentDisposition.match(/filename="?([^"]+)"?/i);
		return match?.[1] ?? '';
	}
	function downloadBlob(blob: Blob, filename: string) {
		const url = URL.createObjectURL(blob);
		const link = document.createElement('a');
		link.href = url;
		link.download = filename;
		document.body.append(link);
		link.click();
		link.remove();
		URL.revokeObjectURL(url);
	}
	function add() {
		editing = null;
		form = emptyForm();
		newSpirits = [emptyInlineSpirit()];
		inlineEditorBusy = false;
		formHouseId = preferredFormHouse();
		form.house_id = formHouseId;
		formPositionQuery = '';
		formPositions = [];
		formTablets = [];
		selectedFormPosition = null;
		quickCreateTablet = false;
		positionRequest++;
		tabletRequest++;
		positionLoading = false;
		tabletLoading = false;
		formOpen = true;
	}
	async function edit(item: Spirit) {
		editing = item;
		formHouseId = item.house_id;
		form = {
			house_id: item.house_id,
			tablet_id: item.tablet_id,
			full_name: item.full_name,
			dharma_name: item.dharma_name,
			birth_year: item.birth_year,
			death_year: item.death_year,
			age: item.age,
			image_url: item.image_url,
			burial_place: item.burial_place,
			sender: item.sender,
			sent_month: item.sent_month,
			notes: item.notes
		};
		formPositionQuery = item.position_name;
		formPositions = [];
		selectedFormPosition = item.position_id
			? { id: item.position_id, name: item.position_name }
			: null;
		formTablets = [];
		quickCreateTablet = false;
		formOpen = true;
		if (!item.position_id) return;
		const request = ++tabletRequest;
		tabletLoading = true;
		try {
			const items = await listTablets(item.position_id);
			if (request === tabletRequest) formTablets = items;
		} finally {
			if (request === tabletRequest) tabletLoading = false;
		}
	}
	function preferredFormHouse() {
		if (houses.length === 1) return houses[0].id;
		let saved: string | null = null;
		try {
			saved = localStorage.getItem('nhalinh:last-spirit-house');
		} catch {
			// Fall back to the current or first accessible house.
		}
		if (saved && houses.some((house) => house.id === saved)) return saved;
		if (houseId && houses.some((house) => house.id === houseId)) return houseId;
		return houses[0]?.id ?? '';
	}
	function changeFormHouse() {
		form.house_id = formHouseId;
		if (houses.length > 1) {
			try {
				localStorage.setItem('nhalinh:last-spirit-house', formHouseId);
			} catch {
				// House selection still works when browser storage is unavailable.
			}
		}
		formPositionQuery = '';
		formPositions = [];
		formTablets = [];
		selectedFormPosition = null;
		form.tablet_id = '';
		positionRequest++;
		tabletRequest++;
		positionLoading = false;
		tabletLoading = false;
	}
	function searchFormPosition() {
		selectedFormPosition = null;
		form.tablet_id = '';
		quickCreateTablet = false;
		formTablets = [];
		tabletRequest++;
		tabletLoading = false;
		if (positionTimer) clearTimeout(positionTimer);
		if (!formPositionQuery.trim()) {
			positionRequest++;
			formPositions = [];
			positionLoading = false;
			return;
		}
		positionTimer = setTimeout(
			() => void loadFormPositions().catch((error) => toastStore.error(message(error))),
			250
		);
	}
	async function loadFormPositions() {
		const request = ++positionRequest;
		if (!formHouseId || !formPositionQuery.trim()) {
			formPositions = [];
			return;
		}
		positionLoading = true;
		try {
			const items = await searchPositions(formHouseId, formPositionQuery);
			if (request === positionRequest) formPositions = items;
		} finally {
			if (request === positionRequest) positionLoading = false;
		}
	}
	async function selectFormPosition(position: Position) {
		positionRequest++;
		selectedFormPosition = position;
		formPositionQuery = position.name;
		form.tablet_id = '';
		quickCreateTablet = false;
		formTablets = [];
		const request = ++tabletRequest;
		tabletLoading = true;
		try {
			const items = await listTablets(position.id);
			if (request === tabletRequest) {
				formTablets = items;
				if (items.length === 1) form.tablet_id = items[0].id;
			}
		} finally {
			if (request === tabletRequest) tabletLoading = false;
		}
	}
	function clearFormPosition() {
		selectedFormPosition = null;
		formPositionQuery = '';
		formPositions = [];
		formTablets = [];
		form.tablet_id = '';
		quickCreateTablet = false;
		positionRequest++;
		tabletRequest++;
		positionLoading = false;
		tabletLoading = false;
	}
	async function save(e: SubmitEvent) {
		e.preventDefault();
		saving = true;
		try {
			if (editing) {
				await updateSpirit(editing.id, form);
			} else if (quickCreateTablet && selectedFormPosition) {
				const rows = validNewSpiritRows();
				await createTablet({
					position_id: selectedFormPosition.id,
					name: rows[0].full_name,
					notes: '',
					spirits: rows
				});
			} else {
				if (selectedFormPosition && !form.tablet_id) {
					throw new Error('Vui lòng chọn bài vị hoặc tạo nhanh bài vị mới');
				}
				const rows = validNewSpiritRows();
				await createSpirits(
					rows.map((spirit) => ({
						...spirit,
						house_id: formHouseId,
						tablet_id: form.tablet_id
					}))
				);
			}
			toastStore.success(
				quickCreateTablet
					? 'Đã tạo bài vị và thêm Hương linh'
					: editing
						? 'Đã cập nhật Hương linh'
						: `Đã thêm ${newSpirits.filter(hasSpiritData).length} Hương linh`
			);
			formOpen = false;
			await load();
		} catch (err) {
			toastStore.error(message(err));
		} finally {
			saving = false;
			quickCreateTablet = false;
		}
	}
	function hasSpiritData(spirit: EditableSpiritInput) {
		return Object.entries(spirit).some(
			([key, value]) => key !== 'id' && String(value).trim() !== ''
		);
	}
	function validNewSpiritRows(): InlineSpiritInput[] {
		const rows = newSpirits.filter(hasSpiritData).map(toInlineSpirit);
		if (rows.length === 0) throw new Error('Vui lòng nhập ít nhất một Hương linh');
		if (rows.length > 500) throw new Error('Chỉ có thể thêm tối đa 500 Hương linh mỗi lần');
		const missingName = rows.findIndex((spirit) => !spirit.full_name.trim());
		if (missingName >= 0) throw new Error(`Dòng ${missingName + 1}: Tên không được để trống`);
		return rows;
	}
	async function selectSpiritImage(file: File) {
		imageUploading = true;
		try {
			form.image_url = await uploadSpiritImage(file);
			toastStore.success('Đã tải ảnh Hương linh');
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			imageUploading = false;
		}
	}
	async function remove(item: Spirit) {
		const ok = await popupStore.confirm({
			title: 'Xoá Hương linh?',
			message: `Thông tin của ${item.full_name} sẽ bị xoá vĩnh viễn.`,
			confirmLabel: 'Xoá',
			tone: 'danger'
		});
		if (!ok) return;
		try {
			await deleteSpirit(item.id);
			toastStore.success('Đã xoá Hương linh');
			await load();
		} catch (e) {
			toastStore.error(message(e));
		}
	}
	function emptyForm(): SpiritInput {
		return {
			house_id: '',
			tablet_id: '',
			full_name: '',
			dharma_name: '',
			birth_year: '',
			death_year: '',
			age: '',
			image_url: '',
			burial_place: '',
			sender: '',
			sent_month: '',
			notes: ''
		};
	}
	function message(e: unknown) {
		return e instanceof Error ? e.message : 'Có lỗi xảy ra';
	}
	function closeActiveForm(event: KeyboardEvent) {
		if (event.key !== 'Escape') return;
		if (cellEdit && !patchSaving) {
			cellEdit = null;
			return;
		}
		if (importPopupOpen && !importBusy) {
			closeImportPopup();
			return;
		}
		if (exportPopupOpen && !exportBusy) {
			closeExportPopup();
			return;
		}
		if (document.querySelector('[role="dialog"][aria-modal="true"]')) return;
		if (formOpen) {
			if (!saving && !imageUploading) formOpen = false;
			return;
		}
		if (contentFullscreen) contentFullscreen = false;
	}
</script>

<svelte:window onkeydown={closeActiveForm} />

<section
	class={[
		'flex h-full min-w-0 flex-col overflow-hidden',
		contentFullscreen && 'fixed inset-0 z-30 bg-[var(--color-bg)]'
	]}
>
	<div class="border-b border-[var(--color-border)] bg-[var(--color-bg)] px-4 py-3 md:px-6 lg:px-8">
		<div class="mx-auto grid max-w-[1320px] gap-2 md:grid-cols-[220px_180px_1fr_auto]">
			<select
				bind:value={houseId}
				onchange={() => void changeHouse()}
				aria-label="Chọn Nhà Linh"
				class="h-11 rounded-md border-[var(--color-border-strong)] text-sm"
				><option value="">Tất cả Nhà Linh</option>{#each houses as h (h.id)}<option value={h.id}
						>{h.name}</option
					>{/each}</select
			>
			<select
				bind:value={areaId}
				onchange={() => void changeArea()}
				aria-label="Chọn khu vực"
				class="h-11 rounded-md border-[var(--color-border-strong)] text-sm"
				><option value="">Tất cả khu vực</option>{#each areas as a (a.id)}<option value={a.id}
						>Khu {a.code}{a.name ? ` – ${a.name}` : ''}</option
					>{/each}</select
			>
			<label class="relative"
				><span
					class="absolute top-3.5 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"
				></span><input
					bind:value={query}
					oninput={search}
					placeholder="Tìm tên, pháp danh, vị trí, người gửi..."
					class="h-11 w-full rounded-md border-[var(--color-border-strong)] pl-9 text-sm"
				/></label
			>
			{#if canWrite}<button
					onclick={() => void add()}
					disabled={houses.length === 0}
					class="h-11 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white disabled:opacity-40"
					><span class="mr-1 icon-[lucide--plus] inline-block h-4 w-4 align-text-bottom"></span>Thêm
					Hương linh</button
				>{/if}
		</div>
	</div>
	<div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden px-4 pb-4 md:px-6 lg:px-8">
		<div class="mx-auto flex min-h-0 w-full max-w-[1320px] min-w-0 flex-1 flex-col">
			<div
				class="z-20 mb-3 flex shrink-0 items-center justify-between gap-3 bg-[var(--color-bg)] pt-2"
			>
				<p class="text-xs text-[var(--color-text-secondary)]">
					{total} Hương linh{selectedHouse ? ` · ${selectedHouse.name}` : ''}
				</p>
				<div class="flex items-center gap-2">
					<button
						type="button"
						onclick={openExportPopup}
						class="h-9 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-xs font-semibold hover:bg-[var(--color-surface-muted)]"
					>
						<span class="mr-1 icon-[lucide--download] inline-block h-4 w-4 align-text-bottom"></span>
						Export Excel
					</button>
					{#if canWrite}
						<button
							type="button"
							onclick={openImportPopup}
							disabled={houses.length === 0}
							class="h-9 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-xs font-semibold hover:bg-[var(--color-surface-muted)] disabled:opacity-50"
						>
							<span class="mr-1 icon-[lucide--upload] inline-block h-4 w-4 align-text-bottom"></span>
							Import Excel
						</button>
					{/if}
					<div
						class="hidden items-center rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] p-0.5 md:flex"
						aria-label="Kiểu hiển thị"
					>
						<button
							type="button"
							onclick={() => setDesktopView('list')}
							aria-pressed={desktopView === 'list'}
							class={[
								'grid h-8 w-9 place-items-center rounded-sm',
								desktopView === 'list'
									? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
									: 'text-[var(--color-text-muted)]'
							]}
							aria-label="Xem dạng danh sách"
						>
							<span class="icon-[lucide--layout-grid] h-4 w-4" aria-hidden="true"></span>
						</button>
						<button
							type="button"
							onclick={() => setDesktopView('table')}
							aria-pressed={desktopView === 'table'}
							class={[
								'grid h-8 w-9 place-items-center rounded-sm',
								desktopView === 'table'
									? 'bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)]'
									: 'text-[var(--color-text-muted)]'
							]}
							aria-label="Xem dạng bảng"
						>
							<span class="icon-[lucide--table-2] h-4 w-4" aria-hidden="true"></span>
						</button>
					</div>
					<button
						type="button"
						onclick={() => (contentFullscreen = !contentFullscreen)}
						aria-label={contentFullscreen
							? 'Thu nhỏ màn hình Hương linh'
							: 'Phóng to toàn màn hình Hương linh'}
						title={contentFullscreen ? 'Thu nhỏ' : 'Phóng to toàn màn hình'}
						class="grid h-9 w-9 shrink-0 place-items-center rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
					>
						<span
							class={[
								'h-4 w-4',
								contentFullscreen ? 'icon-[lucide--minimize-2]' : 'icon-[lucide--maximize-2]'
							]}
							aria-hidden="true"
						></span>
					</button>
				</div>
			</div>
			{#if loading}<div class="min-h-0 flex-1 overflow-y-auto py-16">
					<LoadingIndicator label="Đang tải dữ liệu..." />
				</div>
			{:else if spirits.length === 0}<div
					class="min-h-0 flex-1 overflow-y-auto rounded-md border border-dashed border-[var(--color-border-strong)] py-16 text-center"
				>
					<span class="mx-auto icon-[lucide--search-x] block h-8 w-8 text-[var(--color-text-muted)]"
					></span>
					<p class="mt-3 text-sm">Không có Hương linh phù hợp</p>
					{#if areaId && tablets.length === 0}<p
							class="mt-1 text-xs text-[var(--color-text-secondary)]"
						>
							Hãy tạo bài vị trong mục Cơ cấu tổ chức trước.
						</p>{/if}
				</div>
			{:else if desktopView === 'table'}<div class="min-h-0 flex-1 overflow-y-auto md:hidden">
					<ul class="grid gap-3">
						{#each spirits as item (item.id)}{@render spiritCard(item)}{/each}
					</ul>
				</div>
				{@render spiritTable()}
			{:else}<div class="min-h-0 flex-1 overflow-y-auto">
					<ul class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
						{#each spirits as item (item.id)}{@render spiritCard(item)}{/each}
					</ul>
				</div>{/if}
		</div>
	</div>
</section>

<Lightbox src={lightboxSrc} alt={lightboxAlt} bind:open={lightboxOpen} />

<Popup
	open={importPopupOpen}
	title={importStep === 'guide'
		? 'Hướng dẫn import Excel'
		: importStep === 'upload'
			? 'Chọn file Excel'
			: 'Xem trước import Excel'}
	onClose={closeImportPopup}
>
	<div class="space-y-4 text-sm">
		{#if importStep === 'guide'}
			<label class="block">
				<span class="mb-1.5 block font-medium">Nhà Linh import vào *</span>
				<select
					bind:value={importHouseId}
					class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
				>
					<option value="">Chọn Nhà Linh</option>
					{#each houses as house (house.id)}
						<option value={house.id}>{house.name}</option>
					{/each}
				</select>
			</label>
			<div class="rounded-md bg-[var(--color-surface-muted)] p-3 text-xs text-[var(--color-text-secondary)]">
				<p class="font-semibold text-[var(--color-text)]">File Excel cần đúng format chuẩn</p>
				<p class="mt-2">
					Các cột bắt buộc theo mẫu: Họ tên, Pháp danh, Năm sinh, Năm mất, Tuổi, Ảnh URL,
					Nơi an táng, Người gửi, Tháng gửi, Ghi chú, Vị trí, Bài vị.
				</p>
				<p class="mt-2">
					Nếu vị trí có dạng như <code>38D-10</code>, hệ thống sẽ tự parse thành Khu D, hàng 38,
					cột 10 và tự tạo Khu, Vị trí, Bài vị nếu chưa có.
				</p>
				<p class="mt-2">
					Nếu cột Bài vị để trống thì hệ thống sẽ lấy tên Hương linh làm tên bài vị mặc định.
				</p>
			</div>
			<button
				type="button"
				onclick={downloadImportTemplateFile}
				class="inline-flex h-10 items-center rounded-md border border-[var(--color-border-strong)] px-3 text-xs font-semibold"
			>
				<span class="mr-1 icon-[lucide--file-spreadsheet] h-4 w-4"></span>Tải file mẫu
			</button>
		{:else if importStep === 'upload'}
			<div class="rounded-md bg-[var(--color-surface-muted)] p-3 text-xs text-[var(--color-text-secondary)]">
				File chỉ hỗ trợ định dạng <code>.xlsx</code>. Dữ liệu sẽ được preview trước, chưa ghi vào hệ
				thống cho tới khi bạn xác nhận import.
			</div>
			<label class="block">
				<span class="mb-1.5 block font-medium">File Excel *</span>
				{#key importInputKey}
					<input
						type="file"
						accept=".xlsx"
						onchange={selectImportFile}
						class="block w-full rounded-md border border-[var(--color-border-strong)] px-3 py-2 text-sm"
					/>
				{/key}
			</label>
			{#if importFile}
				<p class="text-xs text-[var(--color-text-secondary)]">
					Đã chọn: <span class="font-medium text-[var(--color-text)]">{importFile.name}</span>
				</p>
			{/if}
		{:else if importPreview}
			<div class="grid grid-cols-2 gap-3 text-xs">
				<div class="rounded-md bg-[var(--color-surface-muted)] p-3">
					<p class="text-[var(--color-text-secondary)]">Tổng số dòng</p>
					<p class="mt-1 text-lg font-semibold text-[var(--color-text)]">{importPreview.total_rows}</p>
				</div>
				<div class="rounded-md bg-[var(--color-surface-muted)] p-3">
					<p class="text-[var(--color-text-secondary)]">Hợp lệ</p>
					<p class="mt-1 text-lg font-semibold text-[var(--color-text)]">{importPreview.valid_rows}</p>
				</div>
				<div class="rounded-md bg-[var(--color-surface-muted)] p-3">
					<p class="text-[var(--color-text-secondary)]">Sẽ tạo mới</p>
					<p class="mt-1 font-semibold text-[var(--color-text)]">
						{importPreview.create_area_count} khu · {importPreview.create_position_count} vị trí ·
						{importPreview.create_tablet_count} bài vị
					</p>
				</div>
				<div class="rounded-md bg-[var(--color-surface-muted)] p-3">
					<p class="text-[var(--color-text-secondary)]">Dòng lỗi</p>
					<p class="mt-1 text-lg font-semibold text-[var(--color-danger)]">{importPreview.invalid_rows}</p>
				</div>
			</div>
			{#if importPreview.errors.length > 0}
				<div class="rounded-md border border-[var(--color-danger)]/30 bg-[var(--color-danger)]/5 p-3">
					<p class="text-sm font-semibold text-[var(--color-danger)]">Các dòng cần sửa trước khi import</p>
					<ul class="mt-2 max-h-44 space-y-2 overflow-y-auto pr-1 text-xs text-[var(--color-text-secondary)]">
						{#each importPreview.errors as issue (`${issue.row_number}-${issue.message}`)}
							<li>Dòng {issue.row_number}: {issue.message}</li>
						{/each}
					</ul>
				</div>
			{/if}
		{/if}
	</div>

	{#snippet footer()}
		<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold"
				onclick={() => {
					if (importStep === 'guide') closeImportPopup();
					else if (importStep === 'upload') importStep = 'guide';
					else importStep = 'upload';
				}}
			>
				{importStep === 'guide' ? 'Đóng' : 'Quay lại'}
			</button>
			<button
				type="button"
				disabled={importBusy || (importStep === 'guide' && !importHouseId) || (importStep === 'upload' && !importFile) || (importStep === 'preview' && Boolean(importPreview?.invalid_rows))}
				class="flex h-11 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-50"
				onclick={() => {
					if (importStep === 'guide') continueImportGuide();
					else if (importStep === 'upload') void requestSpiritImportPreview();
					else void commitSpiritImport();
				}}
			>
				{#if importBusy}
					<span class="icon-[lucide--loader-circle] h-4 w-4 animate-spin" aria-hidden="true"></span>
				{/if}
				{importStep === 'guide'
					? 'Tiếp tục'
					: importStep === 'upload'
						? 'Xem trước'
						: importPreview?.invalid_rows
							? 'Cần sửa file trước'
							: 'Xác nhận import'}
			</button>
		</div>
	{/snippet}
</Popup>

<Popup open={exportPopupOpen} title="Export Excel" onClose={closeExportPopup}>
	<div class="space-y-3 text-sm">
		<label class="flex items-start gap-3 rounded-md border border-[var(--color-border)] p-3">
			<input
				type="radio"
				name="export-scope"
				value="current"
				checked={exportScope === 'current'}
				onchange={() => (exportScope = 'current')}
				class="mt-1"
			/>
			<span>
				<span class="block font-medium">Xuất dữ liệu đang hiển thị</span>
				<span class="mt-1 block text-xs text-[var(--color-text-secondary)]">
					Áp dụng theo bộ lọc hiện tại: Nhà Linh, khu vực và từ khoá tìm kiếm.
				</span>
			</span>
		</label>
		<label class="flex items-start gap-3 rounded-md border border-[var(--color-border)] p-3">
			<input
				type="radio"
				name="export-scope"
				value="all"
				checked={exportScope === 'all'}
				onchange={() => (exportScope = 'all')}
				class="mt-1"
			/>
			<span>
				<span class="block font-medium">Xuất toàn bộ</span>
				<span class="mt-1 block text-xs text-[var(--color-text-secondary)]">
					Xuất toàn bộ Hương linh trong phạm vi bạn được phép xem.
				</span>
			</span>
		</label>
	</div>

	{#snippet footer()}
		<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				class="h-11 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-sm font-semibold"
				onclick={closeExportPopup}>Huỷ</button
			>
			<button
				type="button"
				disabled={exportBusy}
				class="flex h-11 items-center justify-center gap-2 rounded-md bg-[var(--color-primary)] text-sm font-semibold text-white disabled:opacity-50"
				onclick={() => void exportSpiritWorkbook()}
			>
				{#if exportBusy}
					<span class="icon-[lucide--loader-circle] h-4 w-4 animate-spin" aria-hidden="true"></span>
				{/if}
				Tải file Excel
			</button>
		</div>
	{/snippet}
</Popup>

{#if formOpen}<div
		class="fixed inset-0 z-50 grid place-items-end bg-black/40 md:place-items-center"
		role="presentation"
		onclick={(e) => {
			if (e.target === e.currentTarget) formOpen = false;
		}}
	>
		<form
			onsubmit={save}
			class={[
				'max-h-[94dvh] w-full overflow-y-auto rounded-t-xl bg-[var(--color-surface)] shadow-xl md:rounded-xl',
				editing ? 'md:max-w-2xl' : 'md:max-w-6xl'
			]}
		>
			<header
				class="sticky top-0 z-40 flex items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-surface)] px-5 py-4"
			>
				<h2 class="text-lg font-semibold">
					{editing ? 'Sửa Hương linh' : 'Thêm nhiều Hương linh'}
				</h2>
				<button
					type="button"
					onclick={() => (formOpen = false)}
					aria-label="Đóng"
					class="icon-[lucide--x] h-5 w-5 cursor-pointer"
				></button>
			</header>
			<div class="grid gap-4 p-5 md:grid-cols-2">
				{#if editing}<div class="md:col-span-2">
						<SpiritImageUploader
							imageUrl={form.image_url}
							displayName={form.full_name || editing.full_name}
							uploading={imageUploading}
							onselect={selectSpiritImage}
						/>
					</div>{/if}
				{#if houses.length > 1}<label
						><span class="mb-1 block text-sm font-medium">Nhà Linh *</span><select
							bind:value={formHouseId}
							disabled={Boolean(editing)}
							onchange={() => void changeFormHouse()}
							required
							class="h-11 w-full rounded-md border-[var(--color-border-strong)] disabled:opacity-60"
							>{#each houses as house (house.id)}<option value={house.id}>{house.name}</option
								>{/each}</select
						></label
					>{/if}
				<div class={`relative ${houses.length > 1 ? '' : 'md:col-span-2'}`}>
					<label class="block"
						><span class="mb-1 block text-sm font-medium">Vị trí (không bắt buộc)</span>
						<div class="flex gap-2">
							<div class="relative min-w-0 flex-1">
								<input
									bind:value={formPositionQuery}
									oninput={searchFormPosition}
									placeholder="Tìm trực tiếp, ví dụ 1a1..."
									autocomplete="off"
									role="combobox"
									aria-autocomplete="list"
									aria-controls="position-suggestions"
									aria-expanded={positionLoading || formPositions.length > 0}
									class="h-11 w-full rounded-md border-[var(--color-border-strong)] pr-9"
								/>{#if positionLoading}<span
										class="pointer-events-none absolute top-3.5 right-3 icon-[lucide--loader-circle] h-4 w-4 animate-spin text-[var(--color-text-muted)]"
									></span>{/if}
							</div>
							{#if selectedFormPosition}<button
									type="button"
									onclick={clearFormPosition}
									class="h-11 rounded-md border border-[var(--color-border-strong)] px-3 text-xs font-semibold"
									>Bỏ chọn</button
								>{/if}
						</div></label
					>
					{#if !selectedFormPosition && formPositions.length > 0}<div
							id="position-suggestions"
							class="absolute top-full right-0 left-0 z-30 mt-1 max-h-52 overflow-y-auto rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] shadow-lg"
							role="listbox"
						>
							{#each formPositions as position (position.id)}<button
									type="button"
									role="option"
									aria-selected="false"
									onclick={() => void selectFormPosition(position)}
									class="block w-full border-b border-[var(--color-border)] px-3 py-2 text-left text-sm last:border-b-0 hover:bg-[var(--color-primary-soft)]"
									><strong>{position.name}</strong><span
										class="ml-2 text-xs text-[var(--color-text-secondary)]"
										>Khu {position.area_code} · hàng {position.row_number}, cột {position.column_number}
										·
										{position.tablet_count} bài vị</span
									></button
								>{/each}
						</div>{/if}
				</div>
				{#if selectedFormPosition}<label class="md:col-span-2"
						><span class="mb-1 block text-sm font-medium">Bài vị hiện có</span><select
							bind:value={form.tablet_id}
							onchange={() => (quickCreateTablet = false)}
							disabled={tabletLoading}
							class="h-11 w-full rounded-md border-[var(--color-border-strong)] disabled:opacity-60"
							><option value=""
								>{tabletLoading
									? 'Đang tải danh sách bài vị...'
									: 'Chưa chọn — có thể tạo nhanh bài vị mới'}</option
							>{#each formTablets as tablet (tablet.id)}<option value={tablet.id}
									>{tablet.name} · {tablet.spirit_count} Hương linh</option
								>{/each}</select
						></label
					>{:else}<p
						class="rounded-md bg-[var(--color-primary-soft)] px-3 py-2 text-xs text-[var(--color-primary-dark)] md:col-span-2"
					>
						Để trống vị trí để lưu Hương linh vào danh sách chưa xếp.
					</p>{/if}
				{#if editing}{@render field('Họ tên *', 'full_name', true)}{@render field(
						'Pháp danh',
						'dharma_name'
					)}{@render field('Năm sinh', 'birth_year')}{@render field(
						'Năm mất',
						'death_year'
					)}{@render field('Tuổi', 'age')}{@render field('Tháng gửi', 'sent_month')}{@render field(
						'Nơi an táng',
						'burial_place'
					)}{@render field('Người gửi', 'sender')}<label class="md:col-span-2"
						><span class="mb-1 block text-sm font-medium">Ghi chú</span><textarea
							bind:value={form.notes}
							rows="3"
							class="w-full rounded-md border-[var(--color-border-strong)]"></textarea></label
					>{:else}<div class="h-[52dvh] min-h-[360px] md:col-span-2">
						<InlineSpiritEditor
							bind:items={newSpirits}
							onbusychange={(busy) => (inlineEditorBusy = busy)}
						/>
					</div>{/if}
			</div>
			<footer
				class="sticky bottom-0 z-40 flex justify-end gap-3 border-t border-[var(--color-border)] bg-[var(--color-surface)] px-5 py-4"
			>
				<button
					type="button"
					onclick={() => (formOpen = false)}
					class="h-11 cursor-pointer rounded-md border border-[var(--color-border-strong)] px-5 text-sm font-semibold"
					>Huỷ</button
				>{#if !editing && selectedFormPosition && form.tablet_id}<button
						type="submit"
						disabled={saving || imageUploading || inlineEditorBusy}
						onclick={() => (quickCreateTablet = true)}
						class="h-11 cursor-pointer rounded-md border border-[var(--color-primary)] px-4 text-sm font-semibold text-[var(--color-primary-dark)] disabled:cursor-not-allowed disabled:opacity-50"
						>Tạo bài vị mới</button
					>{/if}<button
					type="submit"
					disabled={saving ||
						imageUploading ||
						inlineEditorBusy ||
						Boolean(editing && selectedFormPosition && !form.tablet_id) ||
						(!editing &&
							selectedFormPosition &&
							!form.tablet_id &&
							!newSpirits[0]?.full_name.trim())}
					onclick={() =>
						(quickCreateTablet = Boolean(!editing && selectedFormPosition && !form.tablet_id))}
					class="h-11 cursor-pointer rounded-md bg-[var(--color-primary)] px-6 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50"
					>{saving
						? 'Đang lưu...'
						: !editing && selectedFormPosition && !form.tablet_id
							? 'Tạo bài vị & thêm'
							: !editing && !selectedFormPosition
								? 'Lưu chưa xếp vị trí'
								: 'Lưu'}</button
				>
			</footer>
		</form>
	</div>{/if}

{#snippet field(label: string, key: keyof SpiritInput, required = false)}<label
		><span class="mb-1 block text-sm font-medium">{label}</span><input
			bind:value={form[key]}
			{required}
			class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
		/></label
	>{/snippet}

{#snippet spiritCard(item: Spirit)}<li
		class="rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-4"
	>
		<div class="flex gap-3">
			{#if item.image_url}<button
					type="button"
					onclick={() => openSpiritImage(item)}
					class="shrink-0 rounded-md focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-primary)]"
					aria-label={`Xem ảnh ${item.full_name}`}
					><img
						src={item.image_url}
						alt={item.full_name}
						class="h-20 w-15 rounded-md object-cover"
					/></button
				>{:else}<span
					class="grid h-20 w-15 shrink-0 place-items-center rounded-md bg-[var(--color-primary-soft)] font-semibold text-[var(--color-primary-dark)]"
					>{item.full_name.slice(0, 1)}</span
				>{/if}
			<div class="min-w-0 flex-1">
				<h3 class="truncate font-semibold">{item.full_name}</h3>
				<p class="truncate text-sm text-[var(--color-text-secondary)]">
					{item.dharma_name || 'Chưa có pháp danh'}
				</p>
				<p class="mt-1 text-xs font-semibold text-[var(--color-primary-dark)]">
					{item.position_id
						? `Khu ${item.area_code} · ${item.position_name} · ${item.tablet_name}`
						: 'Chưa xếp vị trí'}
				</p>
			</div>
		</div>
		<dl
			class="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 border-t border-[var(--color-border)] pt-3 text-xs"
		>
			<dt class="text-[var(--color-text-muted)]">Năm sinh – mất</dt>
			<dd class="text-right">{item.birth_year || '?'} – {item.death_year || '?'}</dd>
			<dt class="text-[var(--color-text-muted)]">Người gửi</dt>
			<dd class="truncate text-right">{item.sender || '—'}</dd>
			<dt class="text-[var(--color-text-muted)]">Nơi an táng</dt>
			<dd class="truncate text-right">{item.burial_place || '—'}</dd>
		</dl>
		{#if canWrite}<div class="mt-3 flex justify-end gap-2">
				<button
					type="button"
					onclick={() => void edit(item)}
					class="h-9 rounded-md border border-[var(--color-border-strong)] px-3 text-xs font-semibold"
					>Sửa</button
				><button
					type="button"
					onclick={() => void remove(item)}
					class="h-9 rounded-md border border-[var(--color-danger)] px-3 text-xs font-semibold text-[var(--color-danger)]"
					>Xoá</button
				>
			</div>{/if}
	</li>{/snippet}

{#snippet spiritTable()}<div
		class="hidden min-h-0 min-w-0 flex-1 overflow-auto rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] md:block"
	>
		<table class="min-w-[2480px] table-fixed text-left text-xs">
			<thead class="bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)]">
				<tr>
					{#each spiritColumns as column (column.key)}<th
							class={`${column.width} sticky top-0 z-10 bg-[var(--color-surface-muted)] px-3 py-3`}
						>
							{@render spiritSortHeader(column.label, column.key)}
						</th>{/each}
					{#if canWrite}<th
							class="sticky top-0 z-10 w-32 bg-[var(--color-surface-muted)] px-3 py-3 text-right font-semibold"
							>Thao tác</th
						>{/if}
				</tr>
			</thead>
			<tbody class="divide-y divide-[var(--color-border)]">
				{#each sortedSpirits as item (item.id)}<tr class="hover:bg-[var(--color-primary-soft)]/40">
						{#each spiritColumns as column (column.key)}<td class="relative px-3 py-2 align-middle">
								{#if column.key === 'image_url'}
									{#if item.image_url}<button
											type="button"
											onclick={() => openSpiritImage(item)}
											class="cursor-pointer rounded focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-primary)]"
											aria-label={`Xem ảnh ${item.full_name}`}
											><img
												src={item.image_url}
												alt={item.full_name}
												class="h-14 w-10 rounded object-cover"
											/></button
										>{:else}<span
											class="grid h-14 w-10 place-items-center rounded bg-[var(--color-primary-soft)] font-semibold text-[var(--color-primary-dark)]"
											>{item.full_name.slice(0, 1)}</span
										>{/if}
								{:else if column.key === 'created_at' || column.key === 'updated_at'}<span
										class="whitespace-nowrap">{formatTimestamp(item[column.key])}</span
									>{:else if canWrite && editablePatchKeys.has(column.key)}<button
										type="button"
										onclick={() => beginCellEdit(item, column)}
										class="inline-block max-w-full cursor-pointer truncate rounded-sm text-left transition-colors hover:bg-[var(--color-primary-soft)]"
										title={item[column.key] || ''}>{item[column.key] || '—'}</button
									>{:else}<span
										class="inline-block max-w-full truncate"
										title={item[column.key] || ''}>{item[column.key] || '—'}</span
									>{/if}
								{#if cellEdit?.id === item.id && cellEdit.key === column.key}<form
										bind:this={cellEditRoot}
										onsubmit={saveCellPatch}
										class="absolute top-full left-2 z-30 w-72 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-3 shadow-[var(--shadow-popover)]"
									>
										<label class="block"
											><span class="mb-1 block text-xs font-semibold">{cellEdit.label}</span>
											{#if cellEdit.key === 'notes'}<textarea
													bind:this={cellEditField}
													bind:value={cellEdit.value}
													rows="3"
													class="w-full rounded-md border-[var(--color-border-strong)] text-sm"
												></textarea>{:else}<input
													bind:this={cellEditField}
													bind:value={cellEdit.value}
													required={cellEdit.key === 'full_name'}
													class="h-10 w-full rounded-md border-[var(--color-border-strong)] text-sm"
												/>{/if}</label
										>
										<div class="mt-3 flex justify-end gap-2">
											<button
												type="button"
												disabled={patchSaving}
												onclick={() => (cellEdit = null)}
												class="h-9 cursor-pointer rounded-md border border-[var(--color-border-strong)] px-3 text-xs font-semibold"
												>Hủy</button
											><button
												type="submit"
												disabled={patchSaving}
												class="h-9 cursor-pointer rounded-md bg-[var(--color-primary)] px-3 text-xs font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50"
												>{patchSaving ? 'Đang cập nhật...' : 'Cập nhật'}</button
											>
										</div>
									</form>{/if}
							</td>{/each}
						{#if canWrite}<td class="px-3 py-2 align-middle">
								<div class="flex justify-end gap-1">
									<button
										type="button"
										onclick={() => void edit(item)}
										class="grid h-8 w-8 cursor-pointer place-items-center rounded border border-[var(--color-border-strong)]"
										aria-label={`Sửa ${item.full_name}`}
										><span class="icon-[lucide--pencil] h-3.5 w-3.5" aria-hidden="true"
										></span></button
									>
									<button
										type="button"
										onclick={() => void remove(item)}
										class="grid h-8 w-8 cursor-pointer place-items-center rounded border border-[var(--color-danger)] text-[var(--color-danger)]"
										aria-label={`Xoá ${item.full_name}`}
										><span class="icon-[lucide--trash-2] h-3.5 w-3.5" aria-hidden="true"
										></span></button
									>
								</div>
							</td>{/if}
					</tr>{/each}
			</tbody>
		</table>
	</div>{/snippet}

{#snippet spiritSortHeader(label: string, key: SpiritSortKey)}<button
		type="button"
		onclick={() => sortSpirits(key)}
		class="flex w-full cursor-pointer items-center gap-1.5 font-semibold hover:text-[var(--color-text)]"
		>{label}<span
			class={[
				'h-3.5 w-3.5 shrink-0',
				spiritSortKey !== key
					? 'icon-[lucide--arrow-up-down] opacity-40'
					: spiritSortDirection === 'asc'
						? 'icon-[lucide--arrow-up] text-[var(--color-primary)]'
						: 'icon-[lucide--arrow-down] text-[var(--color-primary)]'
			]}
			aria-hidden="true"
		></span></button
	>{/snippet}
