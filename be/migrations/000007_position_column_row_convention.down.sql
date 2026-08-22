-- Restore the former stored-coordinate convention for a migration rollback.
UPDATE memorial_positions
SET row_number = row_number + 1000000000,
    column_number = column_number + 1000000000;

UPDATE memorial_positions
SET row_number = column_number - 1000000000,
    column_number = row_number - 1000000000;

UPDATE memorial_positions p
SET name = p.row_number::text || upper(a.code) || '-' || p.column_number::text
FROM memorial_areas a
WHERE a.id = p.area_id;

DROP INDEX IF EXISTS idx_memorial_positions_area;
CREATE INDEX idx_memorial_positions_area ON memorial_positions(area_id, row_number, column_number);

COMMENT ON COLUMN memorial_positions.row_number IS 'Hàng của vị trí';
COMMENT ON COLUMN memorial_positions.column_number IS 'Cột của vị trí';
