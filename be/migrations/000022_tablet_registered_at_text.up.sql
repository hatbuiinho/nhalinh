ALTER TABLE memorial_tablets
  ALTER COLUMN registered_at TYPE text USING CASE
    WHEN registered_at IS NULL THEN ''
    ELSE to_char(registered_at, 'DD/MM/YYYY')
  END;

ALTER TABLE memorial_tablets
  ALTER COLUMN registered_at SET DEFAULT '',
  ALTER COLUMN registered_at SET NOT NULL;
