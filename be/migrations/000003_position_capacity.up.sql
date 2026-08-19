ALTER TABLE memorial_positions
  ADD COLUMN max_tablets integer CHECK (max_tablets IS NULL OR max_tablets > 0);

COMMENT ON COLUMN memorial_positions.max_tablets IS 'Số bài vị tối đa tại vị trí; NULL khi chưa cấu hình sức chứa';
