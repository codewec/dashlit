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

## Существующая Linux-система

DashLit можно установить как самостоятельный сервис в Linux с systemd. Установщик поддерживает `amd64`, `arm64` и `armv7`, загружает подходящий бинарник из последнего GitHub Release и проверяет его по опубликованной контрольной сумме SHA-256.

Запустите его с правами root:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/install.sh | sudo bash
```

### Зависимости, устанавливаемые скриптом

Установщик проверяет наличие `curl`, `tar`, `sha256sum` и `openssl`. Если хотя бы одна команда отсутствует, через доступный системный менеджер пакетов устанавливаются:

- Debian и Ubuntu: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`;
- Fedora и другие системы с DNF: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`;
- RHEL-совместимые системы с YUM: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`;
- Alpine: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`.

В целевой системе уже должны использоваться systemd и стандартные команды управления учётными записями: `useradd`, `usermod` и `groupadd`. Скрипт не устанавливает компилятор Go, Node.js, npm, pnpm, сервер базы данных или Docker. Само приложение остаётся самостоятельным бинарником; SQLite используется через встроенный Go-драйвер.

При установке используются следующие пути:

| Назначение | Путь |
| --- | --- |
| Бинарник и установленная версия | `/opt/dashlit` |
| Конфигурация | `/etc/dashlit/dashlit.env` |
| База данных и иконки | `/var/lib/dashlit` |
| unit systemd | `/etc/systemd/system/dashlit.service` |

DashLit работает от отдельного системного пользователя и слушает порт `8080`. Для настройки OIDC и других параметров измените `/etc/dashlit/dashlit.env`, затем примените изменения:

```bash
sudo systemctl restart dashlit
```

### Обновления

Установщик добавляет команду `dashlit-update`. Запустите её для установки последнего релиза:

```bash
sudo dashlit-update
```

Обновление проверяет контрольную сумму релиза, сохраняет конфигурацию и данные и возвращает предыдущий исполняемый файл, если обновлённый сервис не запускается. Перед значительными обновлениями сохраняйте резервную копию `/var/lib/dashlit` и `/etc/dashlit/dashlit.env`.

### Удаление

Чтобы удалить systemd-сервис, исполняемый файл и команду обновления, сохранив конфигурацию и данные, выполните:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/uninstall.sh | sudo bash
```

Каталоги `/etc/dashlit` и `/var/lib/dashlit` останутся на месте, поэтому приложение можно будет установить повторно. Для полного удаления этих каталогов и системной учётной записи `dashlit` передайте `--purge`:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/uninstall.sh | sudo bash -s -- --purge
```

Полное удаление необратимо. Если данные могут понадобиться, сначала создайте их резервную копию.

## Proxmox VE LXC

Выполните следующую команду с правами root в консоли узла Proxmox VE:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/proxmox-lxc.sh)"
```

Скрипт создаёт непривилегированный контейнер Debian 13 с 1 ядром CPU, 512 МБ памяти, диском 4 ГБ, сетью DHCP через `vmbr0` и автоматическим запуском. Сначала он использует самый новый шаблон Debian 13 для архитектуры узла, уже находящийся в выбранном хранилище шаблонов. Загрузка выполняется только тогда, когда подходящего локального шаблона Debian 13 нет. Затем внутри контейнера вызывается обычный Linux-установщик. DashLit слушает порт `80` и доступен по адресу `http://IP_КОНТЕЙНЕРА` без указания порта.

На узле Proxmox скрипт использует только входящие в Proxmox VE инструменты: `pct`, `pvesh`, `pvesm` и `pveam`. В новом контейнере сначала устанавливаются `ca-certificates` и `curl`, после чего обычный установщик проверяет и при необходимости добавляет перечисленные выше инструменты для работы с архивами, контрольными суммами и OpenSSL.

Значения по умолчанию можно изменить переменными окружения:

| Переменная | По умолчанию | Назначение |
| --- | --- | --- |
| `DASHLIT_CTID` | Следующий свободный ID | ID контейнера |
| `DASHLIT_HOSTNAME` | `dashlit` | Имя контейнера |
| `DASHLIT_STORAGE` | Первое активное хранилище `rootdir` | Хранилище корневой файловой системы |
| `DASHLIT_TEMPLATE_STORAGE` | Первое активное хранилище `vztmpl` | Хранилище шаблона Debian |
| `DASHLIT_BRIDGE` | `vmbr0` | Сетевой мост |
| `DASHLIT_IP_CONFIG` | `dhcp` | Значение параметра Proxmox `ip=` |
| `DASHLIT_CORES` | `1` | Количество ядер CPU |
| `DASHLIT_MEMORY` | `512` | Объём памяти в МБ |
| `DASHLIT_DISK` | `4` | Размер диска в ГБ |

Пример:

```bash
DASHLIT_CTID=120 DASHLIT_STORAGE=local-lvm DASHLIT_MEMORY=1024 \
  bash -c "$(curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/proxmox-lxc.sh)"
```

Войти в контейнер можно командой `pct enter CTID`. Для обновления выполните внутри него `dashlit-update`.

Команда удаления удаляет DashLit внутри LXC, но не сам контейнер. Если контейнер больше не нужен, создайте его резервную копию и удалите средствами Proxmox VE.

### Почему используется собственный установщик?

Форма добавления нового приложения Community Scripts сейчас требует не менее 1000 звёзд GitHub или сопоставимого публичного показателя распространённости. DashLit пока не соответствует этому требованию. Вместо постоянной зависимости от изменённого форка ProxmoxVE проект публикует небольшой установщик, который поддерживается и выпускается вместе с самим приложением. Требование можно проверить в [репозитории для contributions Community Scripts](https://github.com/community-scripts/ProxmoxVED/blob/main/.github/ISSUE_TEMPLATE/script_request.yml).

## Обратный прокси

Для доступа из интернета или общей сети публикуйте DashLit через обратный прокси с HTTPS. Проксируйте запросы на порт `8080` контейнера и передавайте исходные заголовки хоста и протокола. Специальная настройка WebSocket не требуется.

При использовании OIDC задайте внешний HTTPS callback и укажите точно такой же URL у провайдера:

```dotenv
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
```

## Фиксация версии

Тег `main` указывает на последний релиз текущего поколения DashLit. Для предсказуемых обновлений используйте конкретный тег релиза, например:

```yaml
image: ghcr.io/codewec/dashlit:v1.0.0
```

Тег `dev` пересобирается после каждого push в ветку `main`. Он может содержать изменения, которые ещё не вошли в релиз, и предназначен для тестирования.

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
