-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS conversations (
    uuid UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    chat_id BIGINT NOT NULL UNIQUE,
    summary TEXT,
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS conversations_chat_id_idx ON conversations(chat_id);

-- +goose Down
DROP TABLE IF EXISTS conversations;
