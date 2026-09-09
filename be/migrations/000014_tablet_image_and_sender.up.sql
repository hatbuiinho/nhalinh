ALTER TABLE memorial_tablets
  ADD COLUMN image_url text NOT NULL DEFAULT '',
  ADD COLUMN sender text NOT NULL DEFAULT '';

UPDATE memorial_tablets t
SET sender = (
  SELECT s.sender
  FROM spirits s
  WHERE s.tablet_id = t.id AND s.deleted_at IS NULL AND s.sender <> ''
  ORDER BY s.created_at, s.id
  LIMIT 1
)
WHERE t.sender = ''
  AND EXISTS (
    SELECT 1 FROM spirits s
    WHERE s.tablet_id = t.id AND s.deleted_at IS NULL AND s.sender <> ''
  );
