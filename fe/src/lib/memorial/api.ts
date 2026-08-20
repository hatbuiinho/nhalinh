import { apiFetch, apiRequest } from '$lib/api/client';

export type House = {
	id: string;
	name: string;
	address: string;
	notes: string;
	active: boolean;
	access_role: 'admin' | 'editor' | 'viewer';
	created_at: string;
	updated_at: string;
};
export type Area = {
	id: string;
	house_id: string;
	code: string;
	name: string;
	notes: string;
	position_count: number;
	tablet_count: number;
	spirit_count: number;
};
export type Position = {
	id: string;
	area_id: string;
	house_id: string;
	house_name: string;
	area_code: string;
	row_number: number;
	column_number: number;
	name: string;
	notes: string;
	tablet_count: number;
	spirit_count: number;
};
export type OccupancySummary = {
	area_count: number;
	position_count: number;
	empty_position_count: number;
	used_position_count: number;
	tablet_count: number;
	spirit_count: number;
	unplaced_spirit_count: number;
};
export type OccupancyArea = {
	id: string;
	code: string;
	name: string;
	position_count: number;
	empty_position_count: number;
	tablet_count: number;
	spirit_count: number;
};
export type Occupancy = {
	house_id: string;
	summary: OccupancySummary;
	areas: OccupancyArea[];
	positions: Position[];
};
export type Tablet = {
	id: string;
	position_id: string;
	house_id: string;
	house_name: string;
	area_id: string;
	area_code: string;
	position_name: string;
	row_number: number;
	column_number: number;
	name: string;
	notes: string;
	spirit_count: number;
};
export type Spirit = {
	id: string;
	tablet_id: string;
	house_id: string;
	house_name: string;
	area_id: string;
	area_code: string;
	position_id: string;
	position_name: string;
	tablet_name: string;
	full_name: string;
	dharma_name: string;
	birth_year: string;
	death_year: string;
	age: string;
	image_url: string;
	burial_place: string;
	sender: string;
	sent_month: string;
	notes: string;
	created_at: string;
	updated_at: string;
};
export type SpiritInput = Pick<
	Spirit,
	| 'house_id'
	| 'tablet_id'
	| 'full_name'
	| 'dharma_name'
	| 'birth_year'
	| 'death_year'
	| 'age'
	| 'image_url'
	| 'burial_place'
	| 'sender'
	| 'sent_month'
	| 'notes'
>;
export type InlineSpiritInput = Omit<SpiritInput, 'house_id' | 'tablet_id'>;
export type EditableSpiritInput = InlineSpiritInput & { id?: string };
export type SpiritImportIssue = {
	row_number: number;
	message: string;
};
export type SpiritImportPreview = {
	total_rows: number;
	valid_rows: number;
	invalid_rows: number;
	create_area_count: number;
	create_position_count: number;
	create_tablet_count: number;
	create_spirit_count: number;
	errors: SpiritImportIssue[];
};
export type SpiritImportResult = {
	created_area_count: number;
	created_position_count: number;
	created_tablet_count: number;
	created_spirit_count: number;
};

export const listHouses = () => apiRequest<House[]>('/api/spirit-houses');
export const createHouse = (input: { name: string; address: string; notes: string }) =>
	apiRequest<House>('/api/spirit-houses', { method: 'POST', body: JSON.stringify(input) });
export const updateHouse = (
	id: string,
	input: { name: string; address: string; notes: string; active: boolean }
) =>
	apiRequest<House>(`/api/spirit-houses/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
export const listAreas = async (houseId: string) =>
	(
		await apiRequest<{ areas: Area[] }>(
			`/api/memorial-areas?house_id=${encodeURIComponent(houseId)}`
		)
	).areas;
export const createArea = (input: {
	house_id: string;
	code: string;
	name: string;
	notes: string;
}) => apiRequest<Area>('/api/memorial-areas', { method: 'POST', body: JSON.stringify(input) });
export const listPositions = async (areaId: string) =>
	(
		await apiRequest<{ positions: Position[] }>(
			`/api/memorial-positions?area_id=${encodeURIComponent(areaId)}`
		)
	).positions;
export const searchPositions = async (houseId: string, query: string, limit = 30) => {
	const params = new URLSearchParams({ house_id: houseId, q: query, limit: String(limit) });
	return (
		await apiRequest<{ positions: Position[] }>(`/api/memorial-positions?${params.toString()}`)
	).positions;
};
export const createPosition = (input: {
	area_id: string;
	row_number: number;
	column_number: number;
	notes: string;
}) =>
	apiRequest<Position>('/api/memorial-positions', { method: 'POST', body: JSON.stringify(input) });
export const createPositions = (
	areaId: string,
	positions: Array<{
		row_number: number;
		column_number: number;
		notes: string;
	}>
) =>
	apiRequest<{ positions: Position[]; skipped_count: number }>('/api/memorial-positions/batch', {
		method: 'POST',
		body: JSON.stringify({ area_id: areaId, positions })
	});
export const updatePosition = (
	id: string,
	input: {
		area_id: string;
		row_number: number;
		column_number: number;
		notes: string;
	}
) =>
	apiRequest<Position>(`/api/memorial-positions/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
export const getOccupancy = (houseId: string) =>
	apiRequest<Occupancy>(`/api/memorial-occupancy?house_id=${encodeURIComponent(houseId)}`);
export const listTablets = async (positionId: string) =>
	(
		await apiRequest<{ tablets: Tablet[] }>(
			`/api/memorial-tablets?position_id=${encodeURIComponent(positionId)}`
		)
	).tablets;
export const createTablet = (input: {
	position_id: string;
	name: string;
	notes: string;
	spirits: InlineSpiritInput[];
	existing_spirit_ids?: string[];
}) => apiRequest<Tablet>('/api/memorial-tablets', { method: 'POST', body: JSON.stringify(input) });
export const updateTablet = (
	id: string,
	input: {
		position_id: string;
		name: string;
		notes: string;
		spirits: EditableSpiritInput[];
	}
) =>
	apiRequest<Tablet>(`/api/memorial-tablets/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
export const listTabletSpirits = async (tabletId: string) =>
	(
		await apiRequest<{ spirits: Spirit[] }>(
			`/api/spirits?tablet_id=${encodeURIComponent(tabletId)}&limit=500&offset=0`
		)
	).spirits;
export async function listSpirits(
	query: string,
	houseId: string,
	areaId: string,
	limit = 100,
	offset = 0
) {
	const p = new URLSearchParams({
		q: query,
		house_id: houseId,
		area_id: areaId,
		limit: String(limit),
		offset: String(offset)
	});
	return apiRequest<{ spirits: Spirit[]; total: number; has_more: boolean }>(`/api/spirits?${p}`);
}
export async function searchUnplacedSpirits(houseId: string, query: string, limit = 20) {
	const params = new URLSearchParams({
		house_id: houseId,
		q: query,
		unplaced: 'true',
		limit: String(limit),
		offset: '0'
	});
	return (
		await apiRequest<{ spirits: Spirit[]; total: number; has_more: boolean }>(
			`/api/spirits?${params}`
		)
	).spirits;
}
export const createSpirit = (input: SpiritInput) =>
	apiRequest<Spirit>('/api/spirits', { method: 'POST', body: JSON.stringify(input) });
export const createSpirits = async (spirits: SpiritInput[]) =>
	(
		await apiRequest<{ spirits: Spirit[] }>('/api/spirits/batch', {
			method: 'POST',
			body: JSON.stringify({ spirits })
		})
	).spirits;
export const updateSpirit = (id: string, input: SpiritInput) =>
	apiRequest<Spirit>(`/api/spirits/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
export const patchSpirit = (id: string, field: string, value: string) =>
	apiRequest<Spirit>(`/api/spirits/${encodeURIComponent(id)}`, {
		method: 'PATCH',
		body: JSON.stringify({ field, value })
	});
export const deleteSpirit = (id: string) =>
	apiRequest<void>(`/api/spirits/${encodeURIComponent(id)}`, { method: 'DELETE' });

export async function downloadSpiritImportTemplate() {
	return apiBlob('/api/spirits/import-template.xlsx');
}

export async function previewSpiritImport(file: File, houseId: string) {
	const body = new FormData();
	body.set('house_id', houseId);
	body.set('file', file);
	return apiRequest<SpiritImportPreview>('/api/spirits/import-preview', {
		method: 'POST',
		body
	});
}

export async function importSpiritsFromExcel(file: File, houseId: string) {
	const body = new FormData();
	body.set('house_id', houseId);
	body.set('file', file);
	return apiRequest<SpiritImportResult>('/api/spirits/import', {
		method: 'POST',
		body
	});
}

export async function exportSpiritsExcel(
	scope: 'current' | 'all',
	filters: { query: string; houseId: string; areaId: string }
) {
	const params = new URLSearchParams({ scope });
	if (scope === 'current') {
		if (filters.query) params.set('q', filters.query);
		if (filters.houseId) params.set('house_id', filters.houseId);
		if (filters.areaId) params.set('area_id', filters.areaId);
	}
	return apiBlob(`/api/spirits/export.xlsx?${params.toString()}`);
}

async function apiBlob(path: string) {
	const response = await apiFetch(path);
	if (!response.ok) {
		let message = `Yêu cầu thất bại (${response.status})`;
		try {
			const body = (await response.json()) as { error?: { message?: string } };
			message = body.error?.message ?? message;
		} catch {
			// Keep fallback when response is not JSON.
		}
		throw new Error(message);
	}
	return {
		blob: await response.blob(),
		filename: response.headers.get('Content-Disposition')
	};
}
