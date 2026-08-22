import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { parsePositionSheet } from './position-sheet-parser.ts';

describe('parsePositionSheet', () => {
	it('parses rows copied from Excel or Google Sheets with an optional header', () => {
		const rows = parsePositionSheet('Cột\tHàng\tGhi chú\n2\t1\tTầng trên\n1\t2\t');
		assert.deepEqual(rows, [
			{ row_number: '1', column_number: '2', notes: 'Tầng trên' },
			{ row_number: '2', column_number: '1', notes: '' }
		]);
	});

	it('keeps duplicate coordinates so the batch API can report skipped rows', () => {
		assert.deepEqual(parsePositionSheet('2\t1\n2\t1'), [
			{ row_number: '1', column_number: '2', notes: '' },
			{ row_number: '1', column_number: '2', notes: '' }
		]);
	});
});
