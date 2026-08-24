ALTER TABLE spirits ADD COLUMN has_urn boolean NOT NULL DEFAULT false;
CREATE INDEX idx_spirits_active_urn ON spirits(house_id, has_urn) WHERE deleted_at IS NULL;
