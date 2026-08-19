CREATE INDEX idx_spirits_unplaced_house_updated
ON spirits(house_id, updated_at DESC)
WHERE tablet_id IS NULL;
