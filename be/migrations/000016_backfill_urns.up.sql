INSERT INTO urns (
  id, spirit_id, code, urn_type, material, color, dimensions,
  storage_location, image_url, status, notes, created_at, updated_at
)
SELECT
  'urn-backfill-' || s.id,
  s.id,
  '', '', '', '', '', '', '', 'installed',
  'Tạo tự động từ trạng thái Có Hũ Cốt trước đây.',
  now(), now()
FROM spirits s
WHERE s.has_urn = true
  AND s.deleted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM urns u WHERE u.spirit_id = s.id);
