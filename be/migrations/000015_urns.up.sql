ALTER TABLE spirits ADD COLUMN has_urn boolean NOT NULL DEFAULT false;

CREATE TABLE urns (
  id text PRIMARY KEY,
  spirit_id text NOT NULL UNIQUE REFERENCES spirits(id) ON DELETE CASCADE,
  code text NOT NULL DEFAULT '',
  urn_type text NOT NULL DEFAULT '',
  material text NOT NULL DEFAULT '',
  color text NOT NULL DEFAULT '',
  dimensions text NOT NULL DEFAULT '',
  storage_location text NOT NULL DEFAULT '',
  installed_at date,
  moved_at date,
  image_url text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'installed' CHECK (status IN ('installed','moved','returned')),
  notes text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_urns_status ON urns(status);
CREATE INDEX idx_urns_code ON urns(code);

CREATE TABLE urn_history (
  id text PRIMARY KEY,
  urn_id text NOT NULL REFERENCES urns(id) ON DELETE CASCADE,
  status text NOT NULL CHECK (status IN ('installed','moved','returned')),
  storage_location text NOT NULL DEFAULT '',
  changed_at date NOT NULL DEFAULT current_date,
  notes text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_urn_history_urn_changed_at ON urn_history(urn_id, changed_at DESC);
