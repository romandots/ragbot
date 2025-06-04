-- +goose Up
ALTER TABLE chunks
    ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW(),
    ADD COLUMN processed_at TIMESTAMPTZ DEFAULT NULL,
    ALTER COLUMN embedding DROP NOT NULL;

-- +goose Down
ALTER TABLE chunks
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS processed_at,
    ALTER COLUMN embedding SET NOT NULL;
