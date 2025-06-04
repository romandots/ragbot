-- Включаем расширение pgvector (если ещё не установлено)
CREATE EXTENSION IF NOT EXISTS vector;

-- Создаём таблицу chunks для хранения фрагментов базы знаний
-- embedding типа VECTOR(1536) соответствует размерности эмбеддинга OpenAI AdaEmbeddingV2
CREATE TABLE IF NOT EXISTS chunks (
                                      id SERIAL PRIMARY KEY,
                                      content TEXT NOT NULL,
    embedding VECTOR(1536) NOT NULL
    );

-- Храним историю сообщений пользователей и бота
CREATE TABLE IF NOT EXISTS conversation_history (
    id SERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
