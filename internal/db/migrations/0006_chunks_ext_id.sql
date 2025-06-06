-- +goose Up
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS ext_id TEXT;
CREATE INDEX IF NOT EXISTS chunks_ext_id_idx ON chunks(ext_id);

-- +goose Down
DROP INDEX IF EXISTS chunks_ext_id_idx;
ALTER TABLE chunks DROP COLUMN IF EXISTS ext_id;
