<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import Lightbox from '$lib/ui/Lightbox.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import InlineSpiritEditor from '$lib/memorial/InlineSpiritEditor.svelte';
	import CardBaiVi from '$lib/memorial/CardBaiVi.svelte';
	import SpiritImageUploader from '$lib/memorial/SpiritImageUploader.svelte';
	import SpiritPortrait from '$lib/memorial/SpiritPortrait.svelte';
	import { memorialRevisionStore } from '$lib/memorial/memorial-revision-store.svelte';
	import { houseFilter } from '$lib/memorial/house-filter.svelte';
	import { uploadSpiritImage } from '$lib/uploads/api';
	import {
		createPositions,
		createTablet,
		createSpirits,
		bulkDeleteSpirits,
		deleteTablet,
		bulkPatchSpirits,
		deleteSpirit,
		downloadSpiritImportTemplate,
		exportSpiritsExcel,
		getOccupancy,
		importSpiritsFromExcel,
		listAreas,
		listHouses,
		listPositions,
		listSpirits,
		listSpiritPositionHistory,
		listTabletSpirits,
		listTablets,
		patchSpirit,
		previewSpiritImport,
		searchPositions,
		updateTablet,
		updateSpirit,
		type Area,
		type EditableSpiritInput,
		type House,
		type Occupancy,
		type InlineSpiritInput,
		type Position,
		type Spirit,
		type SpiritImportPreview,
		type SpiritImportResult,
		type SpiritInput,
		type SpiritPositionHistory,
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
		| 'tablet_image_url'
		| 'tablet_registered_at'
		| 'tablet_sender'
		| 'burial_place'
		| 'sender'
		| 'sent_month'
		| 'notes'
		| 'has_urn'
		| 'created_at'
		| 'updated_at'
	>;
	type SpiritColumn = { key: SpiritSortKey; label: string; defaultWidth: number };
	type SpiritTableGroup = {
		key: string;
		items: Spirit[];
		hasTablet: boolean;
		hasPosition: boolean;
	};
	type HighlightSegment = { text: string; match: boolean };
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
		{ key: 'image_url', label: 'Ảnh', defaultWidth: 48 },
		{ key: 'full_name', label: 'Họ tên', defaultWidth: 120 },
		{ key: 'dharma_name', label: 'Pháp danh', defaultWidth: 96 },
		{ key: 'birth_year', label: 'Năm sinh', defaultWidth: 64 },
		{ key: 'death_year', label: 'Năm mất', defaultWidth: 64 },
		{ key: 'age', label: 'Tuổi', defaultWidth: 48 },
		{ key: 'burial_place', label: 'Nơi an táng', defaultWidth: 144 },
		{ key: 'tablet_registered_at', label: 'Ngày đăng ký', defaultWidth: 104 },
		{ key: 'tablet_sender', label: 'Người gửi', defaultWidth: 128 },
		{ key: 'has_urn', label: 'Tro cốt/hài cốt', defaultWidth: 96 },
		{ key: 'notes', label: 'Ghi chú', defaultWidth: 160 }
	];
	const placementColumns: SpiritColumn[] = [
		{ key: 'tablet_image_url', label: 'Bài Vị', defaultWidth: 92 }
	];
	const columnWidthsStorageKey = 'nhalinh:spirit-table-column-widths:v3';
	const spiritSortKeys = new Set<SpiritSortKey>(['full_name', 'dharma_name', 'birth_year', 'death_year', 'age', 'sender', 'sent_month', 'has_urn']);
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
		urnStatus = $state<'' | 'yes' | 'no'>(''),
		tabletStatusFilter = $state<'' | 'pending' | 'enshrined' | 'taken_home'>(''),
		query = $state(''),
		total = $state(0),
		spiritGroupOffset = $state(0),
		loading = $state(true),
		spiritRequestVersion = $state(0),
		loadingMore = $state(false),
		hasMore = $state(false),
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
		hasPendingRelatedChanges = $state(false),
		patchSaving = $state(false),
		bulkActionOpen = $state(false),
		bulkPatchOpen = $state(false),
		bulkDeleteOpen = $state(false),
		bulkBusy = $state(false),
		bulkField = $state<EditablePatchKey>('sender'),
		bulkValue = $state(''),
		formOpen = $state(false),
		vacantPositionPickerOpen = $state(false),
		vacantPositionPickerLoading = $state(false),
		vacantPositionPicker = $state<Occupancy | null>(null),
		vacantPositionAreaId = $state(''),
		editing = $state<Spirit | null>(null),
		singleSpiritEntry = $state(false),
		spiritEditorVisible = $state(false),
		addingRelatedSpirit = $state(false),
		relatedTabletId = $state(''),
		relatedTabletName = $state(''),
		relatedSpirits = $state<Spirit[]>([]),
		relatedSpiritsLoading = $state(false),
		formHouseId = $state(''),
		formAreaId = $state(''),
		formAreas = $state<Area[]>([]),
		formPositionQuery = $state(''),
		formPositions = $state<Position[]>([]),
		formTablets = $state<Tablet[]>([]),
		selectedFormPosition = $state<Pick<Position, 'id' | 'name'> | null>(null),
		quickCreateTablet = $state(false),
		importInputKey = $state(0),
		timer: ReturnType<typeof setTimeout> | undefined,
		positionTimer: ReturnType<typeof setTimeout> | undefined,
		positionRequest = 0,
		tabletRequest = 0,
		relatedSpiritsRequest = 0;
	let columnWidths = $state<Record<string, number>>({});
	let form = $state<SpiritInput>(emptyForm());
	let relatedSpiritForm = $state<SpiritInput>(emptyForm());
	let newSpirits = $state<EditableSpiritInput[]>([emptyInlineSpirit()]);
	let cellEditRoot = $state<HTMLFormElement>();
	let cellEditField = $state<HTMLInputElement | HTMLTextAreaElement>();
	let cellEdit = $state<{ id: string; key: EditablePatchKey; label: string; value: string } | null>(
		null
	);
	let selectedSpiritIDs = $state<Set<string>>(new Set());
	let deleteConfirmation = $state('');
	let spiritHistory = $state<SpiritPositionHistory[]>([]);
	let spiritHistoryOpen = $state(false);
	let spiritHistoryName = $state('');
	let newTabletForm = $state({ name: '', code: '', registered_at: '', enshrined_at: '', entered_worship_area_at: '', status: 'pending' as Tablet['status'], type: 'spirit' as Tablet['type'], image_url: '', sender: '', notes: '' });
	let selectedHouse = $derived(houses.find((v) => v.id === houseId));
	let selectedFormTablet = $derived(formTablets.find((tablet) => tablet.id === form.tablet_id));
	let creatingTablet = $derived(singleSpiritEntry && !form.tablet_id);
	let hasNewTabletSpirit = $derived(newSpirits.length > 0);
	let showSpiritEditor = $derived(!relatedTabletId || spiritEditorVisible);
	let vacantPickerPositions = $derived(
		(vacantPositionPicker?.positions ?? []).filter(
			(position) => position.area_id === vacantPositionAreaId
		)
	);
	let vacantPickerMaxRow = $derived(
		Math.max(0, ...vacantPickerPositions.map((position) => position.row_number))
	);
	let vacantPickerMaxColumn = $derived(
		Math.max(0, ...vacantPickerPositions.map((position) => position.column_number))
	);
	let vacantPickerRows = $derived(
		Array.from({ length: vacantPickerMaxRow }, (_, index) => index + 1)
	);
	let vacantPickerColumns = $derived(
		Array.from({ length: vacantPickerMaxColumn }, (_, index) => index + 1)
	);
	let vacantPickerByCoordinate = $derived(
		new Map(
			vacantPickerPositions.map((position) => [
				`${position.row_number}:${position.column_number}`,
				position
			])
		)
	);
	let canWrite = $derived(
		authStore.user?.role === 'admin' || selectedHouse?.access_role === 'editor'
	);
	let tableSpiritColumns = $derived([
		...placementColumns,
		...spiritColumns
	]);
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
	let filteredSpirits = $derived(
		tabletStatusFilter ? spirits.filter((item) => item.tablet_status === tabletStatusFilter) : spirits
	);
	let tableSpiritGroups = $derived.by(() => {
		const groups = new Map<string, SpiritTableGroup>();
		for (const item of filteredSpirits) {
			const hasTablet = Boolean(item.tablet_id);
			const hasPosition = Boolean(item.position_id);
			const key = hasTablet ? `tablet:${item.tablet_id}` : `unplaced:${item.id}`;
			const group = groups.get(key) ?? { key, items: [], hasTablet, hasPosition };
			group.items.push(item);
			groups.set(key, group);
		}
		const direction = spiritSortDirection === 'asc' ? 1 : -1;
		const searchQuery = foldSearchText(query);
		const compare = (left: Spirit, right: Spirit) => {
			const leftRank = spiritNameSearchRank(left.full_name, searchQuery);
			const rightRank = spiritNameSearchRank(right.full_name, searchQuery);
			if (leftRank !== rightRank) return leftRank - rightRank;
			const result = vietnameseCollator.compare(
				String(left[spiritSortKey] ?? ''),
				String(right[spiritSortKey] ?? '')
			);
			return (result || vietnameseCollator.compare(left.full_name, right.full_name)) * direction;
		};
		const placementSort = new Set<SpiritSortKey>(['house_name', 'position_name', 'tablet_name']);
		for (const group of groups.values()) group.items.sort(compare);
		const items = [...groups.values()];
		items.sort((left, right) => {
			if (searchQuery || placementSort.has(spiritSortKey))
				return compare(left.items[0], right.items[0]);
			return vietnameseCollator.compare(
				left.items[0].position_name || 'ZZZ',
				right.items[0].position_name || 'ZZZ'
			);
		});
		return items;
	});
	let tableWidth = $derived(
		(canWrite ? 96 : 0) +
			tableSpiritColumns.reduce((total, column) => total + columnWidth(column), 0)
	);
	let selectedCount = $derived(selectedSpiritIDs.size);

	onMount(() => {
		restoreTablePreferences();
		const onHouseSelect = (event: Event) => {
			const nextHouseID = (event as CustomEvent<string>).detail;
			if (nextHouseID === houseId) return;
			houseId = nextHouseID;
			void changeHouse();
		};
		window.addEventListener('memorial-house-select', onHouseSelect);
		document.addEventListener('pointerdown', dismissCellEditOnOutsideClick);
		document.addEventListener('pointerdown', dismissBulkActionsOnOutsideClick);
		void initialize();
		return () => {
			window.removeEventListener('memorial-house-select', onHouseSelect);
			document.removeEventListener('pointerdown', dismissCellEditOnOutsideClick);
			document.removeEventListener('pointerdown', dismissBulkActionsOnOutsideClick);
			if (timer) clearTimeout(timer);
			if (positionTimer) clearTimeout(positionTimer);
		};
	});
	function dismissCellEditOnOutsideClick(event: PointerEvent) {
		if (cellEdit && event.target instanceof Node && !cellEditRoot?.contains(event.target)) {
			cellEdit = null;
		}
	}
	function dismissBulkActionsOnOutsideClick(event: PointerEvent) {
		if (
			bulkActionOpen &&
			event.target instanceof Element &&
			!event.target.closest('[data-bulk-actions]')
		) {
			bulkActionOpen = false;
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
			const savedColumnWidths = JSON.parse(localStorage.getItem(columnWidthsStorageKey) ?? '{}');
			if (savedColumnWidths && typeof savedColumnWidths === 'object') {
				columnWidths = Object.fromEntries(
					Object.entries(savedColumnWidths).filter(
						([, width]) => typeof width === 'number' && Number.isFinite(width) && width >= 56
					)
				) as Record<string, number>;
			}
		} catch {
			// Keep defaults when browser storage is unavailable.
		}
	}
	function columnWidth(column: SpiritColumn) {
		return columnWidths[column.key] ?? column.defaultWidth;
	}
	function beginColumnResize(event: PointerEvent, column: SpiritColumn) {
		event.preventDefault();
		event.stopPropagation();
		const handle = event.currentTarget;
		if (handle instanceof HTMLElement) handle.setPointerCapture(event.pointerId);
		const header = handle instanceof HTMLElement ? handle.closest('th') : null;
		const update = (moveEvent: PointerEvent) => {
			const left = header?.getBoundingClientRect().left;
			const width = left === undefined ? columnWidth(column) : moveEvent.clientX - left;
			columnWidths = {
				...columnWidths,
				[column.key]: Math.max(56, Math.round(width))
			};
		};
		const finish = () => {
			window.removeEventListener('pointermove', update);
			window.removeEventListener('pointerup', finish);
			window.removeEventListener('pointercancel', finish);
			if (handle instanceof HTMLElement && handle.hasPointerCapture(event.pointerId)) {
				handle.releasePointerCapture(event.pointerId);
			}
			try {
				localStorage.setItem(columnWidthsStorageKey, JSON.stringify(columnWidths));
			} catch {
				// Widths still apply for the current session when storage is unavailable.
			}
		};
		window.addEventListener('pointermove', update);
		window.addEventListener('pointerup', finish, { once: true });
		window.addEventListener('pointercancel', finish, { once: true });
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
	function tabletDisplayImage(items: Spirit[]) {
		return items[0]?.tablet_image_url || items.find((item) => item.image_url)?.image_url || '';
	}
	function tabletSavedImage(imageUrl: string, items: Pick<Spirit, 'image_url'>[]) {
		return imageUrl || (items.length === 1 ? items[0].image_url : '');
	}
	function formatTimestamp(value: string) {
		return value ? new Intl.DateTimeFormat('vi-VN').format(new Date(value)) : '—';
	}
	function foldSearchText(value: string) {
		return foldSearchCharacters(value).trim();
	}
	function foldSearchCharacters(value: string) {
		return value
			.normalize('NFD')
			.replace(/[\u0300-\u036f]/g, '')
			.toLowerCase()
			.replace(/đ/g, 'd')
			.replace(/-/g, '');
	}
	function spiritNameSearchRank(name: string, searchQuery: string) {
		if (!searchQuery) return 0;
		const foldedName = foldSearchText(name);
		if (foldedName === searchQuery) return 0;
		if (foldedName.startsWith(searchQuery)) return 1;
		if (foldedName.includes(searchQuery)) return 2;
		return 3;
	}
	function highlightSegments(value: string, rawQuery: string): HighlightSegment[] {
		const searchQuery = foldSearchText(rawQuery);
		if (!value || !searchQuery) return [{ text: value, match: false }];
		let folded = '';
		const starts: number[] = [];
		const ends: number[] = [];
		for (let offset = 0; offset < value.length;) {
			const codePoint = value.codePointAt(offset);
			if (codePoint === undefined) break;
			const char = String.fromCodePoint(codePoint);
			const end = offset + char.length;
			const normalized = foldSearchCharacters(char);
			for (const unit of normalized) {
				folded += unit;
				starts.push(offset);
				ends.push(end);
			}
			offset = end;
		}
		const segments: HighlightSegment[] = [];
		let originalOffset = 0;
		let matchIndex = folded.indexOf(searchQuery);
		while (matchIndex >= 0) {
			const start = starts[matchIndex];
			const end = ends[matchIndex + searchQuery.length - 1];
			if (start > originalOffset)
				segments.push({ text: value.slice(originalOffset, start), match: false });
			if (end > originalOffset)
				segments.push({ text: value.slice(Math.max(start, originalOffset), end), match: true });
			originalOffset = end;
			matchIndex = folded.indexOf(searchQuery, matchIndex + searchQuery.length);
		}
		if (originalOffset < value.length)
			segments.push({ text: value.slice(originalOffset), match: false });
		return segments.length ? segments : [{ text: value, match: false }];
	}
	function beginCellEdit(item: Spirit, column: SpiritColumn) {
		if (!canWrite || !editablePatchKeys.has(column.key)) return;
		cellEdit = {
			id: item.id,
			key: column.key as EditablePatchKey,
			label: column.label,
			value: String(item[column.key] ?? '')
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
			memorialRevisionStore.invalidate();
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
			houseId = houseFilter.id || houses[0]?.id || '';
			if (!houseFilter.id) houseFilter.id = houseId;
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
		const requestVersion = ++spiritRequestVersion;
		const selectedHouseID = houseId;
		const selectedAreaID = areaId;
		const selectedUrnStatus = urnStatus;
		const selectedQuery = query;
		const page = await listSpirits(selectedQuery, selectedHouseID, selectedAreaID, 24, 0, selectedUrnStatus);
		if (
			requestVersion !== spiritRequestVersion ||
			selectedHouseID !== houseId ||
			selectedAreaID !== areaId ||
			selectedUrnStatus !== urnStatus ||
			selectedQuery !== query
		)
			return;
		spirits = page.spirits;
		total = page.total;
		spiritGroupOffset = page.next_offset;
		hasMore = page.has_more;
		selectedSpiritIDs = new Set();
		bulkActionOpen = false;
	}
	async function loadMore() {
		if (loadingMore || !hasMore) return;
		loadingMore = true;
		try {
			const page = await listSpirits(query, houseId, areaId, 24, spiritGroupOffset, urnStatus);
			const known = new Set(spirits.map((item) => item.id));
			spirits = [...spirits, ...page.spirits.filter((item) => !known.has(item.id))];
			spiritGroupOffset = page.next_offset;
			hasMore = page.has_more;
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			loadingMore = false;
		}
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
			memorialRevisionStore.invalidate();
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
		singleSpiritEntry = true;
		spiritEditorVisible = false;
		addingRelatedSpirit = false;
		relatedTabletId = '';
		relatedTabletName = '';
		relatedSpirits = [];
		relatedSpiritsRequest++;
		form = emptyForm();
		newTabletForm = { name: '', code: '', registered_at: '', enshrined_at: '', entered_worship_area_at: '', status: 'pending', type: 'spirit', image_url: '', sender: '', notes: '' };
		newSpirits = [];
		inlineEditorBusy = false;
		hasPendingRelatedChanges = false;
		formHouseId = preferredFormHouse();
		form.house_id = formHouseId;
		formAreaId = '';
		formAreas = [];
		void loadFormAreas();
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
	function fillEditingForm(item: Spirit) {
		editing = item;
		formHouseId = item.house_id;
		formAreaId = item.area_id;
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
			notes: item.notes,
			has_urn: item.has_urn
		};
		formPositionQuery = item.position_name;
		selectedFormPosition = item.position_id
			? { id: item.position_id, name: item.position_name }
			: null;
		quickCreateTablet = false;
	}
	async function loadRelatedSpirits(tabletId: string) {
		const request = ++relatedSpiritsRequest;
		if (!tabletId) {
			relatedSpirits = [];
			relatedSpiritsLoading = false;
			return;
		}
		relatedSpiritsLoading = true;
		try {
			const items = await listTabletSpirits(tabletId);
			if (request === relatedSpiritsRequest) {
				relatedSpirits = items.sort((left, right) =>
					vietnameseCollator.compare(left.full_name, right.full_name)
				);
				if (relatedSpirits[0]) fillEditingForm(relatedSpirits[0]);
			}
		} finally {
			if (request === relatedSpiritsRequest) relatedSpiritsLoading = false;
		}
	}
	async function edit(item: Spirit) {
		fillEditingForm(item);
		hasPendingRelatedChanges = false;
		singleSpiritEntry = false;
		spiritEditorVisible = false;
		addingRelatedSpirit = false;
		relatedTabletId = item.tablet_id;
		relatedTabletName = item.tablet_name;
		relatedSpirits = [];
		formPositions = [];
		formTablets = [];
		formOpen = true;
		void loadFormAreas();
		void loadRelatedSpirits(item.tablet_id);
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
	function selectRelatedSpirit(item: Spirit) {
		if (!canWrite || saving) return;
		spiritEditorVisible = true;
		addingRelatedSpirit = false;
		if (item.id === editing?.id) return;
		fillEditingForm(item);
	}
	async function removeRelatedSpirit(item: Spirit) {
		if (!canWrite || saving) return;
		if (relatedSpirits.length <= 1) {
			toastStore.error(
				'Bài vị cần tối thiểu một Hương linh. Hãy xóa Bài vị nếu muốn xóa bản ghi cuối cùng.'
			);
			return;
		}
		const ok = await popupStore.confirm({
			title: 'Xóa Hương linh?',
			message: `Thông tin của ${item.full_name} sẽ được xóa mềm và không còn xuất hiện trong danh sách.`,
			confirmLabel: 'Xóa',
			tone: 'danger'
		});
		if (!ok) return;
		try {
			await deleteSpirit(item.id);
			relatedSpirits = relatedSpirits.filter((related) => related.id !== item.id);
			if (editing?.id === item.id && relatedSpirits[0]) fillEditingForm(relatedSpirits[0]);
			memorialRevisionStore.invalidate();
			toastStore.success('Đã xóa Hương linh');
			await load();
		} catch (e) {
			toastStore.error(message(e));
		}
	}
	function addRelatedSpirit() {
		const source = editing ?? relatedSpirits[0];
		if (!source) return;
		spiritEditorVisible = true;
		addingRelatedSpirit = true;
		relatedSpiritForm = {
			...emptyForm(),
			house_id: source.house_id,
			tablet_id: source.tablet_id
		};
	}
	async function updateRelatedSpiritList() {
		if (!relatedSpiritForm.full_name.trim()) {
			toastStore.error('Họ tên không được để trống');
			return;
		}
		saving = true;
		try {
			const [created] = await createSpirits([relatedSpiritForm]);
			if (!created) throw new Error('Không thể thêm Hương linh');
			relatedSpirits = [...relatedSpirits, created].sort((left, right) =>
				vietnameseCollator.compare(left.full_name, right.full_name)
			);
			addingRelatedSpirit = false;
			spiritEditorVisible = false;
			toastStore.success('Đã thêm Hương linh vào danh sách');
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			saving = false;
		}
	}
	function addSpiritToNewTablet() {
		if (!selectedFormPosition) {
			toastStore.error('Vui lòng chọn vị trí trước');
			return;
		}
		if (!form.full_name.trim()) {
			toastStore.error('Họ tên không được để trống');
			return;
		}
		newSpirits = [...newSpirits, toInlineSpirit(form)];
		form = { ...emptyForm(), house_id: formHouseId };
		toastStore.success('Đã thêm Hương linh vào danh sách cùng bài vị');
	}
	function removeNewTabletSpirit(index: number) {
		newSpirits = newSpirits.filter((_, itemIndex) => itemIndex !== index);
	}
	function updateEditedSpiritInRelatedList() {
		if (!editing) return;
		if (!form.full_name.trim()) {
			toastStore.error('Họ tên không được để trống');
			return;
		}
		const current = relatedSpirits.find((spirit) => spirit.id === editing?.id);
		if (!current) return;
		const updated = { ...current, ...form };
		relatedSpirits = relatedSpirits.map((spirit) => (spirit.id === updated.id ? updated : spirit));
		editing = updated;
		hasPendingRelatedChanges = true;
		toastStore.success('Đã cập nhập Hương linh vào danh sách');
	}
	function markEditingFormChanged() {
		if (editing && !saving) hasPendingRelatedChanges = true;
	}
	function openSpiritEditorFromRow(event: MouseEvent, item: Spirit) {
		if (!canWrite) return;
		if (
			event.target instanceof Element &&
			event.target.closest('button, input, select, textarea, label')
		) {
			return;
		}
		cellEdit = null;
		void edit(item);
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
		formAreaId = '';
		formAreas = [];
		void loadFormAreas();
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
	async function loadFormAreas() {
		if (!formHouseId) return;
		try {
			formAreas = await listAreas(formHouseId);
		} catch (error) {
			toastStore.error(message(error));
		}
	}
	function changeFormArea() {
		formPositionQuery = '';
		formPositions = [];
		selectedFormPosition = null;
		if (!editing) {
			formTablets = [];
			form.tablet_id = '';
		}
		quickCreateTablet = false;
		positionRequest++;
		tabletRequest++;
		positionLoading = false;
		tabletLoading = false;
	}
	async function openVacantPositionPicker() {
		if (!formHouseId) {
			toastStore.error('Vui lòng chọn Nhà Linh trước');
			return;
		}
		vacantPositionPickerOpen = true;
		vacantPositionPickerLoading = true;
		try {
			const occupancy = await getOccupancy(formHouseId);
			vacantPositionPicker = occupancy;
			vacantPositionAreaId =
				formAreaId && occupancy.areas.some((area) => area.id === formAreaId)
					? formAreaId
					: (occupancy.areas[0]?.id ?? '');
		} catch (error) {
			toastStore.error(message(error));
			vacantPositionPickerOpen = false;
		} finally {
			vacantPositionPickerLoading = false;
		}
	}
	async function selectVacantPosition(
		position: Position | undefined,
		rowNumber: number,
		columnNumber: number
	) {
		if ((position?.tablet_count ?? 0) > 0) return;
		let selected = position;
		if (!selected) {
			vacantPositionPickerLoading = true;
			try {
				const created = await createPositions(vacantPositionAreaId, [
					{ row_number: rowNumber, column_number: columnNumber, notes: '' }
				]);
				selected = created.positions[0];
				if (!selected) {
					const occupancy = await getOccupancy(formHouseId);
					selected = occupancy.positions.find(
						(item) =>
							item.area_id === vacantPositionAreaId &&
							item.row_number === rowNumber &&
							item.column_number === columnNumber
					);
				}
				if (!selected) {
					toastStore.error('Không thể tạo vị trí đã chọn');
					return;
				}
			} catch (error) {
				toastStore.error(message(error));
				return;
			} finally {
				vacantPositionPickerLoading = false;
			}
		}

		const tabletToMove =
			editing && selectedFormTablet
				? {
						...selectedFormTablet,
						position_id: selected.id,
						area_id: selected.area_id,
						area_code: selected.area_code,
						position_name: selected.name,
						row_number: selected.row_number,
						column_number: selected.column_number
					}
				: null;
		formAreaId = selected.area_id;
		selectedFormPosition = selected;
		formPositionQuery = selected.name;
		formPositions = [];
		formTablets = tabletToMove ? [tabletToMove] : [];
		form.tablet_id = tabletToMove?.id ?? '';
		quickCreateTablet = false;
		tabletRequest++;
		tabletLoading = false;
		vacantPositionPickerOpen = false;
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
			const items = formAreaId
				? (await listPositions(formAreaId)).filter((position) => {
						const queryValue = formPositionQuery.trim().toLocaleLowerCase('vi-VN');
						return (
							position.name.toLocaleLowerCase('vi-VN').includes(queryValue) ||
							position.area_code.toLocaleLowerCase('vi-VN').includes(queryValue)
						);
					})
				: await searchPositions(formHouseId, formPositionQuery);
			if (request === positionRequest) formPositions = items;
		} finally {
			if (request === positionRequest) positionLoading = false;
		}
	}
	async function selectFormPosition(position: Position) {
		if (
			(singleSpiritEntry || editing) &&
			position.tablet_count > 0 &&
			position.id !== editing?.position_id
		) {
			toastStore.error('Vị trí này đã có Bài vị. Vui lòng chọn vị trí trống khác');
			return;
		}
		const tabletToMove =
			editing && selectedFormTablet
				? {
						...selectedFormTablet,
						position_id: position.id,
						area_id: position.area_id,
						area_code: position.area_code,
						position_name: position.name,
						row_number: position.row_number,
						column_number: position.column_number
					}
				: null;
		positionRequest++;
		selectedFormPosition = position;
		formAreaId = position.area_id;
		formPositionQuery = position.name;
		form.tablet_id = tabletToMove?.id ?? '';
		quickCreateTablet = false;
		formTablets = tabletToMove ? [tabletToMove] : [];
		if (singleSpiritEntry || tabletToMove) return;
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
	async function save(e?: SubmitEvent) {
		e?.preventDefault();
		saving = true;
		try {
			if (editing) {
				if (selectedFormTablet) {
					if (relatedSpiritsLoading || relatedSpirits.length === 0) {
						throw new Error('Đang tải danh sách Hương linh của Bài vị. Vui lòng thử lại sau.');
					}
					await updateTablet(selectedFormTablet.id, {
						position_id: selectedFormTablet.position_id,
						name: selectedFormTablet.name,
						code: '', registered_at: form.sent_month.trim() || selectedFormTablet.registered_at, enshrined_at: selectedFormTablet.enshrined_at, entered_worship_area_at: selectedFormTablet.entered_worship_area_at, status: selectedFormTablet.status, type: selectedFormTablet.type,
						image_url: tabletSavedImage(selectedFormTablet.image_url, relatedSpirits.map((spirit) => spirit.id === editing?.id ? form : spirit)),
						sender: selectedFormTablet.sender,
						notes: selectedFormTablet.notes,
						spirits: relatedSpirits.map((spirit) => {
							const source = spirit.id === editing?.id ? form : spirit;
							return {
								id: spirit.id,
								full_name: source.full_name,
								dharma_name: source.dharma_name,
								birth_year: source.birth_year,
								death_year: source.death_year,
								age: source.age,
								image_url: source.image_url,
								burial_place: source.burial_place,
								sender: source.sender,
								sent_month: source.sent_month,
								notes: source.notes,
								has_urn: source.has_urn
							};
						})
					});
				} else {
					await updateSpirit(editing.id, form);
				}
				if (addingRelatedSpirit) {
					if (!relatedSpiritForm.full_name.trim()) {
						throw new Error('Họ tên không được để trống');
					}
					await createSpirits([relatedSpiritForm]);
				}
			} else if (singleSpiritEntry) {
				if (
					!selectedFormPosition &&
					Object.values(newTabletForm).some((value) => value.trim() !== '')
				) {
					throw new Error('Vui lòng chọn vị trí trước khi tạo Bài vị');
				}
				if (selectedFormPosition && !form.tablet_id) {
					if (newSpirits.length === 0) {
						throw new Error('Hãy thêm ít nhất một Hương linh vào danh sách trước khi lưu');
					}
					await createTablet({
						position_id: selectedFormPosition.id,
						name: newSpirits[0].full_name.trim(),
						code: '', registered_at: newSpirits.find((spirit) => spirit.sent_month.trim())?.sent_month ?? newTabletForm.registered_at, enshrined_at: newTabletForm.enshrined_at, entered_worship_area_at: newTabletForm.entered_worship_area_at, status: newTabletForm.status, type: newTabletForm.type,
						image_url: tabletSavedImage(newTabletForm.image_url, newSpirits),
						sender: newTabletForm.sender || newSpirits[0]?.sender || '',
						notes: newTabletForm.notes,
						spirits: newSpirits
					});
				} else {
					if (!form.full_name.trim()) {
						throw new Error('Họ tên không được để trống');
					}
					await createSpirits([
						{
							...form,
							house_id: formHouseId,
							tablet_id: form.tablet_id
						}
					]);
				}
			} else if (quickCreateTablet && selectedFormPosition) {
				const rows = validNewSpiritRows();
				await createTablet({
					position_id: selectedFormPosition.id,
					name: rows[0].full_name,
					registered_at: rows.find((spirit) => spirit.sent_month.trim())?.sent_month ?? '',
					image_url: tabletSavedImage('', rows),
					sender: '',
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
						tablet_id: form.tablet_id,
						has_urn: false
					}))
				);
			}
			memorialRevisionStore.invalidate();
			toastStore.success(
				addingRelatedSpirit || singleSpiritEntry
					? 'Đã thêm Hương linh'
					: quickCreateTablet
						? 'Đã tạo bài vị và thêm Hương linh'
						: editing
							? 'Đã cập nhật Hương linh'
							: `Đã thêm ${newSpirits.filter(hasSpiritData).length} Hương linh`
			);
			hasPendingRelatedChanges = false;
			formOpen = false;
			await load();
		} catch (err) {
			toastStore.error(message(err));
		} finally {
			saving = false;
			quickCreateTablet = false;
			singleSpiritEntry = false;
			addingRelatedSpirit = false;
		}
	}
	async function requestCloseForm() {
		if (saving || imageUploading) return;
		if (!hasPendingRelatedChanges) {
			formOpen = false;
			return;
		}
		const shouldSave = await popupStore.confirm({
			title: 'Lưu thông tin đã cập nhập?',
			message: 'Bạn có muốn Lưu thông tin đã cập nhập không?',
			confirmLabel: 'Lưu',
			cancelLabel: 'Hủy'
		});
		if (shouldSave) {
			await save();
			return;
		}
		hasPendingRelatedChanges = false;
		formOpen = false;
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
			markEditingFormChanged();
			toastStore.success('Đã tải ảnh Hương linh');
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			imageUploading = false;
		}
	}
	async function selectRelatedSpiritImage(file: File) {
		imageUploading = true;
		try {
			relatedSpiritForm.image_url = await uploadSpiritImage(file);
			toastStore.success('Đã tải ảnh Hương linh');
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			imageUploading = false;
		}
	}
	async function selectTabletImage(file: File) {
		if (!selectedFormTablet) return;
		imageUploading = true;
		try {
			selectedFormTablet.image_url = await uploadSpiritImage(file);
			markEditingFormChanged();
			toastStore.success('Đã tải ảnh Bài vị');
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			imageUploading = false;
		}
	}
	async function selectNewTabletImage(file: File) {
		imageUploading = true;
		try {
			newTabletForm.image_url = await uploadSpiritImage(file);
			toastStore.success('Đã tải ảnh Bài vị');
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			imageUploading = false;
		}
	}
	async function remove(item: Spirit) {
		const ok = await popupStore.confirm({
			title: 'Ẩn Hương linh?',
			message: `Thông tin của ${item.full_name} sẽ được xóa mềm và không còn xuất hiện trong danh sách.`,
			confirmLabel: 'Xoá',
			tone: 'danger'
		});
		if (!ok) return;
		try {
			await deleteSpirit(item.id);
			memorialRevisionStore.invalidate();
			toastStore.success('Đã xoá Hương linh');
			await load();
		} catch (e) {
			toastStore.error(message(e));
		}
	}
	async function removeTablet(item: Spirit) {
		if (!item.tablet_id) {
			await remove(item);
			return;
		}
		const ok = await popupStore.confirm({
			title: 'Xóa Bài vị?',
			message: `Bài vị ${item.tablet_name || item.position_name} sẽ bị xóa. Các Hương linh thuộc Bài vị sẽ chuyển về danh sách chưa xếp.`,
			confirmLabel: 'Xóa Bài vị',
			tone: 'danger'
		});
		if (!ok) return;
		try {
			await deleteTablet(item.tablet_id);
			memorialRevisionStore.invalidate();
			toastStore.success('Đã xóa Bài vị; Hương linh được chuyển về chưa xếp');
			await load();
		} catch (e) {
			toastStore.error(message(e));
		}
	}
	function toggleSpiritSelection(id: string) {
		const next = new Set(selectedSpiritIDs);
		next.has(id) ? next.delete(id) : next.add(id);
		selectedSpiritIDs = next;
	}
	function toggleVisibleSelection() {
		const ids = spirits.map((item) => item.id);
		const allSelected = ids.length > 0 && ids.every((id) => selectedSpiritIDs.has(id));
		selectedSpiritIDs = allSelected ? new Set() : new Set(ids);
	}
	function clearSpiritSelection() {
		selectedSpiritIDs = new Set();
		bulkActionOpen = false;
	}
	function selectAllVisibleSpirits() {
		selectedSpiritIDs = new Set(spirits.map((item) => item.id));
	}
	async function applyBulkPatch() {
		bulkBusy = true;
		try {
			await bulkPatchSpirits([...selectedSpiritIDs], bulkField, bulkValue);
			memorialRevisionStore.invalidate();
			toastStore.success(`Đã cập nhật ${selectedCount} Hương linh`);
			bulkPatchOpen = false;
			clearSpiritSelection();
			await load();
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			bulkBusy = false;
		}
	}
	async function applyBulkDelete() {
		bulkBusy = true;
		try {
			await bulkDeleteSpirits([...selectedSpiritIDs]);
			memorialRevisionStore.invalidate();
			toastStore.success(`Đã xóa mềm ${selectedCount} Hương linh`);
			bulkDeleteOpen = false;
			deleteConfirmation = '';
			clearSpiritSelection();
			await load();
		} catch (e) {
			toastStore.error(message(e));
		} finally {
			bulkBusy = false;
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
			familiar_name: '',
			gender: '',
			birth_date: '',
			death_date: '',
			birth_lunar: '',
			death_lunar: '',
			status: 'draft',
			entered_worship_area_at: '',
			enshrined_at: '',
			age: '',
			image_url: '',
			burial_place: '',
			sender: '',
			sent_month: '',
			notes: '',
			has_urn: false
		};
	}
	function message(e: unknown) {
		return e instanceof Error ? e.message : 'Có lỗi xảy ra';
	}
	async function showSpiritHistory(item: Spirit) {
		spiritHistoryName = item.full_name;
		spiritHistory = [];
		spiritHistoryOpen = true;
		try { spiritHistory = await listSpiritPositionHistory(item.id); }
		catch (error) { toastStore.error(message(error)); }
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
		if (bulkActionOpen) {
			bulkActionOpen = false;
			return;
		}
		if (bulkPatchOpen && !bulkBusy) {
			bulkPatchOpen = false;
			return;
		}
		if (bulkDeleteOpen && !bulkBusy) {
			bulkDeleteOpen = false;
			return;
		}
		if (document.querySelector('[role="dialog"][aria-modal="true"]')) return;
		if (formOpen) {
			void requestCloseForm();
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
	<div class="mx-auto grid max-w-[1320px] gap-2 md:grid-cols-[180px_150px_160px_1fr_auto]">
			<select
				bind:value={areaId}
				onchange={() => void changeArea()}
				aria-label="Chọn khu vực"
				class="h-11 rounded-md border-[var(--color-border-strong)] text-sm"
				><option value="">Tất cả khu vực</option>{#each areas as a (a.id)}<option value={a.id}
						>Khu {a.code}{a.name ? ` – ${a.name}` : ''}</option
					>{/each}</select
			>
			<select
				bind:value={urnStatus}
				onchange={() => void load()}
				aria-label="Lọc theo Hũ cốt"
				class="h-11 rounded-md border-[var(--color-border-strong)] text-sm"
			><option value="">Tất cả Hũ cốt</option><option value="yes">Có Hũ cốt</option><option value="no">Chưa có Hũ cốt</option></select
			>
			<select bind:value={tabletStatusFilter} aria-label="Lọc theo trạng thái Bài vị" class="h-11 rounded-md border-[var(--color-border-strong)] text-sm"><option value="">Tất cả trạng thái</option><option value="pending">Chờ an vị</option><option value="enshrined">Đã an vị</option><option value="taken_home">Đã thỉnh về</option></select>
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
					{#if canWrite && spirits.length > 0 && selectedCount === 0}<button
							type="button"
							onclick={selectAllVisibleSpirits}
							class="h-9 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-xs font-semibold hover:bg-[var(--color-surface-muted)]"
							>Chọn tất cả</button
						>{/if}
					{#if canWrite && selectedCount > 0}<div class="relative" data-bulk-actions>
							<button
								type="button"
								onclick={() => (bulkActionOpen = !bulkActionOpen)}
								class="h-9 rounded-md bg-[var(--color-primary)] px-3 text-xs font-semibold text-white"
							>
								Đã chọn {selectedCount} · Thao tác
							</button>
							{#if bulkActionOpen}<div
									class="absolute top-full right-0 z-30 mt-1 w-48 rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] p-1 shadow-[var(--shadow-popover)]"
								>
									<button
										type="button"
										onclick={() => {
											bulkActionOpen = false;
											bulkPatchOpen = true;
										}}
										class="flex w-full items-center gap-2 rounded px-3 py-2 text-left text-sm hover:bg-[var(--color-surface-muted)]"
										><span class="icon-[lucide--pencil-line] h-4 w-4"></span>Cập nhật thông tin</button
									>
									<button
										type="button"
										onclick={() => {
											bulkActionOpen = false;
											bulkDeleteOpen = true;
											deleteConfirmation = '';
										}}
										class="flex w-full items-center gap-2 rounded px-3 py-2 text-left text-sm text-[var(--color-danger)] hover:bg-[var(--color-danger-soft)]"
										><span class="icon-[lucide--trash-2] h-4 w-4"></span>Xóa Hương linh</button
									>
								</div>{/if}
						</div>
						<button
							type="button"
							onclick={clearSpiritSelection}
							class="h-9 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-xs font-semibold"
							>Bỏ chọn</button
						>{/if}
					<button
						type="button"
						onclick={openExportPopup}
						class="h-9 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-xs font-semibold hover:bg-[var(--color-surface-muted)]"
					>
						<span class="mr-1 icon-[lucide--download] inline-block h-4 w-4 align-text-bottom"
						></span>
						Export Excel
					</button>
					{#if canWrite}
						<button
							type="button"
							onclick={openImportPopup}
							disabled={houses.length === 0}
							class="h-9 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-3 text-xs font-semibold hover:bg-[var(--color-surface-muted)] disabled:opacity-50"
						>
							<span class="mr-1 icon-[lucide--upload] inline-block h-4 w-4 align-text-bottom"
							></span>
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
							Hãy tạo bài vị trong mục Bài vị trước.
						</p>{/if}
				</div>
			{:else if desktopView === 'table'}<div class="min-h-0 flex-1 overflow-y-auto md:hidden">
					{@render spiritCardGroups()}
					{@render loadMoreButton()}
				</div>
				{@render spiritTable()}
	{:else}<div class="min-h-0 flex-1 overflow-y-auto">
					{@render tabletCardGrid()}
					{@render loadMoreButton()}
				</div>{/if}
		</div>
	</div>
</section>

<Lightbox src={lightboxSrc} alt={lightboxAlt} bind:open={lightboxOpen} />

<Popup open={spiritHistoryOpen} title={`Lịch sử vị trí · ${spiritHistoryName}`} sizeClass="max-w-xl" onClose={() => spiritHistoryOpen = false}>
	{#if spiritHistory.length === 0}<p class="text-sm text-[var(--color-text-secondary)]">Chưa có lịch sử vị trí.</p>{:else}<div class="space-y-3">{#each spiritHistory as item (item.id)}<div class="rounded-md border border-[var(--color-border)] p-3"><div class="flex justify-between gap-3"><strong>{item.change_type === 'placed' ? 'Đã xếp vị trí' : item.change_type === 'moved' ? 'Đã chuyển vị trí' : item.change_type === 'unplaced' ? 'Đã gỡ khỏi vị trí' : 'Tạo hồ sơ'}</strong><span class="text-xs text-[var(--color-text-secondary)]">{new Date(item.changed_at).toLocaleString('vi-VN')}</span></div><p class="mt-1 text-sm text-[var(--color-text-secondary)]">{item.notes || 'Thay đổi vị trí/Bài vị'}</p></div>{/each}</div>{/if}
</Popup>

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
			<div
				class="rounded-md bg-[var(--color-surface-muted)] p-3 text-xs text-[var(--color-text-secondary)]"
			>
				<p class="font-semibold text-[var(--color-text)]">File Excel cần đúng format chuẩn</p>
				<p class="mt-2">
					Các cột bắt buộc theo mẫu: Họ tên, Pháp danh, Năm sinh, Năm mất, Tuổi, Ảnh URL, Nơi an
					táng, Người gửi, Tháng gửi, Ghi chú, Vị trí, Bài vị.
				</p>
				<p class="mt-2">
					Nếu vị trí có dạng như <code>38D-10</code>, hệ thống sẽ tự parse thành Khu D, cột 38, hàng
					10 và tự tạo Khu, Vị trí, Bài vị nếu chưa có.
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
			<div
				class="rounded-md bg-[var(--color-surface-muted)] p-3 text-xs text-[var(--color-text-secondary)]"
			>
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
					<p class="mt-1 text-lg font-semibold text-[var(--color-text)]">
						{importPreview.total_rows}
					</p>
				</div>
				<div class="rounded-md bg-[var(--color-surface-muted)] p-3">
					<p class="text-[var(--color-text-secondary)]">Hợp lệ</p>
					<p class="mt-1 text-lg font-semibold text-[var(--color-text)]">
						{importPreview.valid_rows}
					</p>
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
					<p class="mt-1 text-lg font-semibold text-[var(--color-danger)]">
						{importPreview.invalid_rows}
					</p>
				</div>
			</div>
			{#if (importPreview.errors ?? []).length > 0}
				<div
					class="rounded-md border border-[var(--color-danger)]/30 bg-[var(--color-danger)]/5 p-3"
				>
					<p class="text-sm font-semibold text-[var(--color-danger)]">
						Các dòng cần sửa trước khi import
					</p>
					<ul
						class="mt-2 max-h-44 space-y-2 overflow-y-auto pr-1 text-xs text-[var(--color-text-secondary)]"
					>
						{#each importPreview.errors ?? [] as issue (`${issue.row_number}-${issue.message}`)}
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
				disabled={importBusy ||
					(importStep === 'guide' && !importHouseId) ||
					(importStep === 'upload' && !importFile) ||
					(importStep === 'preview' && Boolean(importPreview?.invalid_rows))}
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

<Popup
	open={bulkPatchOpen}
	title={`Cập nhật ${selectedCount} Hương linh`}
	onClose={() => !bulkBusy && (bulkPatchOpen = false)}
>
	<div class="space-y-3 text-sm">
		<p class="text-[var(--color-text-secondary)]">
			Chỉ field được chọn sẽ được cập nhật. Để trống giá trị nếu muốn xóa nội dung field đó.
		</p>
		<label class="block"
			><span class="mb-1 block font-medium">Field</span><select
				bind:value={bulkField}
				class="h-10 w-full rounded-md border-[var(--color-border-strong)]"
				><option value="dharma_name">Pháp danh</option><option value="birth_year">Năm sinh</option
				><option value="death_year">Năm mất</option><option value="age">Tuổi</option><option
					value="burial_place">Nơi an táng</option
				><option value="sender">Người gửi</option><option value="sent_month">Tháng gửi</option
				><option value="notes">Ghi chú</option></select
			></label
		>
		<label class="block"
			><span class="mb-1 block font-medium">Giá trị mới</span><input
				bind:value={bulkValue}
				class="h-10 w-full rounded-md border-[var(--color-border-strong)]"
			/></label
		>
	</div>
	{#snippet footer()}<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				onclick={() => (bulkPatchOpen = false)}
				class="h-10 rounded-md border border-[var(--color-border-strong)] font-semibold">Huỷ</button
			><button
				type="button"
				disabled={bulkBusy}
				onclick={() => void applyBulkPatch()}
				class="h-10 rounded-md bg-[var(--color-primary)] font-semibold text-white disabled:opacity-50"
				>{bulkBusy ? 'Đang cập nhật...' : 'Cập nhật'}</button
			>
		</div>{/snippet}
</Popup>

<Popup
	open={bulkDeleteOpen}
	title={`Xóa mềm ${selectedCount} Hương linh`}
	onClose={() => !bulkBusy && (bulkDeleteOpen = false)}
>
	<div class="space-y-3 text-sm">
		<div class="rounded-md bg-[var(--color-danger-soft)] p-3 text-[var(--color-danger)]">
			Các Hương linh này sẽ được ẩn khỏi toàn bộ danh sách, tìm kiếm, thống kê và export. Dữ liệu
			vẫn được lưu để có thể khôi phục về sau.
		</div>
		{#if selectedCount > 1}<label class="block"
				><span class="mb-1 block font-medium">Nhập <strong>XÓA</strong> để xác nhận</span><input
					bind:value={deleteConfirmation}
					class="h-10 w-full rounded-md border-[var(--color-border-strong)]"
				/></label
			>{/if}
	</div>
	{#snippet footer()}<div class="grid grid-cols-2 gap-3">
			<button
				type="button"
				onclick={() => (bulkDeleteOpen = false)}
				class="h-10 rounded-md border border-[var(--color-border-strong)] font-semibold">Huỷ</button
			><button
				type="button"
				disabled={bulkBusy || (selectedCount > 1 && deleteConfirmation !== 'XÓA')}
				onclick={() => void applyBulkDelete()}
				class="h-10 rounded-md bg-[var(--color-danger)] font-semibold text-white disabled:opacity-50"
				>{bulkBusy ? 'Đang xóa...' : 'Xác nhận xóa'}</button
			>
		</div>{/snippet}
</Popup>

<Popup
	open={vacantPositionPickerOpen}
	title="Chọn nhanh vị trí trống"
	sizeClass="max-w-4xl"
	layerClass="z-[60]"
	onClose={() => (vacantPositionPickerOpen = false)}
>
	<div class="space-y-3 text-sm">
		<p class="text-[var(--color-text-secondary)]">
			Chọn một ô trống theo đúng sơ đồ hàng/cột. Ô đã có Bài vị được khóa.
		</p>
		<label class="block">
			<span class="mb-1 block font-medium">Khu vực</span>
			<select
				bind:value={vacantPositionAreaId}
				class="h-10 w-full rounded-md border-[var(--color-border-strong)]"
			>
				{#each vacantPositionPicker?.areas ?? [] as area (area.id)}
					<option value={area.id}>Khu {area.code}{area.name ? ` – ${area.name}` : ''}</option>
				{/each}
			</select>
		</label>
		{#if vacantPositionPickerLoading}
			<LoadingIndicator label="Đang tải sơ đồ vị trí..." />
		{:else if vacantPickerPositions.length === 0}
			<p
				class="rounded-md bg-[var(--color-surface-muted)] p-4 text-center text-[var(--color-text-secondary)]"
			>
				Chưa có vị trí trong khu vực này.
			</p>
		{:else}
			<div class="max-h-[55dvh] overflow-auto rounded-md border border-[var(--color-border)] p-2">
				<div
					class="grid min-w-max gap-1"
					style={`grid-template-columns: repeat(${vacantPickerMaxColumn}, minmax(4.5rem, 1fr));`}
				>
					{#each vacantPickerRows as row}
						{#each vacantPickerColumns as column}
							{@const position = vacantPickerByCoordinate.get(`${row}:${column}`)}
							{@const positionName =
								position?.name ??
								`${column}${vacantPositionPicker?.areas.find((area) => area.id === vacantPositionAreaId)?.code ?? ''}-${row}`}
							<button
								type="button"
								disabled={(position?.tablet_count ?? 0) > 0}
								onclick={() => void selectVacantPosition(position, row, column)}
								class={[
									'flex min-h-15 flex-col items-center justify-center rounded-md border px-2 py-1 text-xs font-semibold',
									position?.tablet_count !== undefined && position.tablet_count > 0
										? 'cursor-not-allowed border-[var(--color-border)] bg-[var(--color-surface-muted)] text-[var(--color-text-muted)] opacity-65'
										: 'border-[var(--color-primary)] bg-[var(--color-primary-soft)] text-[var(--color-primary-dark)] hover:bg-[var(--color-primary)] hover:text-white'
								]}
								title={position?.tablet_count !== undefined && position.tablet_count > 0
									? `${positionName} đã có Bài vị`
									: `Chọn ${positionName}`}
							>
								<span>{positionName}</span>
								<span class="mt-0.5 text-[10px] font-normal"
									>{position?.tablet_count !== undefined && position.tablet_count > 0
										? 'Đã có Bài vị'
										: 'Trống'}</span
								>
							</button>
						{/each}
					{/each}
				</div>
			</div>
			<div class="flex items-center gap-4 text-xs text-[var(--color-text-secondary)]">
				<span
					><span
						class="mr-1 inline-block h-3 w-3 rounded-sm border border-[var(--color-primary)] bg-[var(--color-primary-soft)] align-middle"
					></span>Trống, có thể chọn</span
				>
				<span
					><span
						class="mr-1 inline-block h-3 w-3 rounded-sm border border-[var(--color-border)] bg-[var(--color-surface-muted)] align-middle"
					></span>Đã có Bài vị</span
				>
			</div>
		{/if}
	</div>
</Popup>

{#if formOpen}<div
		class="fixed inset-0 z-50 grid place-items-end bg-black/40 md:place-items-center"
		role="presentation"
		onclick={(e) => {
			if (e.target === e.currentTarget) void requestCloseForm();
		}}
	>
		<form
			onsubmit={save}
			oninput={markEditingFormChanged}
			onchange={markEditingFormChanged}
			class={[
				'h-[100dvh] w-full overflow-y-auto rounded-none bg-[var(--color-surface)] shadow-xl md:h-auto md:max-h-[94dvh] md:rounded-xl',
				editing || singleSpiritEntry
					? 'md:max-w-[800px]'
					: relatedTabletId
						? 'md:flex md:h-[94dvh] md:max-w-[calc(100vw-2rem)] md:flex-col md:overflow-hidden'
						: 'md:max-w-6xl'
			]}
		>
			<header
				class="sticky top-0 z-40 flex items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-3 md:px-5 md:py-4"
			>
				<h2 class="text-lg font-semibold">
					{editing
						? 'Sửa Hương linh'
						: relatedTabletId
							? 'Thêm Hương linh vào bài vị'
							: 'Thêm Hương linh'}
				</h2>
				<div class="flex items-center gap-3">{#if editing}<button type="button" onclick={() => { if (editing) void showSpiritHistory(editing); }} class="text-sm font-semibold text-[var(--color-primary-dark)]">Lịch sử vị trí</button>{/if}<button
					type="button"
					onclick={() => void requestCloseForm()}
					aria-label="Đóng"
					class="icon-[lucide--x] h-5 w-5 cursor-pointer"
				></button></div>
			</header>
			<div
				class={[
				'grid gap-4 p-4 md:p-5',
					relatedTabletId || singleSpiritEntry
						? 'md:min-h-0 md:flex-1 md:grid-cols-1 md:overflow-y-auto'
						: 'md:grid-cols-2'
				]}
			>
				{#if !editing}<div class="rounded-md border border-[var(--color-primary)]/30 bg-[var(--color-primary-soft)] p-3 text-sm text-[var(--color-primary-dark)] md:col-span-2">
					<p class="font-semibold">Nhập theo 3 bước</p>
					<p class="mt-1 text-xs leading-relaxed">1. Chọn Nhà Linh và vị trí (có thể để trống). 2. Chọn Bài vị có sẵn hoặc tạo mới tại vị trí đó. 3. Nhập hồ sơ Hương linh và lưu.</p>
				</div>{/if}
				{#if relatedTabletId && editing && selectedFormTablet}<div
						class="w-full shrink-0 overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface-muted)]/35 p-3 md:col-span-2"
					>
						<SpiritImageUploader
							imageUrl={tabletDisplayImage(relatedSpirits)}
							displayName={selectedFormTablet.name}
							uploading={imageUploading}
							tablet
							portraitClass="w-20 text-lg"
							onselect={selectTabletImage}
						/>
						{#if !selectedFormTablet.image_url && relatedSpirits.length === 1 && relatedSpirits[0].image_url}<p class="mt-2 text-xs text-[var(--color-text-secondary)]">Đang dùng mặc định ảnh của Hương linh duy nhất.</p>{:else if relatedSpirits.length > 1}<label class="mt-2 block text-xs text-[var(--color-text-secondary)]">Ảnh đại diện từ Hương linh<select value={selectedFormTablet.image_url} onchange={(event) => { selectedFormTablet.image_url = event.currentTarget.value; markEditingFormChanged(); }} class="mt-1 h-8 w-full rounded border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-2 text-sm"><option value="">Chưa chọn — dùng ảnh Hương linh tạm thời</option>{#each relatedSpirits.filter((spirit) => spirit.image_url) as spirit (spirit.id)}<option value={spirit.image_url}>{spirit.full_name}</option>{/each}</select></label>{/if}
					</div>{:else if creatingTablet}<div class="w-full shrink-0 overflow-hidden rounded-md border border-[var(--color-border)] bg-[var(--color-surface-muted)]/35 p-3 md:col-span-2">
						<SpiritImageUploader
							imageUrl={newTabletForm.image_url || newSpirits.find((spirit) => spirit.image_url)?.image_url || ''}
							displayName={newTabletForm.name || form.full_name}
							uploading={imageUploading}
							tablet
							portraitClass="w-20 text-lg"
							onselect={selectNewTabletImage}
						/>
						{#if !newTabletForm.image_url && newSpirits.length === 1 && newSpirits[0].image_url}<p class="mt-2 text-xs text-[var(--color-text-secondary)]">Ảnh của Hương linh duy nhất sẽ tự trở thành ảnh Bài vị khi lưu.</p>{:else if newSpirits.length > 1}<label class="mt-2 block text-xs text-[var(--color-text-secondary)]">Ảnh đại diện từ Hương linh<select bind:value={newTabletForm.image_url} class="mt-1 h-8 w-full rounded border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-2 text-sm"><option value="">Chưa chọn — dùng ảnh Hương linh tạm thời</option>{#each newSpirits.filter((spirit) => spirit.image_url) as spirit (spirit.id)}<option value={spirit.image_url}>{spirit.full_name}</option>{/each}</select></label>{/if}
					</div>{/if}
				<div
					class={relatedTabletId || singleSpiritEntry
						? 'min-w-0 md:min-h-0 md:overflow-y-auto md:pr-1'
						: 'contents'}
				>
					<div
						class={relatedTabletId || singleSpiritEntry
							? 'grid content-start gap-4 md:grid-cols-2'
							: 'contents'}
					>
						{#if false && houses.length > 1 && !editing}<label
								><span class="mb-1 block text-sm font-medium">Nhà Linh *</span><select
									bind:value={formHouseId}
									disabled={Boolean(editing) || Boolean(relatedTabletId)}
									onchange={() => void changeFormHouse()}
									required
									class="h-11 w-full rounded-md border-[var(--color-border-strong)] disabled:opacity-60"
									>{#each houses as house (house.id)}<option value={house.id}>{house.name}</option
										>{/each}</select
								></label
							>{/if}
						{#if singleSpiritEntry || editing}<label>
								<span class="mb-1 block text-sm font-medium">Khu vực</span>
								<select
									bind:value={formAreaId}
									onchange={changeFormArea}
									disabled={Boolean(relatedTabletId && !editing)}
									class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
								>
									<option value="">Tất cả khu vực</option>
									{#each formAreas as area (area.id)}
										<option value={area.id}
											>Khu {area.code}{area.name ? ` – ${area.name}` : ''}</option
										>
									{/each}
								</select>
							</label>{/if}
						<div
							class={`relative ${houses.length > 1 || singleSpiritEntry || editing ? '' : 'md:col-span-2'}`}
						>
							<label class="block"
								><span class="mb-1 block text-sm font-medium">Bước 1 · Vị trí (không bắt buộc)</span>
								<div class="flex gap-2">
									<div class="relative min-w-0 flex-1">
										<input
											bind:value={formPositionQuery}
											oninput={searchFormPosition}
											placeholder="Tìm trực tiếp, ví dụ 1a1..."
											disabled={Boolean(relatedTabletId && !editing)}
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
									{#if singleSpiritEntry || editing}<button
											type="button"
											onclick={() => void openVacantPositionPicker()}
											class="grid h-11 w-11 shrink-0 place-items-center rounded-md border border-[var(--color-primary)] text-[var(--color-primary-dark)]"
											aria-label="Chọn nhanh vị trí trống"
											title="Chọn nhanh vị trí trống"
											><span class="icon-[lucide--grid-3x3] h-5 w-5" aria-hidden="true"
											></span></button
										>{/if}
									{#if selectedFormPosition && !(relatedTabletId && !editing)}<button
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
												>Khu {position.area_code} · cột {position.column_number}, hàng {position.row_number}
												·
												{position.tablet_count} bài vị</span
											></button
										>{/each}
								</div>{/if}
						</div>
						{#if selectedFormPosition && !editing && !singleSpiritEntry}<label class="md:col-span-2"
								><span class="mb-1 block text-sm font-medium">Bước 2 · Bài vị hiện có</span><select
									bind:value={form.tablet_id}
									onchange={() => (quickCreateTablet = false)}
									disabled={tabletLoading || Boolean(relatedTabletId && !editing)}
									class="h-11 w-full rounded-md border-[var(--color-border-strong)] disabled:opacity-60"
									><option value=""
										>{tabletLoading
											? 'Đang tải danh sách bài vị...'
											: 'Chưa chọn — có thể tạo nhanh bài vị mới'}</option
									>{#each formTablets as tablet (tablet.id)}<option value={tablet.id}
											>{tablet.name} · {tablet.spirit_count} Hương linh</option
										>{/each}</select
								></label
							>{:else if !selectedFormPosition}<p
								class="rounded-md bg-[var(--color-primary-soft)] px-3 py-2 text-xs text-[var(--color-primary-dark)] md:col-span-2"
							>
								Để trống vị trí để lưu Hương linh vào danh sách chưa xếp.
							</p>{/if}
						{#if editing && selectedFormTablet}
							<div class="grid gap-3 md:col-span-2 md:grid-cols-3"><label>
								<span class="mb-1 block text-sm font-medium">Mã Bài vị</span>
								<input value={selectedFormTablet.code} readonly class="h-10 w-full rounded-md border-[var(--color-border-strong)] bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)]" />
							</label><label><span class="mb-1 block text-sm font-medium">Loại Bài vị</span><select bind:value={selectedFormTablet.type} class="h-10 w-full rounded-md border-[var(--color-border-strong)]"><option value="spirit">Hương linh</option><option value="giac_linh">Giác linh</option><option value="family">Gia tiên</option><option value="cuu_huyen">Cửu Huyền Thất Tổ</option><option value="clan">Tộc họ</option><option value="collective">Chư Hương linh</option><option value="fetus">Thai nhi</option><option value="martyr">Anh hùng liệt sĩ</option><option value="victim">Hương linh tử nạn</option><option value="childless">Hương linh vô tự</option><option value="unknown">Hương linh vô danh</option><option value="other">Khác</option></select></label><label><span class="mb-1 block text-sm font-medium">Trạng thái</span><select bind:value={selectedFormTablet.status} class="h-10 w-full rounded-md border-[var(--color-border-strong)]"><option value="pending">Chờ an vị</option><option value="enshrined">Đã an vị</option><option value="taken_home">Đã thỉnh về</option></select></label></div>
							<div class="grid gap-3 md:col-span-2 md:grid-cols-3"><label><span class="mb-1 block text-sm font-medium">Ngày đăng ký</span><input type="text" bind:value={selectedFormTablet.registered_at} placeholder="VD: 00/03/2025" class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label><label><span class="mb-1 block text-sm font-medium">Ngày an vị</span><input type="date" bind:value={selectedFormTablet.enshrined_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label><label><span class="mb-1 block text-sm font-medium">Ngày đưa vào khu thờ</span><input type="date" bind:value={selectedFormTablet.entered_worship_area_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label></div>
							<label>
								<span class="mb-1 block text-sm font-medium">Người gửi</span>
								<input
									value={selectedFormTablet.sender}
									oninput={(event) => (selectedFormTablet.sender = event.currentTarget.value)}
									class="h-10 w-full rounded-md border-[var(--color-border-strong)] bg-[var(--color-surface)]"
								/>
							</label>
							<label class="md:col-span-2">
								<span class="mb-1 block text-sm font-medium">Ghi chú</span>
								<textarea
									value={selectedFormTablet.notes}
									oninput={(event) => (selectedFormTablet.notes = event.currentTarget.value)}
									rows="3"
									class="w-full rounded-md border-[var(--color-border-strong)] bg-[var(--color-surface)]"
								></textarea>
							</label>
							{#if relatedTabletId}
								<section
									class="flex min-h-0 min-w-0 flex-col overflow-hidden rounded-md border border-[var(--color-border)] md:col-span-2"
								>
									<header
										class="flex flex-col items-stretch gap-3 border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] px-3 py-2 sm:flex-row sm:items-center sm:justify-between"
									>
										<div class="min-w-0">
													<h3 class="truncate text-sm font-semibold">
														Hương linh cùng bài vị: {relatedTabletName}
													</h3>
												</div>
										{#if canWrite}<button
												type="button"
												onclick={addRelatedSpirit}
												class="h-10 shrink-0 rounded-md border border-[var(--color-primary)] px-3 text-xs font-semibold text-[var(--color-primary-dark)] sm:h-9"
											>+ Thêm Hương Linh khác</button
											>{/if}
									</header>
									<div class="overflow-auto">
										{#if relatedSpiritsLoading}
											<div class="py-6">
												<LoadingIndicator label="Đang tải Hương linh cùng bài vị..." />
											</div>
										{:else}
											<div class="grid gap-3 p-3 sm:grid-cols-2 xl:grid-cols-3">
												{#each relatedSpirits as related (related.id)}
													<article
														class={[
															'relative overflow-hidden rounded-[5px] border border-[#c7ad78] bg-[#fffdf7] p-3 shadow-[0_1px_2px_rgb(73_53_28_/_10%)]',
															related.id === editing?.id ? 'ring-2 ring-[var(--color-primary)]/45' : ''
														]}
													>
														<span class="absolute inset-1 rounded-[3px] border border-[#c7ad78]/30 pointer-events-none"></span>
														<div class="relative flex items-start gap-3">
															{#if related.image_url}<img
																src={related.image_url}
																alt={`Ảnh ${related.full_name}`}
																class="h-12 w-10 shrink-0 border border-[#6f4b3b] object-cover"
															/>{:else}<span
																class="grid h-12 w-10 shrink-0 place-items-center border border-[#6f4b3b] text-[#6f4b3b]/50"
																><span class="icon-[lucide--user-round] h-4 w-4" aria-hidden="true"></span></span
															>{/if}
															<div class="min-w-0 flex-1">
																<p class="mb-1 text-[10px] font-bold tracking-[0.06em] text-[#80673c]">HƯƠNG LINH</p>
																<p class="truncate font-semibold">{related.full_name}</p>
																<p class="text-sm text-[var(--color-text-secondary)]">
																	{related.dharma_name || 'Chưa có pháp danh'}
																</p>
															</div>
														</div>
														<div class="relative mt-3 grid grid-cols-2 gap-x-3 gap-y-1.5 text-xs text-[var(--color-text-secondary)]">
															<span>Năm sinh: {related.birth_year || '—'}</span>
															<span>Năm mất: {related.death_year || '—'}</span>
															<span>Tuổi: {related.age || '—'}</span>
															<span>Ngày gửi: {related.sent_month || '—'}</span>
															<span class="col-span-2">Nơi an táng: {related.burial_place || '—'}</span>
															{#if related.death_date}<span class="col-span-2">Ngày mất/kỵ: {related.death_date}</span>{/if}
															{#if related.enshrined_at}<span class="col-span-2">Ngày an vị: {related.enshrined_at}</span>{/if}
														</div>
														{#if canWrite}<div class="relative mt-3 grid grid-cols-2 gap-2">
															<button
																type="button"
																onclick={(event) => {
																	event.stopPropagation();
																	selectRelatedSpirit(related);
																}}
																class="h-10 rounded-md border border-[var(--color-primary)] text-sm font-semibold text-[var(--color-primary-dark)]"
															>Chỉnh sửa</button
															>
															<button
																type="button"
																disabled={relatedSpirits.length <= 1}
																onclick={(event) => {
																	event.stopPropagation();
																	void removeRelatedSpirit(related);
																}}
																class="h-10 rounded-md border border-[var(--color-danger)] text-sm font-semibold text-[var(--color-danger)] disabled:cursor-not-allowed disabled:opacity-40"
															>Xóa</button
															>
														</div>{/if}
													</article>
												{/each}
											</div>
											<table class="hidden w-full table-auto text-left text-sm">
												<thead
													class="bg-[var(--color-surface)] text-xs text-[var(--color-text-secondary)]"
												>
													<tr>
														{#if canWrite}<th class="w-16 px-1 py-2 text-center">Thao tác</th>{/if}
														<th class="w-12 px-1 py-2 text-center">Ảnh</th>
														<th class="px-1 py-2">Họ tên</th>
														<th class="px-1 py-2">Pháp danh</th>
														<th class="px-1 py-2">Năm sinh</th>
														<th class="px-1 py-2">Năm mất</th>
														<th class="px-1 py-2">Tuổi</th>
														<th class="px-1 py-2">Tháng gửi</th>
														<th class="px-1 py-2">Nơi an táng</th>
													</tr>
												</thead>
												<tbody class="divide-y divide-[var(--color-border)]">
													{#each relatedSpirits as related (related.id)}
														<tr
															class={[
																'transition-colors',
																related.id === editing?.id
																	? 'bg-[var(--color-primary-soft)]'
																	: canWrite
																		? 'cursor-pointer hover:bg-[var(--color-surface-muted)]'
																		: ''
															]}
															onclick={() => selectRelatedSpirit(related)}
														>
															{#if canWrite}<td class="px-1 py-2">
																	<div class="flex justify-center gap-1">
																		<button
																			type="button"
																			onclick={(event) => {
																				event.stopPropagation();
																				selectRelatedSpirit(related);
																			}}
																			class="grid h-7 w-7 place-items-center rounded border border-[var(--color-primary)] text-[var(--color-primary-dark)]"
																			aria-label={`Chỉnh sửa ${related.full_name}`}
																			title={`Chỉnh sửa ${related.full_name}`}
																			><span
																				class="icon-[lucide--pencil] h-3.5 w-3.5"
																				aria-hidden="true"
																			></span></button
																		><button
																			type="button"
																			disabled={relatedSpirits.length <= 1}
																			onclick={(event) => {
																				event.stopPropagation();
																				void removeRelatedSpirit(related);
																			}}
																			class="grid h-7 w-7 place-items-center rounded border border-[var(--color-danger)] text-[var(--color-danger)] disabled:cursor-not-allowed disabled:opacity-40"
																			aria-label={`Xóa ${related.full_name}`}
																			title={relatedSpirits.length <= 1
																				? 'Bài vị cần tối thiểu một Hương linh'
																				: `Xóa ${related.full_name}`}
																			><span
																				class="icon-[lucide--trash-2] h-3.5 w-3.5"
																				aria-hidden="true"
																			></span></button
																		>
																	</div>
																</td>{/if}
															<td class="px-1 py-2">
																{#if related.image_url}<img
																		src={related.image_url}
																		alt={`Ảnh ${related.full_name}`}
																		class="h-9 w-7 border border-[#6f4b3b] object-cover"
																	/>{:else}<span
																		class="grid h-9 w-7 place-items-center border border-[#6f4b3b] text-[#6f4b3b]/50"
																		><span
																			class="icon-[lucide--user-round] h-4 w-4"
																			aria-hidden="true"
																		></span></span
																	>{/if}
															</td>
															<td class="px-1 py-2 font-medium">{related.full_name}</td>
															<td class="px-1 py-2">{related.dharma_name || '—'}</td>
															<td class="px-1 py-2">{related.birth_year || '—'}</td>
															<td class="px-1 py-2">{related.death_year || '—'}</td>
															<td class="px-1 py-2">{related.age || '—'}</td>
															<td class="px-1 py-2">{related.sent_month || '—'}</td>
															<td class="px-1 py-2">{related.burial_place || '—'}</td>
														</tr>
													{/each}
												</tbody>
											</table>
										{/if}
									</div>
								</section>
							{/if}
						{:else if creatingTablet}
							<div class="grid gap-3 md:col-span-2 md:grid-cols-3"><label>
								<span class="mb-1 block text-sm font-medium">Mã Bài vị</span>
								<input value="Tự động tạo" readonly class="h-10 w-full rounded-md border-[var(--color-border-strong)] bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)]" />
							</label><label><span class="mb-1 block text-sm font-medium">Loại Bài vị</span><select bind:value={newTabletForm.type} class="h-10 w-full rounded-md border-[var(--color-border-strong)]"><option value="spirit">Hương linh</option><option value="giac_linh">Giác linh</option><option value="family">Gia tiên</option><option value="cuu_huyen">Cửu Huyền Thất Tổ</option><option value="clan">Tộc họ</option><option value="collective">Chư Hương linh</option><option value="fetus">Thai nhi</option><option value="martyr">Anh hùng liệt sĩ</option><option value="victim">Hương linh tử nạn</option><option value="childless">Hương linh vô tự</option><option value="unknown">Hương linh vô danh</option><option value="other">Khác</option></select></label><label><span class="mb-1 block text-sm font-medium">Trạng thái *</span><select bind:value={newTabletForm.status} class="h-10 w-full rounded-md border-[var(--color-border-strong)]"><option value="pending">Chờ an vị</option><option value="enshrined">Đã an vị</option><option value="taken_home">Đã thỉnh về</option></select></label></div><div class="grid gap-3 md:col-span-2 md:grid-cols-3"><label><span class="mb-1 block text-sm font-medium">Ngày đăng ký</span><input bind:value={newTabletForm.registered_at} placeholder="VD: 00/03/2025" class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label><label><span class="mb-1 block text-sm font-medium">Ngày an vị</span><input type="date" bind:value={newTabletForm.enshrined_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label><label><span class="mb-1 block text-sm font-medium">Ngày đưa vào khu thờ</span><input type="date" bind:value={newTabletForm.entered_worship_area_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label></div>
							<label>
								<span class="mb-1 block text-sm font-medium">Người gửi</span>
								<input
									bind:value={newTabletForm.sender}
									class="h-10 w-full rounded-md border-[var(--color-border-strong)] bg-[var(--color-surface)]"
								/>
							</label>
							<label class="md:col-span-2">
								<span class="mb-1 block text-sm font-medium">Ghi chú</span>
								<textarea
									bind:value={newTabletForm.notes}
									rows="3"
									class="w-full rounded-md border-[var(--color-border-strong)] bg-[var(--color-surface)]"
								></textarea>
							</label>
						{/if}
						{#if creatingTablet && hasNewTabletSpirit}
							<section
								class="min-w-0 overflow-hidden rounded-md border border-[var(--color-border)] md:col-span-2"
							>
								<header
									class="flex items-center justify-between gap-3 border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] px-3 py-2"
								>
									<div class="min-w-0">
										<h3 class="truncate text-sm font-semibold">
									Hương linh cùng bài vị: {newSpirits[0]?.full_name || 'Bài vị mới'}
										</h3>
										<p class="text-xs text-[var(--color-text-secondary)]">
											Hương linh đầu tiên sẽ được tạo cùng Bài vị này.
										</p>
									</div>
								</header>
								<div class="overflow-auto">
									<table class="w-full table-auto text-left text-sm">
										<thead
											class="bg-[var(--color-surface)] text-xs text-[var(--color-text-secondary)]"
										>
											<tr>
												<th class="w-16 px-1 py-2 text-center">Thao tác</th>
												<th class="w-12 px-1 py-2 text-center">Ảnh</th>
												<th class="px-1 py-2">Họ tên</th>
												<th class="px-1 py-2">Pháp danh</th>
												<th class="px-1 py-2">Năm sinh</th>
												<th class="px-1 py-2">Năm mất</th>
												<th class="px-1 py-2">Tuổi</th>
												<th class="px-1 py-2">Tháng gửi</th>
												<th class="px-1 py-2">Nơi an táng</th>
											</tr>
										</thead>
										<tbody class="divide-y divide-[var(--color-border)]">
											{#each newSpirits as spirit, index (index)}<tr class="bg-[var(--color-primary-soft)]">
												<td class="px-1 py-2">
													<div class="flex justify-center gap-1">
														<button
															type="button"
															onclick={() => removeNewTabletSpirit(index)}
															class="grid h-7 w-7 place-items-center rounded border border-[var(--color-danger)] text-[var(--color-danger)]"
															aria-label={`Xóa ${spirit.full_name}`}
															title={`Xóa ${spirit.full_name}`}
															><span class="icon-[lucide--trash-2] h-3.5 w-3.5" aria-hidden="true"
															></span></button
														>
													</div>
												</td>
												<td class="px-1 py-2">
													{#if spirit.image_url}<img
															src={spirit.image_url}
															alt={`Ảnh ${spirit.full_name}`}
															class="h-9 w-7 border border-[#6f4b3b] object-cover"
														/>{:else}<span
															class="grid h-9 w-7 place-items-center border border-[#6f4b3b] text-[#6f4b3b]/50"
															><span class="icon-[lucide--user-round] h-4 w-4" aria-hidden="true"
															></span></span
														>{/if}
												</td>
												<td class="px-1 py-2 font-medium">{spirit.full_name || '—'}</td>
												<td class="px-1 py-2">{spirit.dharma_name || '—'}</td>
												<td class="px-1 py-2">{spirit.birth_year || '—'}</td>
												<td class="px-1 py-2">{spirit.death_year || '—'}</td>
												<td class="px-1 py-2">{spirit.age || '—'}</td>
												<td class="px-1 py-2">{spirit.sent_month || '—'}</td>
												<td class="px-1 py-2">{spirit.burial_place || '—'}</td>
											</tr>{/each}
										</tbody>
									</table>
								</div>
							</section>
						{/if}
						{#if showSpiritEditor}{#if editing || singleSpiritEntry}<section
									class="rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] md:col-span-2"
								>
									<h3
										class="border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] px-3 py-2 text-sm font-semibold text-[var(--color-primary-dark)]"
									>
										{addingRelatedSpirit
											? 'Thông tin Hương Linh'
											: editing
												? 'Cập nhập thông tin Hương Linh'
												: 'Thông tin Hương Linh'}
									</h3>
									<div class={addingRelatedSpirit ? 'grid gap-2 p-3 sm:grid-cols-2 xl:grid-cols-3' : 'grid gap-4 p-3 md:grid-cols-2'}>
										<div class={addingRelatedSpirit ? 'shrink-0 overflow-visible pb-1 sm:col-span-2 xl:col-span-1' : 'shrink-0 overflow-visible pb-1 md:col-span-2'}>
											<SpiritImageUploader
												imageUrl={addingRelatedSpirit ? relatedSpiritForm.image_url : form.image_url}
												displayName={addingRelatedSpirit
													? relatedSpiritForm.full_name
													: form.full_name || editing?.full_name || ''}
												uploading={imageUploading}
												portraitClass="w-[75px] text-xl"
												compactFrame
												onselect={addingRelatedSpirit ? selectRelatedSpiritImage : selectSpiritImage}
											/>
										</div>
										{@render field('Họ tên *', 'full_name', true, addingRelatedSpirit ? relatedSpiritForm : form)}{@render field(
											'Pháp danh',
											'dharma_name',
											false,
											addingRelatedSpirit ? relatedSpiritForm : form
										)}{@render field(
															'Giới tính',
															'gender',
															false,
															addingRelatedSpirit ? relatedSpiritForm : form
														)}{@render field(
															'Quê quán',
															'familiar_name',
															false,
															addingRelatedSpirit ? relatedSpiritForm : form
														)}{@render field(
															'Ngày mất (dương)',
															'death_date',
															false,
															addingRelatedSpirit ? relatedSpiritForm : form
														)}{@render field(
															'Năm sinh',
											'birth_year',
											false,
											addingRelatedSpirit ? relatedSpiritForm : form
										)}{@render field(
											'Năm mất',
											'death_year',
											false,
											addingRelatedSpirit ? relatedSpiritForm : form
										)}{@render field('Tuổi', 'age', false, addingRelatedSpirit ? relatedSpiritForm : form)}{@render field('Nơi an táng', 'burial_place', false, addingRelatedSpirit ? relatedSpiritForm : form)}
										<label class={addingRelatedSpirit ? 'flex h-9 items-center gap-2 rounded-md border border-[var(--color-border-strong)] px-2 text-xs sm:col-span-2 xl:col-span-1' : 'flex h-11 items-center gap-2 rounded-md border border-[var(--color-border-strong)] px-3 text-sm md:col-span-2'}>
											<input
												type="checkbox"
												checked={(addingRelatedSpirit ? relatedSpiritForm : form).has_urn}
												onchange={(event) => {
													if (addingRelatedSpirit) relatedSpiritForm.has_urn = event.currentTarget.checked;
													else form.has_urn = event.currentTarget.checked;
												}}
												class="h-4 w-4 rounded border-[var(--color-border-strong)] text-[var(--color-primary)]"
											/>
															<span>Lưu tro cốt/hài cốt</span>
										</label>
										{#if addingRelatedSpirit || (singleSpiritEntry && creatingTablet)}<div class="flex justify-end md:col-span-2">
											<button
												type="button"
												onclick={() =>
													addingRelatedSpirit ? void updateRelatedSpiritList() : addSpiritToNewTablet()}
												disabled={saving || imageUploading}
												class="h-11 w-full rounded-md bg-[var(--color-primary)] px-5 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50 md:h-10 md:w-auto"
											>Thêm vào danh sách</button
											>
										</div>{/if}
										{#if editing && selectedFormTablet && !addingRelatedSpirit}<div class="flex justify-end md:col-span-2">
											<button
												type="button"
												onclick={updateEditedSpiritInRelatedList}
												disabled={saving || imageUploading}
												class="h-11 w-full rounded-md bg-[var(--color-primary)] px-5 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50 md:h-10 md:w-auto"
											>Cập nhập vào danh sách</button
											>
										</div>{/if}
									</div>
								</section>
							{:else}<section
									class="rounded-md border border-[var(--color-border)] bg-[var(--color-surface)] md:col-span-2"
								>
									<h3
										class="border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] px-3 py-2 text-sm font-semibold text-[var(--color-primary-dark)]"
									>
										Cập nhập thông tin Hương Linh
									</h3>
									<div class="h-[52dvh] min-h-[360px] p-3">
										<InlineSpiritEditor
											bind:items={newSpirits}
											onbusychange={(busy) => (inlineEditorBusy = busy)}
										/>
									</div>
								</section>{/if}{/if}
					</div>
				</div>
			</div>
			<footer
				class="sticky bottom-0 z-40 flex gap-3 border-t border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-3 md:justify-end md:px-5 md:py-4"
			>
				<button
					type="button"
					onclick={() => void requestCloseForm()}
					class="h-11 flex-1 cursor-pointer rounded-md border border-[var(--color-border-strong)] px-5 text-sm font-semibold md:flex-none"
					>Huỷ</button
				>{#if !editing && !singleSpiritEntry && !relatedTabletId && selectedFormPosition && form.tablet_id}<button
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
						Boolean(editing && selectedFormTablet && relatedSpiritsLoading) ||
						Boolean(editing && selectedFormPosition && !form.tablet_id) ||
						(!editing &&
							!singleSpiritEntry &&
							selectedFormPosition &&
							!form.tablet_id &&
							!newSpirits[0]?.full_name.trim())}
					onclick={() =>
						(quickCreateTablet = Boolean(
							!editing && !singleSpiritEntry && selectedFormPosition && !form.tablet_id
						))}
					class="h-11 flex-1 cursor-pointer rounded-md bg-[var(--color-primary)] px-6 text-sm font-semibold text-white disabled:cursor-not-allowed disabled:opacity-50 md:flex-none"
					>{saving
						? 'Đang lưu...'
						: editing
							? selectedFormTablet
								? 'Lưu Bài vị'
								: 'Lưu'
							: singleSpiritEntry
								? selectedFormPosition && !form.tablet_id
									? 'Lưu Bài vị'
									: 'Lưu'
								: relatedTabletId
									? 'Lưu'
									: !editing && selectedFormPosition && !form.tablet_id
										? 'Tạo bài vị & thêm'
										: !editing && !selectedFormPosition
											? 'Lưu chưa xếp vị trí'
											: 'Lưu'}</button
				>
			</footer>
		</form>
	</div>{/if}

{#snippet field(label: string, key: keyof SpiritInput, required = false, spirit: SpiritInput = form)}<label
		><span class={addingRelatedSpirit ? 'mb-1 block text-xs font-medium' : 'mb-1 block text-sm font-medium'}>{label}</span><input
			type={key === 'birth_date' || key === 'death_date' || key === 'entered_worship_area_at' || key === 'enshrined_at' ? 'date' : 'text'}
			bind:value={spirit[key]}
			{required}
			class={addingRelatedSpirit ? 'h-9 w-full rounded-md border-[var(--color-border-strong)] text-sm' : 'h-11 w-full rounded-md border-[var(--color-border-strong)]'}
		/></label
	>{/snippet}

{#snippet highlight(value: string)}{#each highlightSegments(value, query) as segment}
		{#if segment.match}<mark class="rounded-sm bg-amber-200/90 px-px font-semibold text-amber-950"
				>{segment.text}</mark
			>{:else}{segment.text}{/if}
	{/each}{/snippet}

{#snippet tabletCardGrid()}<div class="grid grid-cols-[repeat(auto-fill,minmax(144px,1fr))] gap-4 p-1 sm:grid-cols-[repeat(auto-fill,144px)]">
		{#each tableSpiritGroups as group (group.key)}{@const first = group.items[0]}<button
				type="button"
				onclick={() => canWrite && void edit(first)}
				disabled={!canWrite}
				class="h-[211px] text-left enabled:cursor-pointer disabled:cursor-default"
				title={group.hasTablet ? `Mở Bài vị ${first.tablet_name}` : 'Hương linh chưa xếp Bài vị'}
			><CardBaiVi
					code={first.position_name || 'Chưa xếp'}
					spiritCount={group.items.length}
					spiritNames={group.items.map((item) => item.full_name)}
					imageUrl={tabletDisplayImage(group.items)}
					showDefaultPortrait
					birthYear={first.birth_year}
					deathYear={first.death_year}
					tabletType={first.tablet_type}
					status={first.status === 'draft' ? 'pending' : 'enshrined'}
					fontSize={16}
				/></button>
		{/each}
	</div>{/snippet}

{#snippet spiritCardGroups()}<div class="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-3">
		{#each tableSpiritGroups as group (group.key)}<section
				class="overflow-hidden rounded-lg border border-[var(--color-border)] bg-[var(--color-surface-muted)]/35"
			>
				<div
					class="flex items-center justify-between gap-3 border-b border-[var(--color-border)] bg-[var(--color-surface-muted)] px-4 py-2.5 text-sm"
				>
					<div class="min-w-0">
						<p class="min-w-0 truncate font-semibold text-[var(--color-primary-dark)]">
							{@render highlight(
								group.hasPosition
									? `Khu ${group.items[0].area_code} · ${group.items[0].position_name} · ${group.items[0].tablet_name}`
									: group.hasTablet
										? `Bài vị chưa xếp · ${group.items[0].tablet_name}`
										: 'Hương linh chưa xếp bài vị'
							)}
						</p>
						<p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">
							Danh sách Hương linh thuộc Bài vị
						</p>
					</div>
					<div class="flex shrink-0 items-center gap-2">
						<span class="text-xs text-[var(--color-text-secondary)]"
							>{group.items.length} Hương linh</span
						>
						{#if canWrite && group.hasTablet}<button
								type="button"
								onclick={() => void edit(group.items[0])}
								class="h-8 rounded-md border border-[var(--color-primary)] px-2.5 text-xs font-semibold text-[var(--color-primary-dark)]"
								>Mở</button
							>{/if}
					</div>
				</div>
				<div class="overflow-x-auto">
					<table class="w-full min-w-[360px] text-left text-xs">
						<thead class="text-xs text-[var(--color-text-secondary)]"
							><tr
								>{#if canWrite}<th class="hidden w-0 p-0"><span class="sr-only">Chọn</span></th
									>{/if}<th class="px-2 py-2 font-medium">Họ tên</th><th
									class="px-2 py-2 font-medium">Pháp danh</th
								><th class="px-2 py-2 font-medium">Sinh</th><th class="px-2 py-2 font-medium"
									>Mất</th
								>{#if canWrite}<th class="w-16 px-4 py-2"><span class="sr-only">Sửa</span></th
									>{/if}</tr
							></thead
						><tbody class="divide-y divide-[var(--color-border)] bg-[var(--color-surface)]"
							>{#each group.items as item (item.id)}<tr
									class={[
										'transition-colors',
										canWrite && 'cursor-pointer hover:bg-[var(--color-primary-soft)]/45'
									]}
									onclick={() => canWrite && void edit(item)}
									>{#if canWrite}<td class="hidden w-0 p-0"
											><input
												type="checkbox"
												checked={selectedSpiritIDs.has(item.id)}
												onchange={(event) => {
													event.stopPropagation();
													toggleSpiritSelection(item.id);
												}}
												aria-label={`Chọn ${item.full_name}`}
												class="h-4 w-4 rounded border-[var(--color-border-strong)] text-[var(--color-primary)]"
											/></td
										>{/if}<td class="px-4 py-2 font-medium">{@render highlight(item.full_name)}</td
									><td class="px-4 py-2">{@render highlight(item.dharma_name || '—')}</td><td
										class="px-4 py-2">{@render highlight(item.birth_year || '—')}</td
									><td class="px-4 py-2">{@render highlight(item.death_year || '—')}</td
									>{#if canWrite}<td class="px-4 py-2 text-right"
											><button
												type="button"
												onclick={(event) => {
													event.stopPropagation();
													void edit(item);
												}}
												class="text-xs font-semibold text-[var(--color-primary-dark)]">Sửa</button
											></td
										>{/if}</tr
								>{/each}</tbody
						>
					</table>
				</div>
			</section>{/each}
	</div>{/snippet}

{#snippet spiritCard(item: Spirit)}<li
		class="rounded-lg border border-[var(--color-border)] bg-[var(--color-surface)] p-4"
	>
		<div class="flex gap-3">
			{#if canWrite}<input
					type="checkbox"
					checked={selectedSpiritIDs.has(item.id)}
					onchange={() => toggleSpiritSelection(item.id)}
					aria-label={`Chọn ${item.full_name}`}
					class="mt-1 h-4 w-4 shrink-0 rounded border-[var(--color-border-strong)] text-[var(--color-primary)]"
				/>{/if}
			<button
				type="button"
				onclick={() => openSpiritImage(item)}
				disabled={!item.image_url}
				class="shrink-0 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-primary)] disabled:cursor-default"
				aria-label={`Xem ảnh ${item.full_name}`}
				><SpiritPortrait
					imageUrl={item.image_url}
					alt={item.full_name}
					sizeClass="h-20 w-15"
				/></button
			>
			<div class="min-w-0 flex-1">
				<h3 class="truncate font-semibold">{@render highlight(item.full_name)}</h3>
				<p class="truncate text-sm text-[var(--color-text-secondary)]">
					{@render highlight(item.dharma_name || 'Chưa có pháp danh')}
				</p>
				<p class="mt-1 text-xs font-semibold text-[var(--color-primary-dark)]">
					{@render highlight(
						item.position_id
							? `Khu ${item.area_code} · ${item.position_name} · ${item.tablet_name}`
							: 'Chưa xếp vị trí'
					)}
				</p>
			</div>
		</div>
		<dl
			class="mt-3 grid grid-cols-2 gap-x-3 gap-y-1 border-t border-[var(--color-border)] pt-3 text-xs"
		>
			<dt class="text-[var(--color-text-muted)]">Năm sinh – mất</dt>
			<dd class="text-right">{item.birth_year || '?'} – {item.death_year || '?'}</dd>
			<dt class="text-[var(--color-text-muted)]">Người gửi</dt>
			<dd class="truncate text-right">{@render highlight(item.sender || '—')}</dd>
			<dt class="text-[var(--color-text-muted)]">Nơi an táng</dt>
			<dd class="truncate text-right">{@render highlight(item.burial_place || '—')}</dd>
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
		<table
			class="spirit-data-table table-fixed text-left text-xs"
			style={`width: ${tableWidth}px; min-width: ${tableWidth}px;`}
		>
			<colgroup>
				{#if canWrite}<col style="width: 72px;" />{/if}
				{#each tableSpiritColumns as column (column.key)}<col
						style={`width: ${columnWidth(column)}px;`}
					/>{/each}
			</colgroup>
			<thead class="bg-[var(--color-surface-muted)] text-[var(--color-text-secondary)]">
				<tr>
					{#if canWrite}<th
							class="sticky top-0 z-10 w-18 bg-[var(--color-surface-muted)] px-2 py-2 text-right font-semibold"
							>Thao tác</th
						>{/if}
					{#each tableSpiritColumns as column (column.key)}<th
							class="relative sticky top-0 z-10 bg-[var(--color-surface-muted)] px-2 py-2"
						>
							{#if spiritSortKeys.has(column.key)}{@render spiritSortHeader(column.label, column.key)}{:else}<span class="block truncate font-semibold">{column.label}</span>{/if}<button
								type="button"
								onpointerdown={(event) => beginColumnResize(event, column)}
								onclick={(event) => event.stopPropagation()}
								class="absolute top-0 right-0 z-30 h-full w-4 cursor-col-resize touch-none select-none before:absolute before:top-1/4 before:right-1 before:h-1/2 before:w-px before:bg-[var(--color-border-strong)] hover:before:bg-[var(--color-primary)]"
								aria-label={`Kéo để đổi độ rộng cột ${column.label}`}
								title="Kéo để đổi độ rộng"
							></button>
						</th>{/each}
				</tr>
			</thead>
			<tbody class="divide-y divide-[var(--color-border)]">
				{#each tableSpiritGroups as group (group.key)}
					{#each group.items as item, itemIndex (item.id)}<tr
							ondblclick={(event) => openSpiritEditorFromRow(event, item)}
							class={['hover:bg-[var(--color-primary-soft)]/40', canWrite && 'cursor-default']}
						>
							{#if canWrite && itemIndex === 0}<td rowspan={group.items.length} class="bg-[var(--color-surface-muted)]/35 px-2 py-1 align-middle"
									><div class="flex flex-col items-center gap-1">
										{#if group.hasTablet}<span class="max-w-[68px] truncate text-[10px] font-bold text-[var(--color-primary-dark)]" title={item.tablet_code}>{item.tablet_code}</span>{/if}<div class="flex justify-center gap-1">
										<button
											type="button"
											onclick={() => void edit(item)}
											class="grid h-8 w-8 cursor-pointer place-items-center rounded border border-[var(--color-border-strong)]"
											aria-label={group.hasTablet ? `Sửa Bài vị ${item.tablet_name}` : `Sửa ${item.full_name}`}
											><span class="icon-[lucide--pencil] h-3.5 w-3.5" aria-hidden="true"></span></button
										><button
											type="button"
											onclick={() => void removeTablet(item)}
											class="grid h-8 w-8 cursor-pointer place-items-center rounded border border-[var(--color-danger)] text-[var(--color-danger)]"
											aria-label={group.hasTablet ? `Xóa Bài vị ${item.tablet_name}` : `Xóa ${item.full_name}`}
											><span class="icon-[lucide--trash-2] h-3.5 w-3.5" aria-hidden="true"></span></button
										></div>
									</div>
									</td
								>{/if}
							{#each tableSpiritColumns as column (column.key)}
								{#if (column.key === 'position_name' && group.hasPosition) || (group.hasTablet && (column.key === 'tablet_name' || column.key === 'tablet_image_url' || column.key === 'tablet_registered_at' || column.key === 'tablet_sender' || column.key === 'house_name'))}
									{#if itemIndex === 0}<td
											rowspan={group.items.length}
											class="bg-[var(--color-surface-muted)]/35 px-3 py-2 align-middle font-medium"
										>
											{#if column.key === 'tablet_image_url'}<button
												type="button"
												onclick={() => canWrite && void edit(item)}
												disabled={!canWrite}
												class="h-[100px] w-[70px] text-left enabled:cursor-pointer focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-primary)] disabled:cursor-default"
												aria-label={`Mở bài vị ${item.tablet_name}`}
												><CardBaiVi
													code={item.position_name || 'Chưa xếp'}
													spiritCount={group.items.length}
													spiritNames={group.items.map((spirit) => spirit.full_name)}
													imageUrl={tabletDisplayImage(group.items)}
																																		birthYear={item.birth_year}
																																		deathYear={item.death_year}
																																		tabletType={item.tablet_type}
																																		status={item.status === 'draft' ? 'pending' : 'enshrined'}
													fontSize={9}
													codeScale={1.35}
												/></button
												>{:else}<span class="block w-full truncate" title={item[column.key] || ''}
													>{@render highlight(item[column.key] || '—')}</span
												>{/if}
										</td>{/if}
								{:else}<td class="relative px-3 py-2 align-middle">
										{#if column.key === 'image_url'}
											{#if item.image_url}<button
													type="button"
													onclick={() => openSpiritImage(item)}
													class="cursor-pointer focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--color-primary)]"
													aria-label={`Xem ảnh ${item.full_name}`}
															><SpiritPortrait
																imageUrl={item.image_url}
																alt={item.full_name}
																sizeClass="h-9 w-6"
																bare
															/></button
														>{:else}<span
															class="grid h-9 w-6 place-items-center border border-amber-400 text-amber-500/45"
													aria-label={`Chưa có ảnh ${item.full_name}`}
													><span class="icon-[lucide--user-round] h-5 w-5" aria-hidden="true"
													></span></span
												>{/if}
										{:else if column.key === 'has_urn'}<span class="block text-center font-medium">
											{item.has_urn ? 'Có' : 'Không'}
										</span>
										{:else if column.key === 'created_at' || column.key === 'updated_at'}<span
												class="block w-full truncate whitespace-nowrap"
												>{@render highlight(formatTimestamp(item[column.key]))}</span
											>{:else if canWrite && editablePatchKeys.has(column.key)}<button
												type="button"
												onclick={() => beginCellEdit(item, column)}
												class="block w-full cursor-pointer truncate rounded-sm text-left transition-colors hover:bg-[var(--color-primary-soft)]"
												title={item[column.key] || ''}
												>{@render highlight(item[column.key] || '—')}</button
											>{:else}<span class="block w-full truncate" title={item[column.key] || ''}
												>{@render highlight(item[column.key] || '—')}</span
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
									</td>{/if}{/each}
							{#if false && canWrite}<td class="px-3 py-2 align-middle">
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
				{/each}
			</tbody>
		</table>
		{@render loadMoreButton()}
	</div>{/snippet}

{#snippet loadMoreButton()}{#if hasMore}<div class="flex justify-center py-4">
			<button
				type="button"
				disabled={loadingMore}
				onclick={() => void loadMore()}
				class="h-10 rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] px-4 text-sm font-semibold hover:bg-[var(--color-surface-muted)] disabled:opacity-50"
				>{loadingMore ? 'Đang tải...' : 'Tải thêm'}</button
			>
		</div>{/if}{/snippet}

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

<style>
	.spirit-data-table :global(th) {
		padding: .25rem;
	}

	.spirit-data-table :global(td) {
		padding: .25rem .375rem;
	}
</style>
