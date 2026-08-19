import assert from 'node:assert/strict';
import { describe, it } from 'node:test';
import { parseSpiritSheet } from './sheet-parser.ts';

describe('parseSpiritSheet', () => {
	it('parses multiple Google Sheets rows with an optional header', () => {
		const items = parseSpiritSheet(
			'Tên\tPháp danh\tNăm sinh\tNăm mất\tTuổi\tHình URL\tNơi an táng\tNgười gửi\tTháng gửi\tGhi chú\nNguyễn Văn A\tThiện Tâm\t1940\t2020\t80\t\tHuế\tGia đình\t8/2026\t\nTrần Thị B\t\t1950'
		);
		assert.equal(items.length, 2);
		assert.equal(items[0].dharma_name, 'Thiện Tâm');
		assert.equal(items[1].birth_year, '1950');
	});
	it('requires a name on every row', () =>
		assert.throws(() => parseSpiritSheet('\tThiện Tâm'), /Tên Hương linh/));
});
