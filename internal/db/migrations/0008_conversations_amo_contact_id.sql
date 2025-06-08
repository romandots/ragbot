-- +goose Up
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS amo_contact_id INTEGER;

-- +goose Down
ALTER TABLE conversations DROP COLUMN IF EXISTS amo_contact_id;
