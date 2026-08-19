DELETE FROM spirits WHERE tablet_id IS NULL;
ALTER TABLE spirits ALTER COLUMN tablet_id SET NOT NULL;
DROP INDEX IF EXISTS idx_spirits_house;
ALTER TABLE spirits DROP COLUMN house_id;
