import { foldSheetText, parseTabularRows } from './tabular-parser.ts';

export type EditablePositionRow = {
	row_number: string;
	column_number: string;
	notes: string;
};

const columns: (keyof EditablePositionRow)[] = ['row_number', 'column_number', 'notes'];

export function emptyPositionRow(): EditablePositionRow {
	return { row_number: '', column_number: '', notes: '' };
}

export function parsePositionSheet(raw: string): EditablePositionRow[] {
	const rows = parseTabularRows(raw);
	if (isHeader(rows[0])) rows.shift();
	if (rows.length === 0) throw new Error('Không có dòng vị trí');
	return rows.map((cells, index) => {
		if (cells.length > columns.length)
			throw new Error(`Dòng ${index + 1} có nhiều hơn ${columns.length} cột`);
		const item = emptyPositionRow();
		columns.forEach((key, column) => (item[key] = cells[column] ?? ''));
		if (!isPositiveInteger(item.row_number))
			throw new Error(`Dòng ${index + 1}: Hàng phải là số nguyên lớn hơn 0`);
		if (!isPositiveInteger(item.column_number))
			throw new Error(`Dòng ${index + 1}: Cột phải là số nguyên lớn hơn 0`);
		return item;
	});
}

function isHeader(cells: string[]) {
	const first = foldSheetText(cells[0] ?? '');
	return first === 'hang' || first === 'row' || first === 'hang vi tri';
}

function isPositiveInteger(value: string) {
	return /^[1-9]\d*$/.test(value);
}
