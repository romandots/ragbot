-- +goose Up
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS username TEXT;

CREATE TABLE IF NOT EXISTS visits (
    id SERIAL PRIMARY KEY,
    ip TEXT,
    user_agent TEXT,
    referer TEXT,
    utm_source TEXT,
    utm_medium TEXT,
    utm_campaign TEXT,
    utm_content TEXT,
    utm_term TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS visits;
ALTER TABLE conversations DROP COLUMN IF EXISTS username;
