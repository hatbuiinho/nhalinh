ALTER TABLE memorial_tablets
  ADD COLUMN tablet_type text NOT NULL DEFAULT 'spirit'
  CHECK (tablet_type IN ('ancestral', 'spirit', 'family'));
