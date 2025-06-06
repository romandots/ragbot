-- +goose Up
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS name TEXT;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS phone TEXT;

-- +goose Down
ALTER TABLE conversations DROP COLUMN IF EXISTS phone;
ALTER TABLE conversations DROP COLUMN IF EXISTS name;
