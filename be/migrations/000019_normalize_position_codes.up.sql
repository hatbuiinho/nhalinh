UPDATE memorial_positions AS p
SET name = p.column_number::text || UPPER(a.code) || '-' || p.row_number::text,
    updated_at = NOW()
FROM memorial_areas AS a
WHERE a.id = p.area_id
  AND p.name IS DISTINCT FROM p.column_number::text || UPPER(a.code) || '-' || p.row_number::text;
