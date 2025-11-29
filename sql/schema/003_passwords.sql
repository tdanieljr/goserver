-- +goose up
ALTER TABLE users ADD COLUMN hash_password TEXT NOT NULL DEFAULT 'unset';

-- +goose down
ALTER TABLE users DROP COLUMN hash_password;
