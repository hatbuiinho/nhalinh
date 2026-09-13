CREATE SEQUENCE IF NOT EXISTS memorial_tablet_code_seq START WITH 1;

WITH numbered AS (
  SELECT id, row_number() OVER (ORDER BY created_at, id) AS number
  FROM memorial_tablets
)
UPDATE memorial_tablets AS tablet
SET code = 'BV' || lpad(numbered.number::text, 6, '0')
FROM numbered
WHERE tablet.id = numbered.id;

SELECT setval('memorial_tablet_code_seq', GREATEST((SELECT count(*) FROM memorial_tablets), 1), true);
CREATE UNIQUE INDEX IF NOT EXISTS memorial_tablets_code_unique ON memorial_tablets(code);
