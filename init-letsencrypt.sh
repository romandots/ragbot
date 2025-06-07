#!/bin/bash

# Загрузить переменные из .env файла
if [ -f .env ]; then
  set -a
  source .env
  set +a
fi

domains=($DOMAIN_NAME)
rsa_key_size=4096
data_path="./certbot"
email="ram.d.kreiz@gmail.com"

if [ -d "$data_path" ]; then
  read -p "Существующие данные найдены для $domains. Продолжить и перезаписать существующие сертификаты? (y/N) " decision
  if [ "$decision" != "Y" ] && [ "$decision" != "y" ]; then
    exit
  fi
fi

if [ ! -e "$data_path/conf/options-ssl-nginx.conf" ] || [ ! -e "$data_path/conf/ssl-dhparams.pem" ]; then
  echo "### Загрузка рекомендуемых TLS параметров ..."
  mkdir -p "$data_path/conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$data_path/conf/options-ssl-nginx.conf"
  curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$data_path/conf/ssl-dhparams.pem"
  echo
fi

echo "### Создание dummy сертификатов для $domains ..."
path="/etc/letsencrypt/live/$domains"
mkdir -p "$data_path/conf/live/$domains"
docker compose run --rm --entrypoint "\
  openssl req -x509 -nodes -newkey rsa:$rsa_key_size -days 1\
    -keyout '$path/privkey.pem' \
    -out '$path/fullchain.pem' \
    -subj '/CN=localhost'" certbot
echo

echo "### Запуск nginx ..."
docker compose up --force-recreate -d nginx
echo

echo "### Удаление dummy сертификатов для $domains ..."
docker compose run --rm --entrypoint "\
  rm -Rf /etc/letsencrypt/live/$domains && \
  rm -Rf /etc/letsencrypt/archive/$domains && \
  rm -Rf /etc/letsencrypt/renewal/$domains.conf" certbot
echo

echo "### Запрос Let's Encrypt сертификата для $domains ..."
#Присоединяем email (опционально)
domain_args=""
for domain in "${domains[@]}"; do
  domain_args="$domain_args -d $domain"
done

# Выбираем между staging и production для избежания rate limits
case "$1" in
  "staging")
    echo "### Получаем staging сертификат ..."
    staging_arg="--staging"
    ;;
  "production")
    echo "### Получаем production сертификат ..."
    staging_arg=""
    ;;
  *)
    echo "### Аргумент не распознан, получаем staging сертификат ..."
    staging_arg="--staging"
    ;;
esac

docker compose run --rm --entrypoint "\
  certbot certonly --webroot -w /var/www/certbot \
    $staging_arg \
    --email $email \
    $domain_args \
    --rsa-key-size $rsa_key_size \
    --agree-tos \
    --force-renewal" certbot
echo

echo "### Перезапуск nginx ..."
docker compose exec nginx nginx -s reload
