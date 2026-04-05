DROP INDEX IF EXISTS idx_rooms_active_expires;
DROP TABLE IF EXISTS rooms;
ALTER TABLE users DROP COLUMN IF EXISTS created_at;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_unique;
