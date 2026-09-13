ALTER TABLE memorial_tablets
  ADD COLUMN status text NOT NULL DEFAULT 'enshrined'
  CHECK (status IN ('pending', 'enshrined', 'moving', 'moved', 'archived'));

UPDATE memorial_tablets SET status = 'moved' WHERE position_id IS NULL;

CREATE INDEX idx_memorial_tablets_status ON memorial_tablets(status);
