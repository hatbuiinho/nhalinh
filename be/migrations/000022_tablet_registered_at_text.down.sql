ALTER TABLE memorial_tablets
  ALTER COLUMN registered_at DROP NOT NULL,
  ALTER COLUMN registered_at DROP DEFAULT,
  ALTER COLUMN registered_at TYPE date USING NULL;
