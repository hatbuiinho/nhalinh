export function parseTabularRows(raw: string): string[][] {
	const rows = raw
		.replace(/\r\n?/g, '\n')
		.split('\n')
		.filter((row) => row.trim())
		.map((row) => row.split('\t').map((cell) => cell.trim()));
	if (rows.length === 0) throw new Error('Không có dữ liệu để nhập');
	return rows;
}

export function foldSheetText(value: string) {
	return value
		.normalize('NFD')
		.replace(/[\u0300-\u036f]/g, '')
		.toLowerCase()
		.trim();
}
