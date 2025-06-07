# RAG Bot

Бот-ассистент на основе ИИ

## Пререквизиты

- Docker и Docker Compose
- Go (для локальной разработки)

## Переменные окружения

Для корректной работы приложения требуются следующие переменные окружения в файле `.env`:

### Обязательные переменные

| Переменная | Описание |
|------------|----------|
| `DATABASE_URL` | URL подключения к PostgreSQL в формате `postgres://user:password@host:port/dbname` |
| `POSTGRES_DB` | Имя базы данных PostgreSQL |
| `POSTGRES_USER` | Имя пользователя PostgreSQL |
| `POSTGRES_PASSWORD` | Пароль пользователя PostgreSQL |
| `USER_TELEGRAM_TOKEN` | Токен для пользовательского Telegram бота |
| `ADMIN_TELEGRAM_TOKEN` | Токен для административного Telegram бота |
| `ADMIN_CHAT_IDS` | Список ID чатов администраторов (через запятую) |

### Опциональные переменные

| Переменная | Описание |
|------------|----------|
| `BASE_URL` | Базовый URL приложения (по умолчанию `localhost:8080`) |
| `DOMAIN_NAME` | Домен для настройки Nginx и SSL |
| `USE_LOCAL_MODEL` | Флаг для использования локальной модели (`true` или `false`) |
| `OPENAI_API_KEY` | API ключ для OpenAI (обязателен, если `USE_LOCAL_MODEL=false`) |
| `USER_TELEGRAM_BOT_NAME` | Имя пользовательского Telegram бота |
| `EDUCATION_FILE_PATH` | Путь к файлу с обучающими материалами |
| `USE_EXTERNAL_SOURCE` | Флаг для использования внешнего источника данных (`true` или `false`) |
| `YANDEX_YML_URL` | URL к файлу Yandex YML |
| `AMO_DOMAIN` | Домен amoCRM (например, `example.amocrm.ru`) |
| `AMO_ACCESS_TOKEN` | OAuth токен доступа для API amoCRM |
| `PREAMBLE` | Преамбула для взаимодействия с моделью |
| `CALL_MANAGER_TRIGGER_WORDS` | Слова-триггеры для вызова менеджера (через запятую) |

## Установка

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/romandots/rag-bot.git
   cd rag-bot
   ```

2. Создайте файл `.env` в корне проекта и добавьте необходимые переменные окружения:

```env
DATABASE_URL=postgres://user:password@db:5432/ragbot
POSTGRES_DB=ragbot
POSTGRES_USER=user
POSTGRES_PASSWORD=password
USER_TELEGRAM_TOKEN=your_user_telegram_token
ADMIN_TELEGRAM_TOKEN=your_admin_telegram_token
ADMIN_CHAT_IDS=123456789,987654321
USE_LOCAL_MODEL=true
OPENAI_API_KEY=your_openai_api_key
EDUCATION_FILE_PATH=education.txt
USE_EXTERNAL_SOURCE=false
DOMAIN_NAME=example.com
```

## Запуск в среде разработки

Для локальной разработки:

1. Убедитесь, что у вас установлены все зависимости:

   ```bash
   go mod download
   ```

2. Запустите PostgreSQL (можно через Docker):

   ```bash
   docker-compose up -d db
   ```

3. Запустите приложение локально:

   ```bash
   go run cmd/ragbot/main.go
   ```

Приложение будет доступно по адресу `http://localhost:8080`.

## Запуск в продакшен-среде

Для запуска в продакшен-среде используется Docker Compose с настройкой Nginx и SSL-сертификатов через Certbot:

1. Убедитесь, что в файле `.env` правильно настроен `DOMAIN_NAME`

2. Инициализируйте SSL-сертификаты (при первом запуске):

   ```bash
   chmod +x init-letsencrypt.sh
   ./init-letsencrypt.sh
   ```

3. Запустите все сервисы:

   ```bash
   docker-compose --env-file .env up -d
   ```

После запуска сервис будет доступен по адресу `https://ваш-домен.com`.

## Миграции базы данных

Для управления схемой используется [Goose](https://github.com/pressly/goose).
При запуске приложения все миграции из директории `internal/db/migrations`
применяются автоматически.

### Ручное управление миграциями

1. Установите утилиту Goose:

   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   ```

2. Применение миграций:

   ```bash
   goose -dir internal/db/migrations postgres $DATABASE_URL up
   ```

3. Откат последней миграции:

   ```bash
   goose -dir internal/db/migrations postgres $DATABASE_URL down
   ```

4. Создание новой миграции:

   ```bash
   goose -dir internal/db/migrations create имя_миграции sql
   ```

