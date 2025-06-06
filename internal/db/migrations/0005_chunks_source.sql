-- +goose Up
ALTER TABLE chunks ADD COLUMN IF NOT EXISTS source TEXT DEFAULT '';
CREATE INDEX IF NOT EXISTS chunks_source_idx ON chunks(source);
CREATE UNIQUE INDEX IF NOT EXISTS chunks_content_uidx ON chunks(content);

-- +goose Down
DROP INDEX IF EXISTS chunks_content_uidx;
DROP INDEX IF EXISTS chunks_source_idx;
ALTER TABLE chunks DROP COLUMN IF EXISTS source;
