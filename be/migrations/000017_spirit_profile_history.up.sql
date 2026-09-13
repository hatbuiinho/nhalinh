ALTER TABLE spirits
  ADD COLUMN familiar_name text NOT NULL DEFAULT '',
  ADD COLUMN gender text NOT NULL DEFAULT '',
  ADD COLUMN birth_date date,
  ADD COLUMN death_date date,
  ADD COLUMN birth_lunar text NOT NULL DEFAULT '',
  ADD COLUMN death_lunar text NOT NULL DEFAULT '',
  ADD COLUMN status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft','worshipping','enshrined','moved','dedicated','archived')),
  ADD COLUMN entered_worship_area_at date,
  ADD COLUMN enshrined_at date;

UPDATE spirits SET status = CASE WHEN tablet_id IS NULL THEN 'draft' ELSE 'worshipping' END;

CREATE TABLE spirit_position_history (
  id text PRIMARY KEY,
  spirit_id text NOT NULL REFERENCES spirits(id) ON DELETE CASCADE,
  from_tablet_id text,
  to_tablet_id text,
  from_position_id text,
  to_position_id text,
  change_type text NOT NULL,
  changed_at timestamptz NOT NULL DEFAULT now(),
  notes text NOT NULL DEFAULT '',
  performed_by text NOT NULL DEFAULT ''
);
CREATE INDEX idx_spirit_position_history_spirit_changed_at ON spirit_position_history(spirit_id, changed_at DESC);

CREATE FUNCTION record_spirit_position_change() RETURNS trigger AS $$
BEGIN
  IF TG_OP = 'INSERT' OR OLD.tablet_id IS DISTINCT FROM NEW.tablet_id THEN
    INSERT INTO spirit_position_history(id,spirit_id,from_tablet_id,to_tablet_id,from_position_id,to_position_id,change_type,changed_at)
    VALUES ('spirit-position-' || md5(random()::text || clock_timestamp()::text), NEW.id,
      CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE OLD.tablet_id END, NEW.tablet_id,
      CASE WHEN TG_OP = 'INSERT' THEN NULL ELSE (SELECT position_id FROM memorial_tablets WHERE id=OLD.tablet_id) END,
      (SELECT position_id FROM memorial_tablets WHERE id=NEW.tablet_id),
      CASE WHEN TG_OP = 'INSERT' THEN 'created' WHEN NEW.tablet_id IS NULL THEN 'unplaced' WHEN OLD.tablet_id IS NULL THEN 'placed' ELSE 'moved' END, now());
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER spirits_position_history AFTER INSERT OR UPDATE OF tablet_id ON spirits FOR EACH ROW EXECUTE FUNCTION record_spirit_position_change();
