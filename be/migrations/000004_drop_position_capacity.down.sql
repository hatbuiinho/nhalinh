ALTER TABLE memorial_positions
  ADD COLUMN max_tablets integer CHECK (max_tablets IS NULL OR max_tablets > 0);
