# Настройка

DashLit читает переменные окружения процесса и необязательный файл `.env`. Переменные процесса имеют приоритет. В контейнерной установке передавайте параметры через Compose или другое средство запуска контейнеров.

## Основные параметры

Для большинства установок в контейнере достаточно безопасного `JWT_SECRET`. Внутренний адрес и пути к данным уже настроены в образе.

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `JWT_SECRET` | Значение для разработки | Секрет для подписи сессий. Обязательно замените в production. |
| `DEV_MODE` | `false` | Подробная диагностика базы данных для разработки. |

## Параметры аутентификации

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `OIDC_ISSUER` | Пусто | URL OIDC issuer. Пустое значение отключает OIDC. |
| `OIDC_CLIENT_ID` | Пусто | Client ID, зарегистрированный у провайдера. |
| `OIDC_CLIENT_SECRET` | Пусто | Client secret, если он требуется провайдером. |
| `OIDC_REDIRECT_URL` | Локальный callback | Внешний URL с окончанием `/api/auth/oidc/callback`. |
| `OIDC_BUTTON_TITLE` | `Sign in with OIDC` | Текст кнопки на странице входа. |
| `DISABLE_PASSWORD_REGISTRATION` | `false` | Запрещает регистрацию новых парольных аккаунтов. |
| `DISABLE_OIDC_REGISTRATION` | `false` | Запрещает неизвестным OIDC-профилям создавать аккаунты. |
| `DISABLE_OIDC_USER_MERGE` | `false` | Не связывает совпадающие парольные и OIDC-профили. |
| `DISABLE_PASSWORD_LOGIN` | `false` | Отключает вход по паролю после полной настройки OIDC. |

`OIDC_ISSUER` и `OIDC_CLIENT_ID` должны быть либо оба заполнены, либо оба пусты. Пока OIDC настроен не полностью, DashLit сохраняет вход по паролю как способ восстановления даже при `DISABLE_PASSWORD_LOGIN=true`.

## Пример production-окружения

```dotenv
JWT_SECRET=use-a-long-random-value
DEV_MODE=false

OIDC_ISSUER=https://id.example.com
OIDC_CLIENT_ID=dashlit
OIDC_CLIENT_SECRET=provider-issued-secret
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
OIDC_BUTTON_TITLE=Войти через Pocket ID

DISABLE_PASSWORD_REGISTRATION=true
DISABLE_OIDC_REGISTRATION=false
DISABLE_OIDC_USER_MERGE=false
DISABLE_PASSWORD_LOGIN=true
```

## Продвинутые настройки хранилища и сети

Эти параметры описывают внутреннее устройство приложения и обычно не нужны в стандартном контейнере:

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `ADDR` | `:8080` | Адрес и порт, прослушиваемые процессом. |
| `DATA_DIR` | `./data` | Каталог базы, загруженных иконок и кеша. Образ задаёт `/data`. |
| `DATABASE_PATH` | `$DATA_DIR/bookmarks.db` | Изменяет расположение файла SQLite. |

Меняйте их только для собственного runtime, нестандартной файловой структуры или запуска из исходников. Родительский каталог `DATABASE_PATH` должен быть доступен для записи. Legacy-миграция ищет `dashboard.json` именно рядом с этим файлом.

## Права внутри контейнера

Готовый образ работает от непривилегированного пользователя с UID и GID `10001`. Именованные Docker-тома инициализируются автоматически. Для bind mount выдайте этому пользователю право записи:

```bash
mkdir -p ./data
sudo chown -R 10001:10001 ./data
```
