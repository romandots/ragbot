# RAG Bot

Краткое описание репозитория и его назначения.

## Пререквизиты

- Docker и Docker Compose
- Переменные окружения:
  - USER_TELEGRAM_TOKEN
  - ADMIN_TELEGRAM_TOKEN
  - ADMIN_CHAT_IDS

## Установка

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/romandots/rag-bot.git
   cd rag-bot
   ```

2. Скачайте модель на основе Llama c `huggingface.co` и поместите готовый
`.gguf`-файл в корень проекта.

3. Создайте файл `.env` в корне проекта и добавьте туда необходимые
переменные окружения:

   ```env
   MODEL_PATH=path_to_your_model/model.gguf
   USER_TELEGRAM_TOKEN=your_user_telegram_token
   ADMIN_TELEGRAM_TOKEN=your_admin_telegram_token
   ADMIN_CHAT_IDS=your_admin_chat_ids_separated_by_commas
   ```

## Запуск

Соберите и запустите контейнеры:

```bash
docker-compose --env-file .env up --build
```

После запуска сервис будет доступен на `http://localhost:8080`.

## Миграции базы данных

Для управления схемой используется [Goose](https://github.com/pressly/goose).
При запуске приложения все миграции из директории `migrations`
применяются автоматически.
При необходимости запустить их вручную установите утилиту:

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Затем выполните:

```bash
goose -dir internal/db/migrations postgres $DATABASE_URL up
```
