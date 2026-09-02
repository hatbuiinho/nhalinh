-- This irreversible data migration is intentionally ordered before tablet consolidation.
-- It permanently removes soft-deleted spirits as requested.
LOCK TABLE spirits, memorial_tablets IN SHARE ROW EXCLUSIVE MODE;

DELETE FROM spirits
WHERE deleted_at IS NOT NULL;

CREATE TEMP TABLE tablet_merge_candidates ON COMMIT DROP AS
WITH ranked_tablets AS (
  SELECT
    id,
    FIRST_VALUE(id) OVER (
      PARTITION BY position_id
      ORDER BY created_at ASC, id ASC
    ) AS keeper_tablet_id,
    ROW_NUMBER() OVER (
      PARTITION BY position_id
      ORDER BY created_at ASC, id ASC
    ) AS position_order
  FROM memorial_tablets
)
SELECT id AS duplicate_tablet_id, keeper_tablet_id
FROM ranked_tablets
WHERE position_order > 1;

UPDATE spirits AS s
SET tablet_id = candidates.keeper_tablet_id,
    updated_at = NOW()
FROM tablet_merge_candidates AS candidates
WHERE s.tablet_id = candidates.duplicate_tablet_id;

DELETE FROM memorial_tablets AS tablet
USING tablet_merge_candidates AS candidates
WHERE tablet.id = candidates.duplicate_tablet_id;
