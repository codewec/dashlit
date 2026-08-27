# DashLit beta

Self-hosted bookmark dashboard (Svelte 5 + Go + SQLite).

Single binary: frontend is built and embedded into the Go server.

## Features

- Multiple dashboards per user, custom slug, privacy (public / users / private)
- Admin can set a main dashboard for `/`
- Groups + items (links) with 1×1 / 1×2 sizes
- Edit mode with modal forms
- Icons: Iconify, URL, upload (stored on disk + Iconify cache)
- Themes: light / dark / system + per-dashboard overrides (API ready)
- Local login/register (first user = admin); OIDC (PocketID) hooks prepared
- Simple search/filter on items

## Stack

- Frontend: Svelte 5, TypeScript, Vite, svelte-spa-router
- Backend: Go, chi, uptrace/bun, SQLite (modernc)
- Auth: JWT (cookie + Bearer)

## Quick start

```bash
# requires: Go 1.22+, Node 20+, npm
make build
./bookmarks
# open http://localhost:8080
```

Dev (two terminals):

```bash
make dev-backend    # :8080 API
make dev-frontend   # :5173 UI with proxy to API
```

## Env

| Variable           | Default                                        | Description              |
| ------------------ | ---------------------------------------------- | ------------------------ |
| ADDR               | `:8080`                                        | Listen address           |
| DATA_DIR           | `./data`                                       | DB + icons               |
| JWT_SECRET         | (dev default)                                  | **Change in production** |
| OIDC_ISSUER        |                                                | PocketID issuer URL      |
| OIDC_CLIENT_ID     |                                                |                          |
| OIDC_CLIENT_SECRET |                                                |                          |
| OIDC_REDIRECT_URL  | `http://localhost:8080/api/auth/oidc/callback` |                          |
| DEV_MODE           | `false`                                        | Verbose SQL              |

## API (prefix `/api`)

- `POST /auth/login|register|logout`, `GET /auth/me`
- `GET/POST /dashboards`, `GET /dashboards/main`, `GET/PUT/DELETE /dashboards/{id}`
- `POST /dashboards/{id}/set-main` (admin)
- Groups/items CRUD + `PUT /dashboards/{id}/layout` (batch positions after DnD)
- `POST /icons/upload`, `GET /icons/{id}`, `GET /icons/iconify/{prefix}/{name}`
