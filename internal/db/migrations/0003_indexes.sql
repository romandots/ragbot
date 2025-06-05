-- +goose Up
-- Create indexes to speed up queries and ensure uniqueness
CREATE UNIQUE INDEX IF NOT EXISTS idx_chunks_content ON chunks (content);
CREATE INDEX IF NOT EXISTS idx_conversation_chat_id ON conversation_history (chat_id);
-- Index for processed chunks selection
CREATE INDEX IF NOT EXISTS idx_chunks_processed_at ON chunks (processed_at);
-- Vector index for similarity search
CREATE INDEX IF NOT EXISTS idx_chunks_embedding ON chunks USING ivfflat (embedding vector_cosine_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_chunks_embedding;
DROP INDEX IF EXISTS idx_chunks_processed_at;
DROP INDEX IF EXISTS idx_conversation_chat_id;
DROP INDEX IF EXISTS idx_chunks_content;
