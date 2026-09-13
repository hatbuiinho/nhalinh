ALTER TABLE memorial_tablets
  ADD COLUMN IF NOT EXISTS code text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS registered_at date,
  ADD COLUMN IF NOT EXISTS enshrined_at date,
  ADD COLUMN IF NOT EXISTS entered_worship_area_at date;

UPDATE memorial_tablets
SET code = name
WHERE code = '';

-- Bài vị cũ lấy ngày đăng ký từ trường "Ngày gửi" của Hương linh đầu tiên
-- có ngày hợp lệ. Dữ liệu theo tháng sẽ được quy về ngày đầu tháng.
UPDATE memorial_tablets t
SET registered_at = source.registered_at
FROM (
  SELECT DISTINCT ON (s.tablet_id)
    s.tablet_id,
    CASE
      -- Tháng 3/2025, 03/2025: giữ tháng, quy về ngày 01 trong dữ liệu.
      WHEN lower(trim(s.sent_month)) ~ '^tháng[[:space:]]*[0-9]{1,2}/[0-9]{4}$'
        THEN to_date(regexp_replace(lower(trim(s.sent_month)), '^tháng[[:space:]]*([0-9]{1,2})/([0-9]{4})$', '\2-\1-01'), 'YYYY-MM-DD')
      WHEN trim(s.sent_month) ~ '^[0-9]{1,2}/[0-9]{4}$'
        THEN to_date(regexp_replace(trim(s.sent_month), '^([0-9]{1,2})/([0-9]{4})$', '\2-\1-01'), 'YYYY-MM-DD')
      -- 2025/03/10: YYYY/MM/DD.
      WHEN trim(s.sent_month) ~ '^[0-9]{4}/[0-9]{2}/[0-9]{2}$'
        AND to_char(to_date(trim(s.sent_month), 'YYYY/MM/DD'), 'YYYY/MM/DD') = trim(s.sent_month)
        THEN to_date(trim(s.sent_month), 'YYYY/MM/DD')
      -- 2025-13-01 và 2025-02-3: YYYY-DD-M(M).
      WHEN trim(s.sent_month) ~ '^[0-9]{4}-[0-9]{1,2}-[0-9]{1,2}$'
        AND to_char(to_date(regexp_replace(trim(s.sent_month), '^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})$', '\1-\3-\2'), 'YYYY-MM-DD'), 'YYYY-M-D') = regexp_replace(trim(s.sent_month), '^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})$', '\1-\3-\2')
        THEN to_date(regexp_replace(trim(s.sent_month), '^([0-9]{4})-([0-9]{1,2})-([0-9]{1,2})$', '\1-\3-\2'), 'YYYY-MM-DD')
      ELSE NULL
    END AS registered_at
  FROM spirits s
  WHERE s.tablet_id IS NOT NULL AND s.deleted_at IS NULL AND trim(s.sent_month) <> ''
  ORDER BY s.tablet_id, s.created_at, s.id
) source
WHERE t.id = source.tablet_id
  AND t.registered_at IS NULL
  AND source.registered_at IS NOT NULL;

UPDATE memorial_tablets t
SET sender = source.sender
FROM (
  SELECT DISTINCT ON (s.tablet_id) s.tablet_id, trim(s.sender) AS sender
  FROM spirits s
  WHERE s.tablet_id IS NOT NULL AND s.deleted_at IS NULL AND trim(s.sender) <> ''
  ORDER BY s.tablet_id, s.created_at, s.id
) source
WHERE t.id = source.tablet_id
  AND t.sender = '';

UPDATE memorial_tablets
SET status = CASE WHEN status = 'pending' THEN 'pending' WHEN status = 'enshrined' THEN 'enshrined' ELSE 'taken_home' END;

ALTER TABLE memorial_tablets DROP CONSTRAINT IF EXISTS memorial_tablets_status_check;
ALTER TABLE memorial_tablets ADD CONSTRAINT memorial_tablets_status_check CHECK (status IN ('pending', 'enshrined', 'taken_home'));

ALTER TABLE memorial_tablets DROP CONSTRAINT IF EXISTS memorial_tablets_tablet_type_check;
UPDATE memorial_tablets
SET tablet_type = CASE tablet_type WHEN 'ancestral' THEN 'cuu_huyen' WHEN 'family' THEN 'family' ELSE 'spirit' END;
ALTER TABLE memorial_tablets ADD CONSTRAINT memorial_tablets_tablet_type_check CHECK (tablet_type IN ('spirit','giac_linh','family','cuu_huyen','clan','collective','fetus','martyr','victim','childless','unknown','other'));

CREATE INDEX IF NOT EXISTS idx_memorial_tablets_code ON memorial_tablets(code);
