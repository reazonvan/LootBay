-- +migrate Up
ALTER TABLE users ADD COLUMN phone VARCHAR(20) UNIQUE;

-- +migrate Down
ALTER TABLE users DROP COLUMN phone; 