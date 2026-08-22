-- Position names now use the convention: Cột + Mã khu + '-' + Hàng.
-- Existing records used the inverse coordinate convention, so swap the stored
-- values in two statements to avoid transient unique-key collisions.
UPDATE memorial_positions
SET row_number = row_number + 1000000000,
    column_number = column_number + 1000000000;

UPDATE memorial_positions
SET row_number = column_number - 1000000000,
    column_number = row_number - 1000000000;

UPDATE memorial_positions p
SET name = p.column_number::text || upper(a.code) || '-' || p.row_number::text
FROM memorial_areas a
WHERE a.id = p.area_id;

DROP INDEX IF EXISTS idx_memorial_positions_area;
CREATE INDEX idx_memorial_positions_area ON memorial_positions(area_id, column_number, row_number);

COMMENT ON COLUMN memorial_positions.row_number IS 'Hàng của vị trí (phần sau dấu gạch ngang trong tên)';
COMMENT ON COLUMN memorial_positions.column_number IS 'Cột của vị trí (phần trước mã khu trong tên)';
