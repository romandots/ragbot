# RAG Bot

Бот-ассистент на основе ИИ

## Пререквизиты

- Docker и Docker Compose
- Go (для локальной разработки)
- Доменное имя, указывающее на IP-адрес вашего сервера (для продакшн-развертывания)
- Открытые порты 80 и 443 на сервере (для продакшн-развертывания)

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

### Переменные для продакшн-развертывания

| Переменная | Описание |
|------------|----------|
| `DOMAIN_NAME` | Домен для настройки Nginx и SSL (обязательно для продакшена) |
| `SSL_EMAIL` | Email для получения SSL-сертификатов Let's Encrypt |

### Опциональные переменные

| Переменная | Описание |
|------------|----------|
| `BASE_URL` | Базовый URL приложения (по умолчанию `localhost:8080`) |
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
| `CERTBOT_STAGING` | Добавить `--staging` для тестовых сертификатов (опционально) |
| `STATS_USER` | Логин для доступа к странице статистики |
| `STATS_PASS` | Пароль для доступа к странице статистики |
| `TELEGRAM_CHANNEL` | Адрес Telegram-канала (без @) |

## Установка

1. Клонируйте репозиторий:

   ```bash
   git clone https://github.com/romandots/rag-bot.git
   cd rag-bot
   ```

2. Создайте файл `.env` по примеру `.env.example` в корне проекта и добавьте необходимые переменные окружения:

```env
# Основные переменные
DATABASE_URL=postgres://ragbot:your_password@db:5432/ragbot
POSTGRES_DB=ragbot
POSTGRES_USER=ragbot
POSTGRES_PASSWORD=your_secure_password_here
USER_TELEGRAM_TOKEN=your_user_telegram_token
ADMIN_TELEGRAM_TOKEN=your_admin_telegram_token
ADMIN_CHAT_IDS=123456789,987654321
TELEGRAM_CHANNEL=your_channel_name

# Переменные для продакшн-развертывания
DOMAIN_NAME=yourdomain.com
SSL_EMAIL=admin@yourdomain.com

# Опциональные переменные
USE_LOCAL_MODEL=true
OPENAI_API_KEY=your_openai_api_key
EDUCATION_FILE_PATH=education.txt
USE_EXTERNAL_SOURCE=false

# Для тестирования с staging сертификатами (раскомментируйте при необходимости)
# CERTBOT_STAGING=--staging
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

## Запуск в продакшн-среде

### Автоматическое развертывание (рекомендуется)

Для продакшн-развертывания используется автоматизированный скрипт, который решает проблему SSL-сертификатов:

1. **Настройте переменные окружения:**

   Убедитесь, что в файле `.env` правильно настроены `DOMAIN_NAME` и `SSL_EMAIL`.

2. **Запустите автоматическое развертывание:**

   ```bash
   chmod +x deploy-production.sh
   ./deploy-production.sh
   ```

Скрипт автоматически:
- Запустит приложение в HTTP-режиме
- Получит SSL-сертификаты от Let's Encrypt
- Переключит nginx на HTTPS-конфигурацию
- Настроит автоматическое обновление сертификатов

### Проверка развертывания

После запуска сервис будет доступен по адресам:
- HTTP: `http://ваш-домен.com`
- HTTPS: `https://ваш-домен.com`

Для проверки статуса сервисов:
```bash
docker-compose ps
```

Для просмотра логов:
```bash
docker-compose logs -f [имя_сервиса]
```

### Остановка сервисов

```bash
# Остановка всех сервисов
docker-compose down

# Остановка с удалением volumes (осторожно - удалит данные БД!)
docker-compose down -v
```

## Важные замечания для продакшна

1. **DNS-настройка**: Убедитесь, что A-запись вашего домена указывает на IP-адрес сервера
2. **Firewall**: Откройте порты 80 и 443 в firewall сервера
3. **Тестирование сертификатов**: При первом развертывании рекомендуется использовать staging-сертификаты (раскомментируйте `CERTBOT_STAGING=--staging` в `.env`)
4. **Мониторинг**: Следите за логами и статусом сервисов через `docker-compose logs` и `docker-compose ps`

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

## Устранение неполадок

### Проблемы с SSL-сертификатами

Если возникают проблемы с получением SSL-сертификатов:

1. Проверьте DNS-настройки:
   ```bash
   dig yourdomain.com
   ```

2. Используйте staging-сертификаты для тестирования:
   ```bash
   # В .env файле
   CERTBOT_STAGING=--staging
   ```

3. Проверьте логи certbot:
   ```bash
   docker-compose logs certbot
   ```

### Проблемы с nginx

Для проверки конфигурации nginx:
```bash
docker-compose exec nginx nginx -t
```

Для перезагрузки конфигурации:
```bash
docker-compose exec nginx nginx -s reload
```