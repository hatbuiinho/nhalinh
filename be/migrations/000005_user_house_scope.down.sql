ALTER TABLE spirit_house_members
  ADD COLUMN role text NOT NULL DEFAULT 'viewer'
  CHECK (role IN ('manager', 'editor', 'viewer'));

UPDATE spirit_house_members hm
SET role = CASE WHEN u.role = 'editor' THEN 'editor' ELSE 'viewer' END
FROM users u
WHERE u.id = hm.user_id;

ALTER TABLE users DROP CONSTRAINT users_role_check;
UPDATE users SET role = 'viewer' WHERE role = 'editor';
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'viewer'));
ALTER TABLE users DROP COLUMN all_houses;
