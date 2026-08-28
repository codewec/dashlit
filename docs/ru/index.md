---
layout: home

hero:
  name: DashLit
  text: Все ваши сервисы на одном дашборде
  tagline: Быстрая и современная домашняя страница для ссылок, инструментов и статусов сервисов — в одном контейнере на вашем сервере.
  image:
    src: /logo.svg
    alt: Логотип DashLit
  actions:
    - theme: brand
      text: Установить DashLit
      link: /ru/guide/installation
    - theme: alt
      text: Открыть руководство
      link: /ru/guide/getting-started
    - theme: alt
      text: Присоединиться к обсуждению
      link: https://github.com/codewec/dashlit/discussions

features:
  - icon: 🧭
    title: Несколько дашбордов
    details: Распределяйте ссылки по дашбордам и группам, выбирайте строки, колонки или masonry-раскладку.
  - icon: 🔐
    title: Гибкое управление доступом
    details: Используйте пароль, OIDC или оба способа. Делайте дашборды публичными, доступными пользователям или приватными.
  - icon: 📡
    title: Статусы сервисов
    details: Проверяйте доступность связанных сервисов и отображайте результат рядом со ссылкой.
  - icon: 🎨
    title: Настройка внешнего вида
    details: Светлые и тёмные темы, чистый режим, собственные иконки и широкая раскладка.
  - icon: ✨
    title: Две библиотеки иконок
    details: Ищите одновременно в selfh.st/icons и Iconify — тысячи иконок сервисов и универсальных наборов в независимой быстрой выдаче.
  - icon: 🌓
    title: Иконки с учётом темы
    details: DashLit автоматически подбирает светлую и тёмную версии selfh.st и сохраняет читаемость монохромных Iconify-иконок.
  - icon: 📦
    title: Простая эксплуатация
    details: Компактный автономный бинарник без runtime-зависимостей, один готовый контейнер и постоянное хранилище на SQLite.
  - icon: 🔁
    title: Удобный перенос
    details: Импортируйте, экспортируйте и клонируйте дашборды, включая миграцию со старых версий DashLit.
---

## Полезная стартовая страница без лишней сложности

DashLit даёт командам, владельцам домашних серверов и отдельным пользователям удобное место для доступа к ежедневным сервисам. Go-сервер уже содержит интерфейс на Svelte, поэтому для запуска достаточно одного контейнера и одного постоянного каталога данных.

<div class="home-screenshots">
  <figure>
    <img src="/dahslit-default.png" alt="Дашборд DashLit со стандартной раскладкой" loading="lazy">
    <figcaption>Все нужные сервисы на одном организованном дашборде</figcaption>
  </figure>
  <figure>
    <img src="/dashlit-clean-themes.png" alt="Чистая раскладка DashLit в светлой и тёмной темах" loading="lazy">
    <figcaption>Чистый режим со светлыми и тёмными темами</figcaption>
  </figure>
</div>

## Запуск за несколько секунд

Создайте `docker-compose.yml`:

```yaml
services:
  dashlit:
    image: ghcr.io/codewec/dashlit:main
    ports:
      - '3000:8080'
    environment:
      JWT_SECRET: replace-with-a-long-random-secret
    volumes:
      - dashlit-data:/data

volumes:
  dashlit-data:
```

Запустите сервис:

```bash
docker compose up -d
```

Откройте `http://localhost:3000`, создайте первую учётную запись и настройте дашборд. Первый пользователь автоматически получает права администратора.

[Перейти к установке →](/ru/guide/installation)

## Помогите сделать DashLit лучше

Есть вопрос по настройке или хотите показать свой вариант использования? [Начните обсуждение](https://github.com/codewec/dashlit/discussions). Если вы нашли ошибку или хотите предложить конкретную функцию, [создайте issue](https://github.com/codewec/dashlit/issues).
