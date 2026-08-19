ALTER TABLE users ADD COLUMN all_houses boolean NOT NULL DEFAULT false;

UPDATE users u
SET role = 'editor'
WHERE u.role <> 'admin'
  AND EXISTS (
    SELECT 1
    FROM spirit_house_members hm
    WHERE hm.user_id = u.id AND hm.role IN ('manager', 'editor')
  );

UPDATE users u
SET all_houses = u.role = 'admin'
  OR NOT EXISTS (SELECT 1 FROM spirit_house_members hm WHERE hm.user_id = u.id);

ALTER TABLE users DROP CONSTRAINT users_role_check;
ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('admin', 'editor', 'viewer'));

ALTER TABLE spirit_house_members DROP COLUMN role;
