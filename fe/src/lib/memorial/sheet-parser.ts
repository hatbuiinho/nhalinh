import type { EditableSpiritInput, InlineSpiritInput } from './api';
import { foldSheetText, parseTabularRows } from './tabular-parser.ts';

const columns: (keyof InlineSpiritInput)[] = [
	'full_name',
	'dharma_name',
	'birth_year',
	'death_year',
	'age',
	'image_url',
	'burial_place',
	'sender',
	'sent_month',
	'notes'
];

export function emptyInlineSpirit(): InlineSpiritInput {
	return {
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

export function toInlineSpirit(input: EditableSpiritInput): InlineSpiritInput {
	return {
		full_name: input.full_name,
		dharma_name: input.dharma_name,
		birth_year: input.birth_year,
		death_year: input.death_year,
		age: input.age,
		image_url: input.image_url,
		burial_place: input.burial_place,
		sender: input.sender,
		sent_month: input.sent_month,
		notes: input.notes
	};
}

export function parseSpiritSheet(raw: string): InlineSpiritInput[] {
	const rows = parseTabularRows(raw);
	if (isHeader(rows[0])) rows.shift();
	if (rows.length === 0) throw new Error('Không có dòng Hương linh');
	return rows.map((cells, index) => {
		if (cells.length > columns.length)
			throw new Error(`Dòng ${index + 1} có nhiều hơn ${columns.length} cột`);
		const item = emptyInlineSpirit();
		columns.forEach((key, column) => (item[key] = cells[column] ?? ''));
		if (!item.full_name) throw new Error(`Dòng ${index + 1}: Tên Hương linh không được để trống`);
		return item;
	});
}

function isHeader(cells: string[]): boolean {
	const first = foldSheetText(cells[0] ?? '');
	return first === 'ten' || first === 'ho ten' || first === 'ten huong linh';
}
