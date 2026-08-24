ALTER TABLE spirits ADD COLUMN deleted_at timestamptz;
CREATE INDEX idx_spirits_active_house ON spirits(house_id, id) WHERE deleted_at IS NULL;
CREATE INDEX idx_spirits_active_tablet ON spirits(tablet_id, id) WHERE deleted_at IS NULL;
