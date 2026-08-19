CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TABLE users (
  id text PRIMARY KEY,
  username text NOT NULL UNIQUE,
  display_name text NOT NULL,
  avatar_url text NOT NULL DEFAULT '',
  password_hash text NOT NULL,
  role text NOT NULL DEFAULT 'viewer' CHECK (role IN ('admin', 'viewer')),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX idx_users_active_display_name ON users(active, display_name, id);

CREATE TABLE user_sessions (
  token_hash text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  expires_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL
);

CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

CREATE TABLE user_devices (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  platform text NOT NULL CHECK (platform IN ('android', 'ios', 'web')),
  push_token text NOT NULL,
  enabled boolean NOT NULL DEFAULT true,
  last_seen_at timestamptz NOT NULL,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (user_id, push_token)
);

CREATE INDEX idx_user_devices_user_enabled ON user_devices(user_id, enabled, updated_at DESC);

CREATE TABLE spirit_houses (
  id text PRIMARY KEY,
  name varchar(120) NOT NULL,
  address text NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE TABLE spirit_house_members (
  house_id text NOT NULL REFERENCES spirit_houses(id) ON DELETE CASCADE,
  user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('manager', 'editor', 'viewer')),
  created_at timestamptz NOT NULL DEFAULT NOW(),
  PRIMARY KEY (house_id, user_id)
);

CREATE INDEX idx_spirit_house_members_user ON spirit_house_members(user_id, house_id);

CREATE TABLE memorial_areas (
  id text PRIMARY KEY,
  house_id text NOT NULL REFERENCES spirit_houses(id) ON DELETE CASCADE,
  code varchar(30) NOT NULL,
  name varchar(120) NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (house_id, code)
);

CREATE INDEX idx_memorial_areas_house ON memorial_areas(house_id, code);

CREATE TABLE memorial_positions (
  id text PRIMARY KEY,
  area_id text NOT NULL REFERENCES memorial_areas(id) ON DELETE CASCADE,
  name varchar(120) NOT NULL,
  row_number integer NOT NULL CHECK (row_number > 0),
  column_number integer NOT NULL CHECK (column_number > 0),
  notes text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (area_id, row_number, column_number),
  UNIQUE (area_id, name)
);

CREATE TABLE memorial_tablets (
  id text PRIMARY KEY,
  position_id text NOT NULL REFERENCES memorial_positions(id) ON DELETE CASCADE,
  name varchar(120) NOT NULL,
  notes text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL,
  UNIQUE (position_id, name)
);

CREATE INDEX idx_memorial_positions_area ON memorial_positions(area_id, row_number, column_number);
CREATE INDEX idx_memorial_tablets_position ON memorial_tablets(position_id, name);

CREATE TABLE spirits (
  id text PRIMARY KEY,
  tablet_id text NOT NULL REFERENCES memorial_tablets(id) ON DELETE CASCADE,
  full_name text NOT NULL,
  dharma_name text NOT NULL DEFAULT '',
  birth_year text NOT NULL DEFAULT '',
  death_year text NOT NULL DEFAULT '',
  age text NOT NULL DEFAULT '',
  image_url text NOT NULL DEFAULT '',
  burial_place text NOT NULL DEFAULT '',
  sender text NOT NULL DEFAULT '',
  sent_month text NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL,
  updated_at timestamptz NOT NULL
);

CREATE INDEX idx_spirits_tablet ON spirits(tablet_id, full_name);

COMMENT ON COLUMN memorial_positions.area_id IS 'Khu vực của vị trí; Nhà Linh được xác định qua memorial_areas.house_id';
COMMENT ON COLUMN memorial_positions.row_number IS 'Hàng của vị trí';
COMMENT ON COLUMN memorial_positions.column_number IS 'Cột của vị trí';
