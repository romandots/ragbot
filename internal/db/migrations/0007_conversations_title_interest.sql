-- +goose Up
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS interest TEXT;

-- +goose Down
ALTER TABLE conversations
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS interest;
