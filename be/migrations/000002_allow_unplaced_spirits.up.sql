ALTER TABLE spirits ADD COLUMN house_id text REFERENCES spirit_houses(id) ON DELETE CASCADE;

UPDATE spirits s
SET house_id = a.house_id
FROM memorial_tablets t
JOIN memorial_positions p ON p.id = t.position_id
JOIN memorial_areas a ON a.id = p.area_id
WHERE s.tablet_id = t.id;

ALTER TABLE spirits ALTER COLUMN house_id SET NOT NULL;
ALTER TABLE spirits ALTER COLUMN tablet_id DROP NOT NULL;

CREATE INDEX idx_spirits_house ON spirits(house_id, full_name);
