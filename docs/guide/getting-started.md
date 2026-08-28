# Introduction

DashLit is a self-hosted start page for bookmarks and services. It combines a visual dashboard, multi-user access controls, OIDC authentication, basic endpoint availability checks, and portable dashboard exports in one application.

## What you can build

A DashLit installation can contain multiple dashboards. Each dashboard contains ordered groups, and each group contains ordered links. You can rearrange both with drag and drop and select the layout that best matches the content.

- **Public dashboards** are visible without signing in.
- **Authenticated dashboards** are visible to every signed-in user.
- **Private dashboards** are visible only to their owner and administrators.
- A **system main dashboard** acts as the landing dashboard for visitors.
- Each user can select an owned dashboard as their **personal default**.

Dashboards support rows, columns, and masonry layouts, regular or wide page widths, clean mode, light and dark icons, and per-link availability checks.

## Accounts and administration

The first account created in an empty database becomes an administrator. Later accounts are regular users unless their access is changed by an administrator.

Password login and OIDC can be enabled together. When an OIDC identity has the same username as an existing password account, DashLit links them by default so the user retains both login methods. This behavior can be disabled in configuration.

## How DashLit stores data

The default container stores all persistent state under `/data`:

```text
/data/
├── bookmarks.db       # SQLite database
├── icons/             # uploaded icons
└── icon-cache/        # rebuildable remote icon cache
```

Keep this directory on persistent storage and include the database and uploaded icons in backups.

## Next steps

1. [Install DashLit](/guide/installation).
2. Review the [configuration reference](/guide/configuration).
3. Configure [OIDC](/guide/oidc) if you use a central identity provider.
4. Learn how to [create and share dashboards](/guide/usage).
