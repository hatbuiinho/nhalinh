ALTER TABLE memorial_tablets ADD COLUMN house_id text;
UPDATE memorial_tablets t SET house_id=a.house_id FROM memorial_positions p JOIN memorial_areas a ON a.id=p.area_id WHERE p.id=t.position_id;
ALTER TABLE memorial_tablets ALTER COLUMN house_id SET NOT NULL, ADD CONSTRAINT memorial_tablets_house_id_fkey FOREIGN KEY (house_id) REFERENCES spirit_houses(id) ON DELETE CASCADE, ALTER COLUMN position_id DROP NOT NULL;
CREATE INDEX idx_memorial_tablets_house_unplaced ON memorial_tablets(house_id, position_id);
