<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/auth/auth-store.svelte';
	import { router } from '$lib/navigation/router.svelte';
	import { popupStore } from '$lib/ui/popup-store.svelte';
	import { toastStore } from '$lib/ui/toast-store.svelte';
	import LoadingIndicator from '$lib/ui/LoadingIndicator.svelte';
	import Popup from '$lib/ui/Popup.svelte';
	import InlineSpiritEditor from '$lib/memorial/InlineSpiritEditor.svelte';
	import UnplacedSpiritPicker from '$lib/memorial/UnplacedSpiritPicker.svelte';
	import InlinePositionEditor from '$lib/memorial/InlinePositionEditor.svelte';
	import PositionMap from '$lib/memorial/PositionMap.svelte';
	import SpiritPortrait from '$lib/memorial/SpiritPortrait.svelte';
	import { memorialRevisionStore } from '$lib/memorial/memorial-revision-store.svelte';
	import { houseFilter } from '$lib/memorial/house-filter.svelte';
	import {
		createArea,
		createHouse,
		deleteArea,
		deleteHouse,
		createPositions,
		createTablet,
		deletePosition,
		deleteTablet,
		listAreas,
		listHouses,
		listPositions,
		listTabletSpirits,
		listTablets,
		listUnplacedTablets,
		moveTablet,
		searchPositions,
		updatePosition,
		updateArea,
		updateHouse,
		updateTablet,
		type EditableSpiritInput,
		type Area,
		type House,
		type Position,
		type Spirit,
		type Tablet
	} from '$lib/memorial/api';
	import { emptyInlineSpirit } from '$lib/memorial/sheet-parser';
	import { emptyPositionRow, type EditablePositionRow } from '$lib/memorial/position-sheet-parser';

	type Mode = 'house' | 'area' | 'position' | 'tablet' | '';
	type PositionView = 'table' | 'map';
	type CachedPositionTablets = { tablets: Tablet[]; previews: Map<string, Spirit[]> };
	type PositionSortKey =
		| 'house_name'
		| 'area_code'
		| 'name'
		| 'row_number'
		| 'column_number'
		| 'tablet_count'
		| 'spirit_count';
	const positionSortKeys: PositionSortKey[] = [
		'house_name',
		'area_code',
		'name',
		'row_number',
		'column_number',
		'tablet_count',
		'spirit_count'
	];
	const positionSortStorageKey = 'nhalinh:position-sort';
	let houses = $state<House[]>([]),
		areas = $state<Area[]>([]),
		positions = $state<Position[]>([]),
		tablets = $state<Tablet[]>([]),
		unplacedTablets = $state<Tablet[]>([]);
	let houseId = $state(''),
		areaId = $state(''),
		positionId = $state(''),
		loading = $state(true),
		saving = $state(false),
		inlineEditorBusy = $state(false),
		drawerOpen = $state(false),
		drawerLoading = $state(false),
		positionQuery = $state(''),
		positionSortKey = $state<PositionSortKey>('column_number'),
		positionSortDirection = $state<'asc' | 'desc'>('asc'),
		positionView = $state<PositionView>('map'),
		positionFullscreen = $state(false),
		deleteTabletOpen = $state(false),
		deleteTabletSpirits = $state(false),
		deletePositionOpen = $state(false),
		unplacedTabletsOpen = $state(false),
		unplacedTabletQuery = $state(''),
		assignmentTarget = $state<Position | null>(null),
		movingTablet = $state<Tablet | null>(null),
		moveTabletOpen = $state(false),
		movePositionQuery = $state(''),
		movePositionResults = $state<Position[]>([]),
		houseManagerOpen = $state(false),
		areaManagerOpen = $state(false),
		formSnapshot = $state(''),
		mode = $state<Mode>('');
	let editingHouse = $state<House | null>(null);
	let editingArea = $state<Area | null>(null);
	let editingPosition = $state<Position | null>(null);
	let newPositions = $state<EditablePositionRow[]>([emptyPositionRow()]);
	let drawerPosition = $state<Position | null>(null);
	let editingTablet = $state<Tablet | null>(null);
	let tabletNameInput = $state<HTMLInputElement | null>(null);
	let tabletToDelete = $state<Tablet | null>(null);
	let tabletSpiritPreviews = $state<Map<string, Spirit[]>>(new Map());
	let positionTabletCache = $state<Map<string, CachedPositionTablets>>(new Map());
	let drawerRequest = 0;
	let seenMemorialRevision = memorialRevisionStore.revision;
	let memorialDataStale = $state(false);
	let selectedUnplacedSpirits = $state<Spirit[]>([]);
	let houseForm = $state({ name: '', address: '', notes: '' }),
		areaForm = $state({ code: '', name: '', notes: '' }),
		positionForm = $state<{
			row_number: number;
			column_number: number;
			notes: string;
		}>({ row_number: 1, column_number: 1, notes: '' }),
		tabletForm = $state<{ name: string; code: string; registered_at: string; enshrined_at: string; entered_worship_area_at: string; image_url: string; sender: string; notes: string; status: Tablet['status']; type: Tablet['type']; spirits: EditableSpiritInput[] }>({
			name: '', code: '', registered_at: '', enshrined_at: '', entered_worship_area_at: '',
			image_url: '', sender: '', notes: '', status: 'pending', type: 'spirit',
			spirits: [emptyInlineSpirit()]
		});
	let house = $derived(houses.find((v) => v.id === houseId));
	let canWrite = $derived(authStore.user?.role === 'admin' || house?.access_role === 'editor');
	let selectedPosition = $derived(
		drawerPosition ?? positions.find((position) => position.id === positionId)
	);
	let filteredUnplacedTablets = $derived(
		unplacedTablets.filter((tablet) =>
			normalizeSearch(tablet.name).includes(normalizeSearch(unplacedTabletQuery))
		)
	);
	$effect(() => {
		if (mode !== 'tablet' || editingTablet || !tabletNameInput) return;
		const frame = requestAnimationFrame(() => {
			tabletNameInput?.focus();
			tabletNameInput?.setSelectionRange(tabletForm.name.length, tabletForm.name.length);
		});
		return () => cancelAnimationFrame(frame);
	});
	let filteredPositions = $derived.by(() => {
		const query = normalizeSearch(positionQuery);
		const items = positions.filter((position) =>
			normalizeSearch(
				`${position.name} ${position.notes} ${position.area_code} ${position.house_name} ${position.row_number} ${position.column_number}`
			).includes(query)
		);
		return items.sort((left, right) => {
			const leftValue = left[positionSortKey];
			const rightValue = right[positionSortKey];
			const compared =
				typeof leftValue === 'number' && typeof rightValue === 'number'
					? leftValue - rightValue
					: String(leftValue).localeCompare(String(rightValue), 'vi', {
							numeric: true,
							sensitivity: 'base'
						});
			return positionSortDirection === 'asc' ? compared : -compared;
		});
	});

	onMount(() => {
		restorePositionSort();
		restorePositionView();
		const onHouseSelect = (event: Event) => { houseId = (event as CustomEvent<string>).detail; void selectHouse(); };
		window.addEventListener('memorial-house-select', onHouseSelect);
		void init();
		return () => window.removeEventListener('memorial-house-select', onHouseSelect);
	});
	$effect(() => {
		const revision = memorialRevisionStore.revision;
		if (revision === seenMemorialRevision) return;
		seenMemorialRevision = revision;
		positionTabletCache = new Map();
		tablets = [];
		tabletSpiritPreviews = new Map();
		memorialDataStale = true;
	});
	$effect(() => {
		if (router.current.name !== 'structure' || !memorialDataStale || !houseId) return;
		memorialDataStale = false;
		void refreshAfterMemorialChange();
	});
	async function init() {
		loading = true;
		try {
			houses = await listHouses();
			houseId = houseFilter.id || houses[0]?.id || '';
			await selectHouse();
		} catch (e) {
			toastStore.error(msg(e));
		} finally {
			loading = false;
		}
	}
	async function selectHouse() {
		areas = houseId ? await listAreas(houseId) : [];
		await refreshUnplacedTablets();
		areaId = areas[0]?.id ?? '';
		await selectArea();
	}
	async function refreshUnplacedTablets() {
		unplacedTablets = houseId ? await listUnplacedTablets(houseId) : [];
	}
	async function refreshAfterMemorialChange() {
		try {
			areas = await listAreas(houseId);
			if (!areas.some((area) => area.id === areaId)) areaId = areas[0]?.id ?? '';
			positions = areaId ? await listPositions(areaId) : [];
			if (!positions.some((position) => position.id === positionId)) {
				positionId = positions[0]?.id ?? '';
			}
			drawerPosition = null;
			await refreshUnplacedTablets();
		} catch (e) {
			toastStore.error(msg(e));
		}
	}
	async function selectArea() {
		positionQuery = '';
		positions = areaId ? await listPositions(areaId) : [];
		positionId = positions[0]?.id ?? '';
		drawerPosition = null;
		await selectPosition();
	}
	function invalidatePositionTabletCache(...ids: string[]) {
		const next = new Map(positionTabletCache);
		for (const id of ids) next.delete(id);
		positionTabletCache = next;
	}
	function cachePositionTablets(id: string, data: CachedPositionTablets) {
		const next = new Map(positionTabletCache);
		next.delete(id);
		next.set(id, data);
		while (next.size > 24) next.delete(next.keys().next().value!);
		positionTabletCache = next;
	}
	async function selectPosition(force = false) {
		const selectedID = positionId;
		if (!drawerOpen || !selectedID) {
			tablets = [];
			tabletSpiritPreviews = new Map();
			return;
		}
		const cached = positionTabletCache.get(selectedID);
		if (!force && cached) {
			tablets = cached.tablets;
			tabletSpiritPreviews = cached.previews;
			return;
		}
		const request = ++drawerRequest;
		const items = await listTablets(selectedID);
		const previews = new Map(
			await Promise.all(
				items.map(async (tablet) => [tablet.id, await listTabletSpirits(tablet.id)] as const)
			)
		);
		cachePositionTablets(selectedID, { tablets: items, previews });
		if (request !== drawerRequest || !drawerOpen || positionId !== selectedID) return;
		tablets = items;
		tabletSpiritPreviews = previews;
	}
	function openTabletDelete(tablet: Tablet) {
		tabletToDelete = tablet;
		deleteTabletSpirits = false;
		deleteTabletOpen = true;
	}
	async function confirmTabletDelete() {
		if (!tabletToDelete) return;
		saving = true;
		try {
			await deleteTablet(tabletToDelete.id, deleteTabletSpirits);
			toastStore.success(deleteTabletSpirits ? 'Đã xoá Bài vị và Hương linh' : 'Đã xoá Bài vị; Hương linh được đưa về chưa xếp');
			deleteTabletOpen = false;
			invalidatePositionTabletCache(positionId);
			await selectPosition(true);
			await refreshUnplacedTablets();
		} catch (e) { toastStore.error(msg(e)); } finally { saving = false; }
	}
	function openUnplacedTablets(target: Position | null = null) {
		assignmentTarget = target;
		unplacedTabletQuery = '';
		unplacedTabletsOpen = true;
	}
	async function assignUnplacedTablet(tablet: Tablet) {
		if (!assignmentTarget) return;
		saving = true;
		try {
			await moveTablet(tablet.id, assignmentTarget.id);
			toastStore.success(`Đã thêm ${tablet.name} vào ${assignmentTarget.name}`);
			await refreshUnplacedTablets();
			invalidatePositionTabletCache(assignmentTarget.id);
			await selectPosition(true);
			unplacedTabletsOpen = false;
		} catch (e) {
			toastStore.error(msg(e));
		} finally {
			saving = false;
		}
	}
	function openTabletMove(tablet: Tablet) {
		movingTablet = tablet;
		movePositionQuery = '';
		movePositionResults = [];
		moveTabletOpen = true;
	}
	async function findMovePositions() {
		if (!houseId || !movePositionQuery.trim()) {
			movePositionResults = [];
			return;
		}
		try {
			movePositionResults = await searchPositions(houseId, movePositionQuery, 20);
		} catch (e) {
			toastStore.error(msg(e));
		}
	}
	async function confirmTabletMove(target: Position) {
		if (!movingTablet) return;
		saving = true;
		try {
			await moveTablet(movingTablet.id, target.id);
			toastStore.success(`Đã chuyển ${movingTablet.name} sang ${target.name}`);
			moveTabletOpen = false;
			invalidatePositionTabletCache(movingTablet.position_id, target.id);
			await selectPosition(true);
			await refreshUnplacedTablets();
		} catch (e) {
			toastStore.error(msg(e));
		} finally {
			saving = false;
		}
	}
	async function confirmPositionDelete() {
		if (!editingPosition) return;
		saving = true;
		try {
			await deletePosition(editingPosition.id);
			toastStore.success('Đã xoá vị trí; các Bài vị được giữ lại ở danh sách chưa xếp');
			deletePositionOpen = false;
			mode = '';
			drawerOpen = false;
			invalidatePositionTabletCache(editingPosition.id);
			await refreshUnplacedTablets();
			await selectArea();
		} catch (e) {
			toastStore.error(msg(e));
		} finally {
			saving = false;
		}
	}
	async function openPositionDrawer(position: Position) {
		positionId = position.id;
		drawerPosition = position;
		if (canWrite && position.tablet_count === 0) {
			drawerOpen = false;
			await open('tablet');
			return;
		}
		drawerOpen = true;
		const cached = positionTabletCache.get(position.id);
		if (cached) {
			tablets = cached.tablets;
			tabletSpiritPreviews = cached.previews;
			drawerLoading = false;
			return;
		}
		tablets = [];
		tabletSpiritPreviews = new Map();
		drawerLoading = true;
		try {
			await selectPosition();
		} catch (e) {
			toastStore.error(msg(e));
		} finally {
			drawerLoading = false;
		}
	}
	async function openPositionAtEmptyCoordinate(rowNumber: number, columnNumber: number) {
		await open('position');
		newPositions = [{ row_number: String(rowNumber), column_number: String(columnNumber), notes: '' }];
		captureFormSnapshot();
	}
	async function fillEmptyCoordinates(maxRow: number, maxColumn: number) {
		if (!areaId) return;
		const occupied = new Set(positions.map((position) => `${position.row_number}:${position.column_number}`));
		const missing: Array<{ row_number: number; column_number: number; notes: string }> = [];
		for (let row = 1; row <= maxRow; row += 1) {
			for (let column = 1; column <= maxColumn; column += 1) {
				if (!occupied.has(`${row}:${column}`)) missing.push({ row_number: row, column_number: column, notes: '' });
			}
		}
		if (missing.length === 0) {
			toastStore.success('Tất cả tọa độ trong sơ đồ đã có mã vị trí');
			return;
		}
		const confirmed = await popupStore.confirm({
			title: `Tạo ${missing.length} mã vị trí?`,
			message: `Các tọa độ trống trong phạm vi cột 1–${maxColumn}, hàng 1–${maxRow} sẽ được tạo thành Vị trí thật.`,
			confirmLabel: 'Tạo mã vị trí',
			cancelLabel: 'Hủy'
		});
		if (!confirmed) return;
		saving = true;
		try {
			await createPositions(areaId, missing);
			await selectArea();
			memorialRevisionStore.invalidate();
			toastStore.success(`Đã tạo ${missing.length} mã vị trí`);
		} catch (error) {
			toastStore.error(msg(error));
		} finally {
			saving = false;
		}
	}
	function currentFormState() {
		switch (mode) {
			case 'house': return { houseForm };
			case 'area': return { areaForm };
			case 'position': return editingPosition ? { positionForm } : { newPositions };
			case 'tablet': return { tabletForm, selectedUnplacedSpiritIDs: selectedUnplacedSpirits.map((spirit) => spirit.id) };
			default: return {};
		}
	}
	function captureFormSnapshot() { formSnapshot = JSON.stringify(currentFormState()); }
	function hasUnsavedFormChanges() { return formSnapshot !== JSON.stringify(currentFormState()); }
	async function open(next: Mode) {
		mode = next;
		if (next !== 'house') editingHouse = null;
		if (next !== 'area') editingArea = null;
		if (next !== 'position') editingPosition = null;
		if (next !== 'tablet') editingTablet = null;
		if (next === 'house') houseForm = { name: '', address: '', notes: '' };
		if (next === 'area') areaForm = { code: '', name: '', notes: '' };
		if (next === 'position') {
			editingPosition = null;
			newPositions = [emptyPositionRow()];
			positionForm = {
				row_number: 1,
				column_number: 1,
				notes: ''
			};
		}
		if (next === 'tablet') {
			editingTablet = null;
			selectedUnplacedSpirits = [];
			tabletForm = { name: '', code: '', registered_at: '', enshrined_at: '', entered_worship_area_at: '', image_url: '', sender: '', notes: '', status: 'pending', type: 'spirit', spirits: [emptyInlineSpirit()] };
		}
		captureFormSnapshot();
	}
	function editPosition(position: Position) {
		positionId = position.id;
		editingPosition = position;
		positionForm = {
			row_number: position.row_number,
			column_number: position.column_number,
			notes: position.notes
		};
		mode = 'position';
		captureFormSnapshot();
	}
	async function editTablet(tablet: Tablet) {
		try {
			const items = await listTabletSpirits(tablet.id);
			positionId = tablet.position_id;
			editingTablet = tablet;
			selectedUnplacedSpirits = [];
			tabletForm = {
				name: tablet.name, code: tablet.code || tablet.name, registered_at: tablet.registered_at, enshrined_at: tablet.enshrined_at, entered_worship_area_at: tablet.entered_worship_area_at,
				image_url: tablet.image_url, sender: tablet.sender, notes: tablet.notes, status: tablet.status, type: tablet.type,
				spirits: items.map((spirit) => ({
					id: spirit.id,
					full_name: spirit.full_name,
					dharma_name: spirit.dharma_name,
					birth_year: spirit.birth_year,
					death_year: spirit.death_year,
					age: spirit.age,
					image_url: spirit.image_url,
					burial_place: spirit.burial_place,
					sender: spirit.sender,
					sent_month: spirit.sent_month,
					notes: spirit.notes
				}))
			};
			mode = 'tablet';
			captureFormSnapshot();
		} catch (e) {
			toastStore.error(msg(e));
		}
	}
	async function save(e: SubmitEvent) {
		e.preventDefault();
		if (mode === 'tablet' && editingTablet && !canWrite) return;
		saving = true;
		try {
			let successMessage = 'Đã lưu dữ liệu';
			if (mode === 'house') {
				const v = editingHouse
					? await updateHouse(editingHouse.id, { ...houseForm, active: editingHouse.active })
					: await createHouse(houseForm);
				houses = await listHouses();
				houseId = v.id;
				await selectHouse();
			} else if (mode === 'area') {
				const v = editingArea
					? await updateArea(editingArea.id, areaForm)
					: await createArea({ house_id: houseId, ...areaForm });
				areas = await listAreas(houseId);
				areaId = v.id;
				await selectArea();
			} else if (mode === 'position') {
				let selectedID = '';
				if (editingPosition) {
					const updated = await updatePosition(editingPosition.id, {
						area_id: areaId,
						row_number: positionForm.row_number,
						column_number: positionForm.column_number,
						notes: positionForm.notes
					});
					selectedID = updated.id;
				} else {
					const rows = newPositions.filter((row) =>
						Object.values(row).some((value) => value.trim())
					);
					const created = await createPositions(
						areaId,
						rows.map((row) => ({
							row_number: Number(row.row_number),
							column_number: Number(row.column_number),
							notes: row.notes
						}))
					);
					selectedID = created.positions[0]?.id ?? '';
					if (created.skipped_count > 0) {
						successMessage = created.positions.length
							? `Đã thêm ${created.positions.length} vị trí, bỏ qua ${created.skipped_count} vị trí đã tồn tại`
							: `Đã bỏ qua ${created.skipped_count} vị trí đã tồn tại`;
					}
				}
				positions = await listPositions(areaId);
				positionId = selectedID;
				await selectPosition();
				editingPosition = null;
			} else if (mode === 'tablet') {
				const targetPositionID = positionId;
				const spirits = tabletForm.spirits.filter((row) =>
					Object.values(row).some((value) => value.trim())
				);
				const tabletImage = tabletForm.image_url || (spirits.length === 1 ? spirits[0].image_url : '');
				const payload = {
					position_id: targetPositionID,
					name: tabletForm.code, code: tabletForm.code, registered_at: tabletForm.registered_at, enshrined_at: tabletForm.enshrined_at, entered_worship_area_at: tabletForm.entered_worship_area_at,
					image_url: tabletImage, sender: tabletForm.sender, notes: tabletForm.notes, status: tabletForm.status, type: tabletForm.type,
					spirits
				};
				if (editingTablet) await updateTablet(editingTablet.id, payload);
				else {
					await createTablet({
						...payload,
						existing_spirit_ids: selectedUnplacedSpirits.map((spirit) => spirit.id)
					});
				}
				invalidatePositionTabletCache(targetPositionID);
				await selectPosition(true);
				areas = await listAreas(houseId);
				positions = await listPositions(areaId);
				editingTablet = null;
			}
			toastStore.success(successMessage);
			mode = '';
		} catch (err) {
			toastStore.error(msg(err));
		} finally {
			saving = false;
		}
	}
	function msg(e: unknown) {
		return e instanceof Error ? e.message : 'Có lỗi xảy ra';
	}
	function normalizeSearch(value: string) {
		return value
			.normalize('NFD')
			.replace(/[\u0300-\u036f]/g, '')
			.toLowerCase()
			.replace(/-/g, '')
			.trim();
	}
	function sortPositions(key: PositionSortKey) {
		if (positionSortKey === key) {
			positionSortDirection = positionSortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			positionSortKey = key;
			positionSortDirection = 'asc';
		}
		try {
			localStorage.setItem(
				positionSortStorageKey,
				JSON.stringify({ key: positionSortKey, direction: positionSortDirection })
			);
		} catch {
			// Sorting still works when browser storage is unavailable.
		}
	}
	function restorePositionSort() {
		try {
			const raw = localStorage.getItem(positionSortStorageKey);
			if (!raw) return;
			const saved = JSON.parse(raw) as { key?: string; direction?: string };
			if (
				positionSortKeys.includes(saved.key as PositionSortKey) &&
				(saved.direction === 'asc' || saved.direction === 'desc')
			) {
				positionSortKey = saved.key as PositionSortKey;
				positionSortDirection = saved.direction;
			}
		} catch {
			localStorage.removeItem(positionSortStorageKey);
		}
	}
	function setPositionView(view: PositionView) {
		positionView = view;
		try {
			localStorage.setItem('nhalinh:position-view', view);
		} catch {
			// The view still changes without browser storage.
		}
	}
	function restorePositionView() {
		try {
			const saved = localStorage.getItem('nhalinh:position-view');
			if (saved === 'table' || saved === 'map') positionView = saved;
		} catch {
			// Keep the default map view when browser storage is unavailable.
		}
	}
	function sortAria(key: PositionSortKey): 'ascending' | 'descending' | 'none' {
		if (positionSortKey !== key) return 'none';
		return positionSortDirection === 'asc' ? 'ascending' : 'descending';
	}
	function closeActiveOverlay(event: KeyboardEvent) {
		if (event.key !== 'Escape') return;
		if (document.querySelector('[role="dialog"][aria-modal="true"]')) return;
		if (mode) {
			if (!saving && !inlineEditorBusy) void requestCloseMode();
			return;
		}
		if (drawerOpen) drawerOpen = false;
		else if (positionFullscreen) positionFullscreen = false;
	}
	function tabletTypeLabel(type: Tablet['type']) {
		return ({ spirit: 'Hương linh', giac_linh: 'Giác linh', family: 'Gia tiên', cuu_huyen: 'Cửu Huyền Thất Tổ', clan: 'Tộc họ', collective: 'Chư Hương linh', fetus: 'Thai nhi', martyr: 'Anh hùng liệt sĩ', victim: 'Hương linh tử nạn', childless: 'Hương linh vô tự', unknown: 'Hương linh vô danh', other: 'Khác' })[type];
	}
	function tabletStatusLabel(status: Tablet['status']) {
		return status === 'pending' ? 'Chờ an vị' : status === 'enshrined' ? 'Đã an vị' : 'Đã thỉnh về';
	}
	function editHouse(item: House) {
		editingHouse = item; houseForm = { name: item.name, address: item.address, notes: item.notes }; houseManagerOpen = false; mode = 'house'; captureFormSnapshot();
	}
	function editArea(item: Area) {
		editingArea = item; areaForm = { code: item.code, name: item.name, notes: item.notes }; areaManagerOpen = false; mode = 'area'; captureFormSnapshot();
	}
	async function removeHouse(item: House) {
		if (!await popupStore.confirm({ title: 'Xóa Nhà Linh?', message: `Nhà Linh ${item.name} chỉ có thể xóa khi không còn Khu vực, Bài vị hay Hương linh liên quan.`, confirmLabel: 'Xóa', cancelLabel: 'Hủy', tone: 'danger' })) return;
		try { await deleteHouse(item.id); houses = await listHouses(); houseId = houses[0]?.id ?? ''; await selectHouse(); toastStore.success('Đã xóa Nhà Linh'); } catch (e) { toastStore.error(msg(e)); }
	}
	async function removeArea(item: Area) {
		if (!await popupStore.confirm({ title: 'Xóa Khu vực?', message: `Khu ${item.code} chỉ có thể xóa khi không còn Vị trí.`, confirmLabel: 'Xóa', cancelLabel: 'Hủy', tone: 'danger' })) return;
		try { await deleteArea(item.id); areas = await listAreas(houseId); areaId = areas[0]?.id ?? ''; await selectArea(); toastStore.success('Đã xóa Khu vực'); } catch (e) { toastStore.error(msg(e)); }
	}
	async function requestCloseMode() {
		if (!mode || saving || inlineEditorBusy) return;
		if (mode === 'tablet' && editingTablet && !canWrite) {
			mode = '';
			return;
		}
		if (!hasUnsavedFormChanges()) {
			mode = '';
			return;
		}
		const confirmed = await popupStore.confirm({
			title: 'Bỏ thay đổi chưa lưu?',
			message: 'Các thông tin đang nhập sẽ không được lưu.',
			confirmLabel: 'Bỏ thay đổi',
			cancelLabel: 'Tiếp tục nhập',
			tone: 'danger'
		});
		if (confirmed) mode = '';
	}
</script>

<svelte:window onkeydown={closeActiveOverlay} />

<section
	class="h-full overflow-y-auto px-4 py-4 md:px-6 lg:flex lg:min-w-0 lg:flex-col lg:overflow-hidden lg:px-8"
>
	<div class="mx-auto w-full max-w-[1320px] lg:flex lg:min-h-0 lg:flex-1 lg:flex-col">
		<div class="mb-5 flex shrink-0 flex-wrap gap-2">
			{#if authStore.user?.role === 'admin'}<button type="button" onclick={() => (houseManagerOpen = true)} class="h-11 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white shadow-sm"><span class="mr-1.5 icon-[lucide--building-2] inline-block h-4 w-4 align-text-bottom"></span>Quản lý Nhà Linh</button>{/if}
			{#if canWrite && houseId}<button type="button" onclick={() => (areaManagerOpen = true)} class="h-11 rounded-md border border-[var(--color-primary)] bg-[var(--color-primary-soft)] px-4 text-sm font-semibold text-[var(--color-primary-dark)]"><span class="mr-1.5 icon-[lucide--map] inline-block h-4 w-4 align-text-bottom"></span>Quản lý Khu vực</button>{/if}
			{#if houseId}<button type="button" onclick={() => openUnplacedTablets()} class="h-11 rounded-md border border-dashed border-[var(--color-border-strong)] bg-[var(--color-surface)] px-4 text-sm font-semibold text-[var(--color-primary-dark)]"><span class="mr-1.5 icon-[lucide--list] inline-block h-4 w-4 align-text-bottom"></span>Bài vị chưa xếp ({unplacedTablets.length})</button>{/if}
		</div>
		{#if loading}<div class="py-16">
				<LoadingIndicator label="Đang tải cơ cấu tổ chức..." />
			</div>{:else if houses.length === 0}<div
				class="rounded-md border border-dashed py-16 text-center"
			>
				<p class="font-semibold">Chưa có Nhà Linh</p>
			</div>{:else}<div class="flex min-h-0 flex-1 flex-col gap-5">
				<aside>
					<div class="mb-2 flex items-center justify-between">
						<h2 class="text-sm font-semibold">Khu vực</h2>
					</div>
					<div class="flex gap-2 overflow-x-auto pb-1">
						{#each areas as area (area.id)}<button
								type="button"
								onclick={() => {
									areaId = area.id;
									void selectArea();
								}}
								class={[
									'w-[138px] shrink-0 rounded-md border px-2.5 py-2 text-left',
									areaId === area.id
										? 'border-[var(--color-primary)] bg-[var(--color-primary-soft)]'
										: 'border-[var(--color-border)] bg-[var(--color-surface)]'
								]}
								><span class="text-sm font-semibold">Khu {area.code}</span>{#if area.name}<span
										class="ml-1 text-xs">· {area.name}</span
									>{/if}<span class="mt-0.5 block text-[11px] text-[var(--color-text-secondary)]"
									>{area.position_count} vị trí · {area.tablet_count} bài vị</span
								></button
							>{/each}
					</div>
				</aside>
				<main
					class={[
						'min-w-0 lg:flex lg:min-h-0 lg:flex-col',
						positionFullscreen &&
							'fixed inset-0 z-30 flex min-h-0 flex-col overflow-hidden bg-[var(--color-bg)] p-4 sm:p-6'
					]}
				>
					<div class="mb-3 flex shrink-0 items-center justify-between">
						<div>
							<h2 class="font-semibold">Vị trí</h2>
							<p class="text-xs text-[var(--color-text-secondary)]">
								Mỗi vị trí có thể chứa nhiều bài vị.
							</p>
						</div>
						<div class="flex items-center gap-2">
							<div class="flex rounded-md bg-[var(--color-surface-muted)] p-1">
								<button
									type="button"
									onclick={() => setPositionView('table')}
									class={[
										'grid h-8 w-8 place-items-center rounded',
										positionView === 'table'
											? 'bg-[var(--color-surface)] text-[var(--color-primary-dark)] shadow-sm'
											: 'text-[var(--color-text-muted)]'
									]}
									aria-label="Xem vị trí dạng bảng"
									><span class="icon-[lucide--table-2] h-4 w-4" aria-hidden="true"></span></button
								><button
									type="button"
									onclick={() => setPositionView('map')}
									class={[
										'grid h-8 w-8 place-items-center rounded',
										positionView === 'map'
											? 'bg-[var(--color-surface)] text-[var(--color-primary-dark)] shadow-sm'
											: 'text-[var(--color-text-muted)]'
									]}
									aria-label="Xem vị trí dạng sơ đồ"
									><span class="icon-[lucide--map] h-4 w-4" aria-hidden="true"></span></button
								>
							</div>
							{#if areaId}<button
									type="button"
									onclick={() => (positionFullscreen = !positionFullscreen)}
									aria-label={positionFullscreen
										? 'Thu nhỏ màn hình vị trí'
										: 'Phóng to toàn màn hình vị trí'}
									title={positionFullscreen ? 'Thu nhỏ' : 'Phóng to toàn màn hình'}
									class="grid h-10 w-10 shrink-0 place-items-center rounded-md border border-[var(--color-border-strong)] bg-[var(--color-surface)] text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)] hover:text-[var(--color-primary-dark)]"
								>
									<span
										class={[
											'h-4 w-4',
											positionFullscreen ? 'icon-[lucide--minimize-2]' : 'icon-[lucide--maximize-2]'
										]}
										aria-hidden="true"
									></span>
								</button>{/if}
							{#if canWrite && areaId}<button
									type="button"
									onclick={() => void open('position')}
									class="h-10 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white"
									>Thêm vị trí</button
								>{/if}
						</div>
					</div>
					{#if positions.length > 0 && positionView === 'table'}<label
							class="relative mb-3 block shrink-0"
							><span
								class="absolute top-3.5 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"
							></span><input
								bind:value={positionQuery}
								placeholder="Tìm vị trí, ví dụ 1a1..."
								class="h-11 w-full rounded-md border-[var(--color-border-strong)] pl-9 text-sm"
							/></label
						>{/if}
					{#if positions.length === 0}<div
							class="rounded-md border border-dashed py-12 text-center text-sm text-[var(--color-text-secondary)] lg:min-h-0 lg:flex-1"
						>
							Chưa có vị trí trong khu vực này
						</div>{:else if positionView === 'map'}<PositionMap
							{positions}
							fullscreen={positionFullscreen}
							areaCode={areas.find((area) => area.id === areaId)?.code ?? ''}
							onposition={(position) => void openPositionDrawer(position)}
							onemptyposition={canWrite
								? ({ rowNumber, columnNumber }) =>
									void openPositionAtEmptyCoordinate(rowNumber, columnNumber)
								: undefined}
							onfillcoordinates={canWrite ? (bounds) => void fillEmptyCoordinates(bounds.maxRow, bounds.maxColumn) : undefined}
						/>
					{:else}<div
							class={[
								'overflow-auto rounded-md border border-[var(--color-border)] bg-[var(--color-surface)]',
								positionFullscreen
									? 'min-h-0 flex-1'
									: 'max-h-[calc(100dvh-17rem)] min-h-[360px] lg:min-h-0 lg:flex-1'
							]}
						>
							<table
								class="w-full min-w-[920px] border-separate border-spacing-0 text-left text-sm"
							>
								<thead
									class="bg-[var(--color-surface-muted)] text-xs text-[var(--color-text-secondary)]"
									><tr
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('house_name')}
											>{@render sortHeader('Nhà Linh', 'house_name')}</th
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('area_code')}
											>{@render sortHeader('Khu vực', 'area_code')}</th
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('name')}>{@render sortHeader('Tên vị trí', 'name')}</th
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('column_number')}
											>{@render sortHeader('Cột', 'column_number', true)}</th
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('row_number')}
											>{@render sortHeader('Hàng', 'row_number', true)}</th
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('tablet_count')}
											>{@render sortHeader('Bài vị', 'tablet_count', true)}</th
										><th
											class="sticky top-0 z-10 bg-[var(--color-surface-muted)] px-4 py-3"
											aria-sort={sortAria('spirit_count')}
											>{@render sortHeader('Hương linh', 'spirit_count', true)}</th
										>{#if canWrite}<th
												class="sticky top-0 z-10 w-16 bg-[var(--color-surface-muted)] px-4 py-3"
												><span class="sr-only">Thao tác</span></th
											>{/if}</tr
									></thead
								><tbody class="divide-y divide-[var(--color-border)]"
									>{#each filteredPositions as position (position.id)}<tr
											onclick={() => void openPositionDrawer(position)}
											class={[
												'cursor-pointer hover:bg-[var(--color-surface-muted)]',
												positionId === position.id && 'bg-[var(--color-primary-soft)]'
											]}
											><td class="px-4 py-3">{position.house_name}</td><td class="px-4 py-3"
												>Khu {position.area_code}</td
											><td class="px-4 py-3 font-semibold text-[var(--color-primary-dark)]"
												>{position.name}</td
											><td class="px-4 py-3 text-right">{position.column_number}</td><td
												class="px-4 py-3 text-right">{position.row_number}</td
											><td class="px-4 py-3 text-right">{position.tablet_count}</td><td
												class="px-4 py-3 text-right">{position.spirit_count}</td
											>{#if canWrite}<td class="px-4 py-2 text-right"
													><button
														type="button"
														aria-label={`Sửa vị trí ${position.name}`}
														title="Sửa vị trí"
														onclick={(event) => {
															event.stopPropagation();
															editPosition(position);
														}}
														class="grid h-9 w-9 place-items-center rounded-md text-[var(--color-primary-dark)] hover:bg-[var(--color-surface-muted)]"
														><span class="icon-[lucide--pencil] h-4 w-4"></span></button
													></td
												>{/if}</tr
										>{/each}{#if filteredPositions.length === 0}<tr
											><td
												colspan={canWrite ? 8 : 7}
												class="px-4 py-10 text-center text-sm text-[var(--color-text-secondary)]"
												>Không tìm thấy vị trí phù hợp</td
											></tr
										>{/if}</tbody
								>
							</table>
						</div>{/if}
				</main>
			</div>{/if}
	</div>
</section>

<Popup open={houseManagerOpen} title="Danh sách Nhà Linh" onClose={() => (houseManagerOpen = false)}>
	<div class="space-y-3">
		<div class="flex justify-end"><button type="button" onclick={() => { houseManagerOpen = false; void open('house'); }} class="h-10 rounded-md bg-[var(--color-primary)] px-3 text-sm font-semibold text-white"><span class="mr-1 icon-[lucide--plus] inline-block h-4 w-4 align-text-bottom"></span>Thêm Nhà Linh</button></div>
		<div class="space-y-2">
		{#each houses as item (item.id)}<div class="flex items-center gap-3 rounded-md border border-[var(--color-border)] p-3"><div class="min-w-0 flex-1"><p class="truncate font-semibold">{item.name}</p><p class="mt-0.5 truncate text-xs text-[var(--color-text-secondary)]">{item.address || 'Chưa có địa chỉ'} · {item.active ? 'Đang sử dụng' : 'Ngừng sử dụng'}</p></div><button type="button" onclick={() => editHouse(item)} class="h-8 rounded-md border px-2 text-xs font-semibold">Sửa</button><button type="button" onclick={() => void removeHouse(item)} class="h-8 rounded-md border border-[var(--color-danger)] px-2 text-xs font-semibold text-[var(--color-danger)]">Xóa</button></div>{:else}<p class="py-8 text-center text-sm text-[var(--color-text-secondary)]">Chưa có Nhà Linh.</p>{/each}
		</div>
	</div>
	{#snippet footer()}<button type="button" onclick={() => (houseManagerOpen = false)} class="ml-auto h-10 rounded-md border px-5 font-semibold">Đóng</button>{/snippet}
</Popup>

<Popup open={areaManagerOpen} title="Danh sách Khu vực" onClose={() => (areaManagerOpen = false)}>
	<div class="space-y-3">
		<div class="flex justify-end"><button type="button" onclick={() => { areaManagerOpen = false; void open('area'); }} class="h-10 rounded-md bg-[var(--color-primary)] px-3 text-sm font-semibold text-white"><span class="mr-1 icon-[lucide--plus] inline-block h-4 w-4 align-text-bottom"></span>Thêm Khu vực</button></div>
		<div class="space-y-2">
		{#each areas as item (item.id)}<div class="flex items-center gap-3 rounded-md border border-[var(--color-border)] p-3"><div class="min-w-0 flex-1"><p class="truncate font-semibold">Khu {item.code}{item.name ? ` · ${item.name}` : ''}</p><p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">{item.position_count} vị trí · {item.tablet_count} Bài vị · {item.spirit_count} Hương linh</p></div><button type="button" onclick={() => editArea(item)} class="h-8 rounded-md border px-2 text-xs font-semibold">Sửa</button><button type="button" disabled={item.position_count > 0} title={item.position_count > 0 ? 'Cần xóa hết vị trí trước' : 'Xóa khu vực'} onclick={() => void removeArea(item)} class="h-8 rounded-md border border-[var(--color-danger)] px-2 text-xs font-semibold text-[var(--color-danger)] disabled:opacity-40">Xóa</button></div>{:else}<p class="py-8 text-center text-sm text-[var(--color-text-secondary)]">Chưa có Khu vực.</p>{/each}
		</div>
	</div>
	{#snippet footer()}<button type="button" onclick={() => (areaManagerOpen = false)} class="ml-auto h-10 rounded-md border px-5 font-semibold">Đóng</button>{/snippet}
</Popup>

{#if drawerOpen}<div
		class="fixed inset-0 z-40 bg-black/35"
		role="presentation"
		onclick={(event) => {
			if (event.target === event.currentTarget) drawerOpen = false;
		}}
	>
		<aside
			aria-label="Danh sách bài vị"
			class="ml-auto flex h-full w-full max-w-xl flex-col bg-[var(--color-bg)] shadow-2xl"
		>
			<header
				class="border-b border-[var(--color-border)] bg-[var(--color-surface)] px-4 py-4 sm:px-5"
			>
				<div class="flex items-start justify-between gap-4">
					<div>
						<p class="text-xs font-medium text-[var(--color-text-secondary)]">
							{selectedPosition?.house_name} · Khu {selectedPosition?.area_code}
						</p>
						<h2 class="mt-1 text-lg font-semibold">Bài vị tại {selectedPosition?.name}</h2>
						<p class="mt-1 text-xs text-[var(--color-text-secondary)]">
							Cột {selectedPosition?.column_number}, hàng {selectedPosition?.row_number}
						</p>
					</div>
					<button
						type="button"
						onclick={() => (drawerOpen = false)}
						aria-label="Đóng danh sách bài vị"
						class="grid h-10 w-10 shrink-0 place-items-center rounded-md hover:bg-[var(--color-surface-muted)]"
						><span class="icon-[lucide--x] h-5 w-5"></span></button
					>
				</div>
				{#if canWrite}<div class="mt-4 flex flex-wrap justify-end gap-2">
						<button
							type="button"
							onclick={() => selectedPosition && editPosition(selectedPosition)}
							class="h-11 rounded-md border border-[var(--color-border-strong)] px-3 text-sm font-semibold text-[var(--color-text-secondary)] hover:bg-[var(--color-surface-muted)]"
							>Sửa vị trí</button
						>
						<button
							type="button"
							onclick={() => {
								editingPosition = selectedPosition ?? null;
								deletePositionOpen = Boolean(editingPosition);
							}}
							class="h-11 rounded-md px-3 text-sm font-semibold text-[var(--color-danger)] hover:bg-[var(--color-danger-soft)]"
							>Xoá vị trí</button
						>
						<button
							type="button"
							onclick={() => openUnplacedTablets(selectedPosition ?? null)}
							class="h-11 rounded-md border border-[var(--color-primary)] px-4 text-sm font-semibold text-[var(--color-primary-dark)]"
							>Thêm bài vị có sẵn</button
						>
						<button
							type="button"
							onclick={() => void open('tablet')}
							class="h-11 shrink-0 rounded-md bg-[var(--color-primary)] px-4 text-sm font-semibold text-white"
							><span class="mr-1 icon-[lucide--plus] inline-block h-4 w-4 align-text-bottom"
							></span>Thêm</button
						>
					</div>{/if}
			</header>
			<div class="flex-1 overflow-y-auto p-4 sm:p-5">
				{#if !drawerLoading}<p class="mb-3 text-xs text-[var(--color-text-secondary)]">
						{tablets.length} bài vị
					</p>{/if}
				{#if drawerLoading}<div class="py-16">
						<LoadingIndicator label="Đang tải danh sách bài vị..." />
					</div>{:else if tablets.length === 0}<div
						class="rounded-md border border-dashed border-[var(--color-border-strong)] py-12 text-center text-sm text-[var(--color-text-secondary)]"
					>
						Chưa có bài vị tại vị trí này
					</div>{:else}<div class="space-y-3">
					{#each tablets as tablet (tablet.id)}<article class="rounded-md border border-[#d8c39c] bg-[#fffdf8] p-4 shadow-sm">
							<header class="flex items-start justify-between gap-3 border-b border-[#eadfc9] pb-3"><div><p class="text-xs font-semibold text-[#80673c]">{selectedPosition?.name}</p><h3 class="mt-1 font-semibold text-[#382c22]">{tablet.name}</h3></div><div class="space-y-1 text-right"><span class="block rounded-sm bg-[#f4e5be] px-1.5 py-0.5 text-[10px] font-semibold text-[#604820]">{tabletTypeLabel(tablet.type)}</span><span class="block text-xs font-semibold text-[#406b5b]">{tabletStatusLabel(tablet.status)}</span></div></header>
							<section class="pt-3"><p class="mb-2 text-xs font-semibold text-[#80673c]">Danh sách Hương linh</p><div class="space-y-3">{#each tabletSpiritPreviews.get(tablet.id) ?? [] as spirit (spirit.id)}<div class="flex gap-3"><SpiritPortrait imageUrl={spirit.image_url} alt={spirit.full_name} sizeClass="h-16 w-12 shrink-0" /><div class="min-w-0 flex-1"><p class="font-semibold text-[#382c22]">{spirit.full_name}</p>{#if spirit.dharma_name}<p class="text-sm text-[var(--color-text-secondary)]">Pháp danh: {spirit.dharma_name}</p>{/if}{#if spirit.birth_year || spirit.death_year}<p class="text-sm text-[var(--color-text-secondary)]">{spirit.birth_year || '?'} – {spirit.death_year || '?'}</p>{/if}<details class="mt-1 text-xs text-[var(--color-text-secondary)]"><summary class="cursor-pointer font-semibold text-[var(--color-primary-dark)]">Thông tin thêm</summary><div class="mt-1 space-y-0.5"><p>{spirit.death_date ? `Ngày mất/kỵ: ${spirit.death_date}` : ''}</p><p>{spirit.enshrined_at ? `Ngày an vị: ${spirit.enshrined_at}` : ''}</p><p>{spirit.notes ? `Ghi chú: ${spirit.notes}` : ''}</p></div></details></div></div>{/each}</div></section>
							{#if canWrite}<footer class="mt-4 flex flex-wrap justify-end gap-2 border-t border-[#eadfc9] pt-3"><button type="button" onclick={() => void editTablet(tablet)} class="h-9 rounded-md border border-[var(--color-border-strong)] px-3 text-xs font-semibold">Sửa Bài vị</button><button type="button" onclick={() => openTabletMove(tablet)} class="h-9 rounded-md border border-[var(--color-primary)] px-3 text-xs font-semibold text-[var(--color-primary-dark)]">Chuyển vị trí</button><button type="button" onclick={() => openTabletDelete(tablet)} class="h-9 rounded-md border border-[var(--color-danger)] px-3 text-xs font-semibold text-[var(--color-danger)]">Xoá</button></footer>{/if}
						</article>{/each}
					</div>{/if}
			</div>
		</aside>
	</div>{/if}

{#if mode}<div
		class="fixed inset-0 z-50 grid place-items-end bg-black/40 md:place-items-center"
		role="presentation"
	onclick={(e) => {
			if (e.target === e.currentTarget) void requestCloseMode();
		}}
	>
		<form
			onsubmit={save}
			class={[
				'max-h-[94dvh] w-full rounded-t-xl bg-[var(--color-surface)] p-5 md:rounded-xl',
				mode === 'tablet' || (mode === 'position' && !editingPosition)
					? 'flex h-[94dvh] flex-col overflow-hidden'
					: 'overflow-y-auto',
				mode === 'tablet'
					? 'md:max-w-6xl'
					: mode === 'position' && !editingPosition
						? 'md:max-w-4xl'
						: 'md:max-w-lg'
			]}
		>
			<h2 class="mb-4 shrink-0 text-lg font-semibold">
				{mode === 'house'
					? editingHouse ? 'Sửa Nhà Linh' : 'Thêm Nhà Linh'
					: mode === 'area'
						? editingArea ? 'Sửa khu vực' : 'Thêm khu vực'
						: mode === 'position'
							? editingPosition
								? 'Sửa vị trí'
								: 'Thêm nhiều vị trí'
							: mode === 'tablet'
								? editingTablet
									? canWrite
										? 'Sửa bài vị và danh sách Hương linh'
										: 'Thông tin bài vị và danh sách Hương linh'
									: 'Thêm bài vị'
								: ''}
			</h2>
			<div
				class={mode === 'tablet' || (mode === 'position' && !editingPosition)
					? 'flex min-h-0 flex-1 flex-col'
					: 'space-y-4'}
			>
				{#if mode === 'house'}{@render input(
						'Tên Nhà Linh *',
						houseForm,
						'name',
						true
					)}{@render input('Địa chỉ', houseForm, 'address')}{@render textarea('Ghi chú', houseForm)}
				{:else if mode === 'area'}{@render input(
						'Mã khu vực *',
						areaForm,
						'code',
						true
					)}{@render input('Tên mô tả', areaForm, 'name')}{@render textarea('Ghi chú', areaForm)}
				{:else if mode === 'position'}{#if editingPosition}<div
							class="rounded-md bg-[var(--color-primary-soft)] px-3 py-2 text-sm text-[var(--color-primary-dark)]"
						>
							Tên vị trí tự động:
							<strong
											>{positionForm.column_number}{areas.find((a) => a.id === areaId)?.code ??
													''}-{positionForm.row_number}</strong
							>
						</div>
						<div class="grid grid-cols-2 gap-3">
							<label
								><span class="mb-1 block text-sm">Cột *</span><input
									bind:value={positionForm.column_number}
									type="number"
									min="1"
									required
									class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
								/></label
							><label
								><span class="mb-1 block text-sm">Hàng *</span><input
									bind:value={positionForm.row_number}
									type="number"
									min="1"
									required
									class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
								/></label
							>
						</div>
						{@render textarea('Ghi chú', positionForm)}
					{:else}<InlinePositionEditor
							bind:items={newPositions}
							areaCode={areas.find((area) => area.id === areaId)?.code ?? ''}
						/>{/if}
				{:else if mode === 'tablet'}<div
						class="shrink-0 space-y-3 border-b border-[var(--color-border)] pb-3"
					>
						<div>
							{#if editingTablet}<p class="mb-3 rounded-md bg-[var(--color-primary-soft)] px-3 py-2 text-sm text-[var(--color-primary-dark)]">Vị trí: {editingTablet.position_name || 'Chưa xếp vị trí'}</p>{/if}
							<label class="block"
				><span class="mb-1 block text-sm">Mã bài vị *</span><input
					bind:this={tabletNameInput}
					bind:value={tabletForm.code}
									required={canWrite}
									readonly={!canWrite}
									class="h-10 w-full rounded-md border-[var(--color-border-strong)]"
								/></label
							>
						</div>
						<div class="grid gap-3 md:grid-cols-2"><label class="block"><span class="mb-1 block text-sm">Ảnh Bài vị (URL)</span><input bind:value={tabletForm.image_url} readonly={!canWrite} placeholder="Tự lấy ảnh Hương linh khi chỉ có một người" class="h-10 w-full rounded-md border-[var(--color-border-strong)]" />{#if !tabletForm.image_url && tabletForm.spirits.length === 1 && tabletForm.spirits[0].image_url}<span class="mt-1 block text-xs text-[var(--color-text-secondary)]">Đang dùng mặc định ảnh của Hương linh.</span>{:else if !tabletForm.image_url && tabletForm.spirits.length > 1}<span class="mt-1 block text-xs text-[var(--color-text-secondary)]">Nhiều Hương linh: hãy chọn hoặc tải ảnh Bài vị; khi chưa chọn sẽ dùng ảnh Hương linh có sẵn để hiển thị.</span>{/if}</label><label class="block"><span class="mb-1 block text-sm">Người gửi</span><input bind:value={tabletForm.sender} readonly={!canWrite} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label></div>
						<div class="grid gap-3 md:grid-cols-2"><label class="block"><span class="mb-1 block text-sm">Loại Bài vị</span><select bind:value={tabletForm.type} disabled={!canWrite} class="h-10 w-full rounded-md border-[var(--color-border-strong)]"><option value="spirit">Hương linh</option><option value="giac_linh">Giác linh</option><option value="family">Gia tiên</option><option value="cuu_huyen">Cửu Huyền Thất Tổ</option><option value="clan">Tộc họ</option><option value="collective">Chư Hương linh</option><option value="fetus">Thai nhi</option><option value="martyr">Anh hùng liệt sĩ</option><option value="victim">Hương linh tử nạn</option><option value="childless">Hương linh vô tự</option><option value="unknown">Hương linh vô danh</option><option value="other">Khác</option></select></label><label class="block"><span class="mb-1 block text-sm">Trạng thái *</span><select bind:value={tabletForm.status} disabled={!canWrite} class="h-10 w-full rounded-md border-[var(--color-border-strong)]"><option value="pending">Chờ an vị</option><option value="enshrined">Đã an vị</option><option value="taken_home">Đã thỉnh về</option></select></label></div><div class="grid gap-3 md:grid-cols-3"><label><span class="mb-1 block text-sm">Ngày đăng ký</span><input type="date" bind:value={tabletForm.registered_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label><label><span class="mb-1 block text-sm">Ngày an vị</span><input type="date" bind:value={tabletForm.enshrined_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label><label><span class="mb-1 block text-sm">Ngày đưa vào khu thờ</span><input type="date" bind:value={tabletForm.entered_worship_area_at} class="h-10 w-full rounded-md border-[var(--color-border-strong)]" /></label></div>
						{@render textarea('Ghi chú', tabletForm)}
						{#if !editingTablet}<UnplacedSpiritPicker
								houseId={selectedPosition?.house_id ?? houseId}
								bind:selected={selectedUnplacedSpirits}
							/>{/if}
					</div>
					<InlineSpiritEditor
						bind:items={tabletForm.spirits}
						onbusychange={(busy) => (inlineEditorBusy = busy)}
						allowExistingUploads={Boolean(editingTablet)}
						requireFirst={Boolean(editingTablet) || selectedUnplacedSpirits.length === 0}
						readOnly={Boolean(editingTablet) && !canWrite}
					/>
				{/if}
			</div>
			<div
				class={[
					'flex shrink-0 justify-end gap-3',
					mode === 'tablet' || (mode === 'position' && !editingPosition)
						? 'border-t border-[var(--color-border)] pt-3'
						: 'mt-5'
				]}
			>
				{#if mode === 'position' && editingPosition && canWrite}<button
					type="button"
					onclick={() => (deletePositionOpen = true)}
					class="mr-auto h-11 rounded-md px-3 text-sm font-semibold text-[var(--color-danger)] hover:bg-[var(--color-danger-soft)]"
					>Xoá vị trí</button
				>{/if}<button
					type="button"
					onclick={() => void requestCloseMode()}
					class="h-11 rounded-md border px-5 text-sm font-semibold"
					>{mode === 'tablet' && editingTablet && !canWrite ? 'Đóng' : 'Huỷ'}</button
				>{#if !(mode === 'tablet' && editingTablet && !canWrite)}<button
						type="submit"
						disabled={saving || inlineEditorBusy}
						class="h-11 rounded-md bg-[var(--color-primary)] px-6 text-sm font-semibold text-white disabled:opacity-50"
						>Lưu</button
					>{/if}
			</div>
		</form>
</div>{/if}

<Popup open={unplacedTabletsOpen} title={assignmentTarget ? `Thêm Bài vị vào ${assignmentTarget.name}` : 'Bài vị chưa xếp'} onClose={() => !saving && (unplacedTabletsOpen = false)}>
	<div class="space-y-3">
		<p class="text-sm text-[var(--color-text-secondary)]">{assignmentTarget ? 'Chọn Bài vị để gán vào Vị trí này.' : 'Các Bài vị này đang được giữ lại nhưng chưa thuộc Vị trí nào.'}</p>
		<label class="relative block"><span class="absolute top-3 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"></span><input bind:value={unplacedTabletQuery} placeholder="Tìm tên Bài vị..." class="h-10 w-full rounded-md border-[var(--color-border-strong)] pl-9 text-sm" /></label>
		<div class="max-h-[50dvh] space-y-2 overflow-y-auto pr-1">
			{#each filteredUnplacedTablets as tablet (tablet.id)}<div class="flex items-center gap-3 rounded-md border border-[var(--color-border)] p-3"><div class="min-w-0 flex-1"><p class="truncate font-semibold">{tablet.name}</p><p class="mt-0.5 text-xs text-[var(--color-text-secondary)]">{tablet.spirit_count} Hương linh</p></div>{#if assignmentTarget}<button type="button" disabled={saving} onclick={() => void assignUnplacedTablet(tablet)} class="h-9 shrink-0 rounded-md bg-[var(--color-primary)] px-3 text-xs font-semibold text-white disabled:opacity-50">Gán vào vị trí này</button>{/if}</div>{:else}<p class="py-8 text-center text-sm text-[var(--color-text-secondary)]">Không có Bài vị chưa xếp phù hợp.</p>{/each}
		</div>
	</div>
	{#snippet footer()}<button type="button" onclick={() => (unplacedTabletsOpen = false)} class="ml-auto h-10 rounded-md border border-[var(--color-border-strong)] px-5 font-semibold">Đóng</button>{/snippet}
</Popup>

<Popup open={moveTabletOpen} title="Chuyển Bài vị sang Vị trí khác" onClose={() => !saving && (moveTabletOpen = false)}>
	<div class="space-y-3">
		<p class="text-sm text-[var(--color-text-secondary)]">Bài vị <strong class="text-[var(--color-text)]">{movingTablet?.name}</strong> sẽ được chuyển, danh sách Hương linh vẫn giữ nguyên.</p>
		<label class="relative block"><span class="absolute top-3 left-3 icon-[lucide--search] h-4 w-4 text-[var(--color-text-muted)]"></span><input bind:value={movePositionQuery} oninput={() => void findMovePositions()} placeholder="Tìm vị trí, ví dụ 38c1..." class="h-10 w-full rounded-md border-[var(--color-border-strong)] pl-9 text-sm" /></label>
		<div class="max-h-[50dvh] space-y-2 overflow-y-auto pr-1">
			{#if movePositionResults.length}{#each movePositionResults as position (position.id)}<button type="button" disabled={saving} onclick={() => void confirmTabletMove(position)} class="flex w-full items-center justify-between rounded-md border border-[var(--color-border)] p-3 text-left hover:bg-[var(--color-primary-soft)] disabled:opacity-50"><span><strong>{position.name}</strong><span class="ml-2 text-xs text-[var(--color-text-secondary)]">Khu {position.area_code} · Cột {position.column_number}, hàng {position.row_number}</span></span><span class="text-xs font-semibold text-[var(--color-primary-dark)]">Chuyển</span></button>{/each}{:else if movePositionQuery.trim()}<p class="py-8 text-center text-sm text-[var(--color-text-secondary)]">Không tìm thấy vị trí phù hợp.</p>{/if}
		</div>
	</div>
	{#snippet footer()}<button type="button" onclick={() => (moveTabletOpen = false)} class="ml-auto h-10 rounded-md border border-[var(--color-border-strong)] px-5 font-semibold">Huỷ</button>{/snippet}
</Popup>

<Popup open={deletePositionOpen} title="Xoá Vị trí" onClose={() => !saving && (deletePositionOpen = false)}>
	<div class="space-y-3 text-sm"><p>Vị trí <strong>{editingPosition?.name}</strong> sẽ bị xoá.</p><div class="rounded-md bg-[var(--color-primary-soft)] p-3 text-[var(--color-text-secondary)]">Bài vị và Hương linh trong Vị trí sẽ không bị xoá; Bài vị được đưa về danh sách <strong>Bài vị chưa xếp</strong>.</div></div>
	{#snippet footer()}<div class="grid grid-cols-2 gap-3"><button type="button" onclick={() => (deletePositionOpen = false)} class="h-10 rounded-md border border-[var(--color-border-strong)] font-semibold">Huỷ</button><button type="button" disabled={saving} onclick={() => void confirmPositionDelete()} class="h-10 rounded-md bg-[var(--color-danger)] font-semibold text-white disabled:opacity-50">{saving ? 'Đang xoá...' : 'Xác nhận xoá'}</button></div>{/snippet}
</Popup>

<Popup open={deleteTabletOpen} title="Xoá Bài vị" onClose={() => !saving && (deleteTabletOpen = false)}>
	<div class="space-y-3 text-sm">
		<p>Bài vị <strong>{tabletToDelete?.name}</strong> sẽ bị xoá.</p>
		<label class="flex cursor-pointer items-start gap-3 rounded-md bg-[var(--color-danger-soft)] p-3"><input type="checkbox" bind:checked={deleteTabletSpirits} class="mt-0.5 h-4 w-4" /><span><strong>Xoá các Hương linh trong Bài vị</strong><br /><span class="text-[var(--color-text-secondary)]">Mặc định tắt: Hương linh sẽ được giữ lại ở danh sách chưa xếp.</span></span></label>
	</div>
	{#snippet footer()}<div class="grid grid-cols-2 gap-3"><button type="button" onclick={() => (deleteTabletOpen = false)} class="h-10 rounded-md border border-[var(--color-border-strong)] font-semibold">Huỷ</button><button type="button" disabled={saving} onclick={() => void confirmTabletDelete()} class="h-10 rounded-md bg-[var(--color-danger)] font-semibold text-white disabled:opacity-50">{saving ? 'Đang xoá...' : 'Xác nhận xoá'}</button></div>{/snippet}
</Popup>

{#snippet sortHeader(label: string, key: PositionSortKey, right = false)}<button
		type="button"
		onclick={() => sortPositions(key)}
		class={['flex w-full items-center gap-1 font-semibold', right && 'justify-end']}
		>{label}<span
			class={[
				'h-3.5 w-3.5',
				positionSortKey !== key
					? 'icon-[lucide--arrow-up-down] opacity-40'
					: positionSortDirection === 'asc'
						? 'icon-[lucide--arrow-up] text-[var(--color-primary-dark)]'
						: 'icon-[lucide--arrow-down] text-[var(--color-primary-dark)]'
			]}
		></span></button
	>{/snippet}

{#snippet input(
	label: string,
	object: Record<string, string | number>,
	key: string,
	required = false
)}<label class="block"
		><span class="mb-1 block text-sm">{label}</span><input
			value={object[key]}
			oninput={(e) => (object[key] = e.currentTarget.value)}
			{required}
			class="h-11 w-full rounded-md border-[var(--color-border-strong)]"
		/></label
	>{/snippet}
{#snippet textarea(label: string, object: { notes: string })}<label class="block"
		><span class="mb-1 block text-sm">{label}</span><textarea
			bind:value={object.notes}
			rows="3"
			class="w-full rounded-md border-[var(--color-border-strong)]"></textarea></label
	>{/snippet}
