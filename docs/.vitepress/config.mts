import { defineConfig } from 'vitepress';

export default defineConfig({
  title: 'DashLit',
  description: 'Documentation for the modern, self-hosted dashboard for links and services.',
  base: '/dashlit/',
  cleanUrls: true,
  lastUpdated: true,
  locales: {
    root: {
      label: 'English',
      lang: 'en-US',
      title: 'DashLit',
      description: 'Documentation for the modern, self-hosted dashboard for links and services.',
    },
    ru: {
      label: 'Русский',
      lang: 'ru-RU',
      link: '/ru/',
      title: 'DashLit',
      description: 'Документация современного самостоятельно размещаемого дашборда для ссылок и сервисов.',
      themeConfig: {
        nav: [
          { text: 'Руководство', link: '/ru/guide/getting-started' },
          { text: 'Настройка', link: '/ru/guide/configuration' },
          { text: 'Миграция', link: '/ru/guide/migration' },
          { text: 'Изменения', link: '/ru/changelog' },
          {
            text: 'Сообщество',
            items: [
              { text: 'Обсуждения', link: 'https://github.com/codewec/dashlit/discussions' },
              { text: 'Предложения и ошибки', link: 'https://github.com/codewec/dashlit/issues' },
            ],
          },
        ],
        sidebar: {
          '/ru/guide/': [
            {
              text: 'Начало работы',
              items: [
                { text: 'Знакомство', link: '/ru/guide/getting-started' },
                { text: 'Установка', link: '/ru/guide/installation' },
                { text: 'Настройка', link: '/ru/guide/configuration' },
              ],
            },
            {
              text: 'Работа с DashLit',
              items: [
                { text: 'Дашборды и доступ', link: '/ru/guide/usage' },
                { text: 'OIDC-аутентификация', link: '/ru/guide/oidc' },
                { text: 'Миграция со старой версии', link: '/ru/guide/migration' },
                { text: 'Резервные копии и обновления', link: '/ru/guide/backups' },
              ],
            },
            {
              text: 'Проект',
              items: [
                { text: 'Галерея скриншотов', link: '/ru/guide/screenshots' },
                { text: 'История изменений', link: '/ru/changelog' },
              ],
            },
          ],
        },
        editLink: {
          pattern: 'https://github.com/codewec/dashlit/edit/beta/docs/:path',
          text: 'Изменить страницу на GitHub',
        },
        outline: { level: [2, 3], label: 'На этой странице' },
        lastUpdated: { text: 'Обновлено' },
        docFooter: { prev: 'Предыдущая страница', next: 'Следующая страница' },
        returnToTopLabel: 'Наверх',
        sidebarMenuLabel: 'Меню',
        darkModeSwitchLabel: 'Оформление',
        lightModeSwitchTitle: 'Переключить на светлую тему',
        darkModeSwitchTitle: 'Переключить на тёмную тему',
        langMenuLabel: 'Выбрать язык',
        skipToContentLabel: 'Перейти к содержимому',
        footer: {
          message: 'Распространяется по лицензии MIT.',
          copyright: 'Copyright © участники проекта DashLit',
        },
      },
    },
  },
  head: [
    ['link', { rel: 'icon', href: '/dashlit/logo.svg' }],
    ['meta', { name: 'theme-color', content: '#8839ef' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:title', content: 'DashLit' }],
    ['meta', { property: 'og:description', content: 'A modern, fast, and self-hosted dashboard for your links and services.' }],
    [
      'script',
      {
        defer: '',
        src: 'https://umami.0x2d.dev/script.js',
        'data-website-id': '0df202c2-a5ed-4c92-a2b3-979d7246ce5f',
      },
    ],
  ],
  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'DashLit',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Configuration', link: '/guide/configuration' },
      { text: 'Migration', link: '/guide/migration' },
      { text: 'Changelog', link: '/changelog' },
      {
        text: 'Community',
        items: [
          { text: 'Discussions', link: 'https://github.com/codewec/dashlit/discussions' },
          { text: 'Ideas and bug reports', link: 'https://github.com/codewec/dashlit/issues' },
        ],
      },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Get started',
          items: [
            { text: 'Introduction', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Configuration', link: '/guide/configuration' },
          ],
        },
        {
          text: 'Use DashLit',
          items: [
            { text: 'Dashboards and access', link: '/guide/usage' },
            { text: 'OIDC authentication', link: '/guide/oidc' },
            { text: 'Migrate from legacy', link: '/guide/migration' },
            { text: 'Backups and upgrades', link: '/guide/backups' },
          ],
        },
        {
          text: 'Project',
          items: [
            { text: 'Screenshot gallery', link: '/guide/screenshots' },
            { text: 'Changelog', link: '/changelog' },
          ],
        },
      ],
    },
    socialLinks: [{ icon: 'github', link: 'https://github.com/codewec/dashlit' }],
    editLink: {
      pattern: 'https://github.com/codewec/dashlit/edit/beta/docs/:path',
      text: 'Edit this page on GitHub',
    },
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © DashLit contributors',
    },
    search: {
      provider: 'local',
      options: {
        locales: {
          ru: {
            translations: {
              button: { buttonText: 'Поиск', buttonAriaLabel: 'Поиск' },
              modal: {
                displayDetails: 'Показать подробности',
                resetButtonTitle: 'Сбросить поиск',
                backButtonTitle: 'Закрыть поиск',
                noResultsText: 'Ничего не найдено',
                footer: {
                  selectText: 'выбрать',
                  selectKeyAriaLabel: 'Enter',
                  navigateText: 'перейти',
                  navigateUpKeyAriaLabel: 'Стрелка вверх',
                  navigateDownKeyAriaLabel: 'Стрелка вниз',
                  closeText: 'закрыть',
                  closeKeyAriaLabel: 'Escape',
                },
              },
            },
          },
        },
      },
    },
    i18nRouting: true,
    outline: { level: [2, 3], label: 'On this page' },
    lastUpdated: { text: 'Last updated' },
    docFooter: { prev: 'Previous page', next: 'Next page' },
  },
});
