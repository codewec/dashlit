<p align="center">
  <img src="frontend/src/assets/dashlit.svg" width="112" height="112" alt="DashLit logo">
</p>

<h1 align="center">DashLit</h1>

<p align="center">A modern, fast, and self-hosted dashboard for your links and services.</p>

<p align="center">
  <a href="https://github.com/codewec/dashlit/actions/workflows/docker.yml"><img alt="Container build" src="https://github.com/codewec/dashlit/actions/workflows/docker.yml/badge.svg?branch=beta"></a>
  <a href="https://github.com/codewec/dashlit/pkgs/container/dashlit"><img alt="GHCR" src="https://img.shields.io/badge/GHCR-beta-blue?logo=docker"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/codewec/dashlit"></a>
</p>

> [!IMPORTANT]
> DashLit is currently in beta. Back up the data volume before upgrading and review release notes for breaking changes.

## Highlights

- Multiple dashboards with public, authenticated, or private visibility
- Flexible rows, columns, and masonry layouts
- Drag-and-drop groups and items with desktop and touch support
- Per-user default dashboards and a system default dashboard
- URL availability monitoring with live status chips
- Local password authentication and OIDC, including Pocket ID
- User profile and administration pages
- Import, export, and cloning for dashboards, groups, and items
- Multiple light and dark Catppuccin-inspired themes
- A single Go binary with the Svelte frontend embedded

## Run with Docker

The beta image is published for `linux/amd64` and `linux/arm64`.

```bash
docker pull ghcr.io/codewec/dashlit:beta
docker run -d \
  --name dashlit \
  --restart unless-stopped \
  -p 8080:8080 \
  -e JWT_SECRET='replace-with-a-long-random-secret' \
  -v dashlit-data:/data \
  ghcr.io/codewec/dashlit:beta
```

Open [http://localhost:8080](http://localhost:8080). The first account created with password authentication becomes an administrator.

### Docker Compose

For production, copy [`.env.example`](.env.example) to `.env`, set at least `JWT_SECRET`, and run:

```bash
docker compose -f docker-compose.prod.yml up -d
```

The production compose file pulls `ghcr.io/codewec/dashlit:beta`. Pin a prerelease by setting `DASHLIT_TAG`, for example:

```dotenv
DASHLIT_TAG=v1.0.0-beta.1
```

To build the image from the current checkout instead:

```bash
docker compose up --build -d
```

Application state, uploaded icons, and the SQLite database are stored under `/data` in the container.

## Configuration

DashLit reads both process environment variables and a `.env` file. Process environment variables take precedence.

| Variable                        | Default                  | Description                                            |
| ------------------------------- | ------------------------ | ------------------------------------------------------ |
| `ADDR`                          | `:8080`                  | HTTP listen address                                    |
| `DATA_DIR`                      | `./data`                 | Database, uploaded icons, and cache directory          |
| `DATABASE_PATH`                 | `$DATA_DIR/bookmarks.db` | SQLite database path                                   |
| `JWT_SECRET`                    | Development value        | Signing secret; always replace in production           |
| `DEV_MODE`                      | `false`                  | Enable development diagnostics                         |
| `OIDC_ISSUER`                   | Empty                    | OIDC issuer URL; leave empty to disable OIDC           |
| `OIDC_CLIENT_ID`                | Empty                    | OIDC client ID                                         |
| `OIDC_CLIENT_SECRET`            | Empty                    | OIDC client secret                                     |
| `OIDC_REDIRECT_URL`             | Local callback           | Public callback URL                                    |
| `OIDC_BUTTON_TITLE`             | `Sign in with OIDC`      | OIDC button label                                      |
| `DISABLE_PASSWORD_REGISTRATION` | `false`                  | Disable password registration                          |
| `DISABLE_OIDC_REGISTRATION`     | `false`                  | Prevent OIDC from creating new users                   |
| `DISABLE_OIDC_USER_MERGE`       | `false`                  | Prevent OIDC identities from linking to existing users |
| `DISABLE_PASSWORD_LOGIN`        | `false`                  | Disable password login once OIDC is configured         |

When DashLit is behind a reverse proxy, use the external HTTPS address for the callback:

```dotenv
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
```

Configure the same URL in your OIDC provider. The proxy should forward the original host and protocol headers.

## Development

Requirements: Go 1.25+, Node.js 22+, and pnpm.

```bash
git clone https://github.com/codewec/dashlit.git
cd dashlit
cp .env.example .env

cd frontend && pnpm install && cd ..
make dev-backend
```

In another terminal:

```bash
make dev-frontend
```

The frontend runs at `http://localhost:5173` and proxies API requests to the Go server at `http://localhost:8080`.

Build the embedded production binary with:

```bash
make build
./app
```

## Releases and changelog

DashLit uses [Conventional Commits](https://www.conventionalcommits.org/) and [git-cliff](https://git-cliff.org/). The Makefile downloads the pinned standalone binary into `./bin` automatically. To install it explicitly or preview pending release notes, run:

```bash
make git-cliff-install
make changelog-preview
```

During beta, releases use prerelease tags such as `v1.0.0-beta.1`. A tag publishes both the versioned container tag and `beta`, and creates a GitHub prerelease. The `latest` container tag is intentionally not published yet.

See [RELEASING.md](RELEASING.md) for the complete maintainer release procedure.

## Data and backups

Stop the container before copying the data volume, or use SQLite's online backup tooling. At minimum, preserve:

- `bookmarks.db`
- `icons/`
- `icon-cache/` (optional; it can be rebuilt)

## License

DashLit is available under the [MIT License](LICENSE).

<img src="https://umami.0x2d.dev/p/laer6FLsW" width="1" height="1">
