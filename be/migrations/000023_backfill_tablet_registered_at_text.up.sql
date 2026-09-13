-- Với Bài vị chưa có ngày đăng ký, giữ nguyên văn "Ngày gửi" của Hương linh
-- đầu tiên có dữ liệu. Trường đích là text nên hỗ trợ cả ngày/tháng không đầy đủ.
UPDATE memorial_tablets AS tablet
SET registered_at = source.sent_month
FROM (
  SELECT DISTINCT ON (spirit.tablet_id)
    spirit.tablet_id,
    trim(spirit.sent_month) AS sent_month
  FROM spirits AS spirit
  WHERE spirit.tablet_id IS NOT NULL
    AND spirit.deleted_at IS NULL
    AND trim(spirit.sent_month) <> ''
  ORDER BY spirit.tablet_id, spirit.created_at, spirit.id
) AS source
WHERE tablet.id = source.tablet_id
  AND trim(COALESCE(tablet.registered_at, '')) = '';
