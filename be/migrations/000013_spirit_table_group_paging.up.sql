CREATE INDEX idx_spirits_active_house_tablet_name
ON spirits(house_id, tablet_id, full_name, id)
WHERE deleted_at IS NULL;
