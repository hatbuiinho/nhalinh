DELETE FROM urns
WHERE id LIKE 'urn-backfill-%'
  AND notes = 'Tạo tự động từ trạng thái Có Hũ Cốt trước đây.';
