# Установка

Готовый контейнер поддерживает `linux/amd64`, `linux/arm64` и `linux/arm/v7` (armhf). Docker Compose рекомендуется для воспроизводимого хранения конфигурации и описания томов.

## Docker Compose

Создайте каталог установки и сохраните следующий файл как `docker-compose.yml`:

```yaml
services:
  dashlit:
    image: ghcr.io/codewec/dashlit:main
    container_name: dashlit
    restart: unless-stopped
    ports:
      - '3000:8080'
    environment:
      JWT_SECRET: '${JWT_SECRET}'
    volumes:
      - dashlit-data:/data

volumes:
  dashlit-data:
```

Рядом создайте файл `.env`:

```dotenv
JWT_SECRET=replace-with-a-long-random-secret
```

Используйте случайно сгенерированный секрет и не добавляйте его в Git. Запустите сервис:

```bash
docker compose up -d
```

Откройте `http://localhost:3000` и зарегистрируйте первую учётную запись — она получит права администратора.

## Docker CLI

```bash
docker pull ghcr.io/codewec/dashlit:main
docker run -d \
  --name dashlit \
  --restart unless-stopped \
  -p 3000:8080 \
  -e JWT_SECRET='replace-with-a-long-random-secret' \
  -v dashlit-data:/data \
  ghcr.io/codewec/dashlit:main
```

## Обратный прокси

Для доступа из интернета или общей сети публикуйте DashLit через обратный прокси с HTTPS. Проксируйте запросы на порт `8080` контейнера и передавайте исходные заголовки хоста и протокола. Специальная настройка WebSocket не требуется.

При использовании OIDC задайте внешний HTTPS callback и укажите точно такой же URL у провайдера:

```dotenv
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
```

## Фиксация версии

Тег `main` указывает на последнюю сборку текущего поколения DashLit. Для предсказуемых обновлений используйте конкретный тег релиза, например:

```yaml
image: ghcr.io/codewec/dashlit:v1.0.0
```

Тег `latest` намеренно не используется для текущего поколения, чтобы существующие установки старой версии не обновились автоматически.

Перед обновлением прочитайте [историю изменений](/ru/changelog), сохраните `/data`, загрузите новый образ и пересоздайте контейнер.

## Сборка из исходников

Для разработки необходимы Go 1.25+, Node.js 22+ и pnpm:

```bash
git clone https://github.com/codewec/dashlit.git
cd dashlit
cp .env.example .env
cd frontend && pnpm install && cd ..
make build
./app
```

Для разработки с автоматическим обновлением запустите `make dev-backend` и `make dev-frontend` в разных терминалах.
