DROP TRIGGER IF EXISTS spirits_position_history ON spirits;
DROP FUNCTION IF EXISTS record_spirit_position_change();
DROP TABLE IF EXISTS spirit_position_history;
ALTER TABLE spirits
  DROP COLUMN IF EXISTS enshrined_at,
  DROP COLUMN IF EXISTS entered_worship_area_at,
  DROP COLUMN IF EXISTS status,
  DROP COLUMN IF EXISTS death_lunar,
  DROP COLUMN IF EXISTS birth_lunar,
  DROP COLUMN IF EXISTS death_date,
  DROP COLUMN IF EXISTS birth_date,
  DROP COLUMN IF EXISTS gender,
  DROP COLUMN IF EXISTS familiar_name;
