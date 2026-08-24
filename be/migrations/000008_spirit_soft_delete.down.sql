DROP INDEX IF EXISTS idx_spirits_active_tablet;
DROP INDEX IF EXISTS idx_spirits_active_house;
ALTER TABLE spirits DROP COLUMN deleted_at;
