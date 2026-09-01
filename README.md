<p align="center">
  <img src="frontend/src/assets/dashlit.svg" width="112" height="112" alt="DashLit logo">
</p>

<h1 align="center">DashLit</h1>

<p align="center">A modern, fast, and self-hosted dashboard for your links and services.</p>

<p align="center">
  <a href="https://github.com/codewec/dashlit/actions/workflows/docker.yml"><img alt="Build" src="https://img.shields.io/github/actions/workflow/status/codewec/dashlit/docker.yml?branch=main&label=build"></a>
  <a href="https://github.com/codewec/dashlit/releases"><img alt="GitHub Release" src="https://img.shields.io/github/v/release/codewec/dashlit"></a>
  <a href="https://codewec.github.io/dashlit/"><img alt="Documentation" src="https://img.shields.io/badge/docs-read-brightgreen?logo=readthedocs"></a>
  <a href="https://github.com/codewec/dashlit/discussions"><img alt="GitHub Discussions" src="https://img.shields.io/github/discussions/all/codewec/dashlit"></a>
  <a href="https://catppuccin.com/"><img alt="Catppuccin themes" src="https://img.shields.io/badge/themes-Catppuccin-cba6f7?logo=catppuccin&logoColor=1e1e2e"></a>
  <a href="LICENSE"><img alt="License" src="https://img.shields.io/github/license/codewec/dashlit"></a>
</p>

> [!TIP]
> **Visit the [DashLit documentation](https://codewec.github.io/dashlit/)** for complete installation and configuration guides, detailed feature descriptions, migration instructions and screenshots.

## Screenshots

<p align="center">
  <img src=".github/screenshots/dahslit-default.png" width="49%" alt="DashLit default dashboard">
  <img src=".github/screenshots/dashlit-clean-themes.png" width="49%" alt="DashLit clean layout and themes">
</p>

## Highlights

- Multiple dashboards with public, authenticated, or private visibility
- Flexible rows, columns, and masonry layouts
- Drag-and-drop groups and items with desktop and touch support
- Per-user default dashboards and a system default dashboard
- URL availability monitoring with live status chips
- Local password authentication and OIDC, including Pocket ID
- User profile and administration pages
- Import, export, and cloning for dashboards, groups, and items
- Built-in icon search across selfh.st/icons and Iconify
- Automatic light/dark icon pairing from selfh.st and light rendering of monochrome Iconify icons on dark themes
- Multiple light and dark Catppuccin-inspired themes
- A single Go binary with the Svelte frontend embedded

## Run with Docker

The current release is published for `linux/amd64`, `linux/arm64`, and `linux/arm/v7` (armhf) under the `main` tag.

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

Open [http://localhost:3000](http://localhost:3000). The first account created with password authentication becomes an administrator.

### Docker Compose

Download the production Compose file, generate a random `JWT_SECRET`, and start DashLit:

```bash
mkdir dashlit && cd dashlit
curl -fsSLo docker-compose.yml https://raw.githubusercontent.com/codewec/dashlit/main/docker-compose.yml
printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env
docker compose up -d
```

The production compose file pulls `ghcr.io/codewec/dashlit:main`. Pin a release by setting `DASHLIT_TAG`, for example:

```dotenv
DASHLIT_TAG=v1.0.0
```

To build the image from the current checkout instead:

```bash
docker compose up --build -d
```

Application state, uploaded icons, and the SQLite database are stored under `/data` in the container.

## Install on Linux or Proxmox

DashLit can run directly on an existing systemd-based Linux installation without Docker. Run the installer as root:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/install.sh | sudo bash
```

The installer downloads the matching `amd64`, `arm64`, or `armv7` binary from the latest GitHub Release, verifies its SHA-256 checksum, creates a dedicated `dashlit` user, and starts a hardened systemd service on port `8080`. Configuration is stored in `/etc/dashlit/dashlit.env` and persistent data in `/var/lib/dashlit`.

Install future releases with:

```bash
sudo dashlit-update
```

Remove the service and binary while keeping configuration and data for a later reinstall:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/uninstall.sh | sudo bash
```

To permanently remove configuration and application data as well, pass `--purge`:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/uninstall.sh | sudo bash -s -- --purge
```

To create a dedicated unprivileged Debian LXC on a Proxmox VE host, run this command in the Proxmox shell:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/proxmox-lxc.sh)"
```

The container receives its own address over DHCP and serves DashLit on port `80`, so it opens directly at `http://CONTAINER_IP`. The script automatically chooses the next container ID and suitable storage. Its defaults can be overridden with `DASHLIT_CTID`, `DASHLIT_HOSTNAME`, `DASHLIT_STORAGE`, `DASHLIT_TEMPLATE_STORAGE`, `DASHLIT_BRIDGE`, `DASHLIT_IP_CONFIG`, `DASHLIT_CORES`, `DASHLIT_MEMORY`, and `DASHLIT_DISK`.

Community Scripts currently requires new application requests to demonstrate at least 1,000 GitHub stars or comparable public adoption. DashLit does not yet meet that threshold, so the project provides and maintains its own installer instead of publishing an unofficial dependency on a fork of ProxmoxVE.

## Migrate from legacy DashLit

Legacy releases stored dashboard data in `dashboard.json`. To migrate, place that file beside the new SQLite database before creating any users. With the standard container paths, the files are:

```text
/data/dashboard.json
/data/bookmarks.db
```

Start DashLit with an empty database. When the login page reports that legacy data was found, enable the import switch and create the first user with a password or OIDC. DashLit converts the legacy groups and links into a private dashboard owned by that first user and selects it as their personal default.

Detection is intentionally performed only at server startup while the users table is empty. If the option does not appear, verify the paths and file permissions, then check the container logs for JSON parsing errors. Back up the legacy directory before migrating and keep the original file until the imported dashboard has been verified.

See the [complete migration guide](https://codewec.github.io/dashlit/guide/migration) for supported fields and troubleshooting.

## Configuration

DashLit reads both process environment variables and a `.env` file. Process environment variables take precedence.

Most container installations only need to set `JWT_SECRET` and, when required, the OIDC options.

| Variable                        | Default             | Description                                            |
| ------------------------------- | ------------------- | ------------------------------------------------------ |
| `JWT_SECRET`                    | Development value   | Signing secret; always replace in production           |
| `INITIAL_ADMIN_USERNAME`        | Empty               | Username for one-time first administrator creation     |
| `INITIAL_ADMIN_PASSWORD`        | Empty               | Password for one-time first administrator creation     |
| `DEV_MODE`                      | `false`             | Enable development diagnostics                         |
| `OIDC_ISSUER`                   | Empty               | OIDC issuer URL; leave empty to disable OIDC           |
| `OIDC_CLIENT_ID`                | Empty               | OIDC client ID                                         |
| `OIDC_CLIENT_SECRET`            | Empty               | OIDC client secret                                     |
| `OIDC_REDIRECT_URL`             | Local callback      | Public callback URL                                    |
| `OIDC_BUTTON_TITLE`             | `Sign in with OIDC` | OIDC button label                                      |
| `OIDC_INSECURE_SKIP_TLS_VERIFY` | `false`             | Disable OIDC TLS verification for local testing only   |
| `DISABLE_PASSWORD_REGISTRATION` | `false`             | Disable password registration                          |
| `DISABLE_OIDC_REGISTRATION`     | `false`             | Prevent OIDC from creating new users                   |
| `DISABLE_OIDC_USER_MERGE`       | `false`             | Prevent OIDC identities from linking to existing users |
| `DISABLE_PASSWORD_LOGIN`        | `false`             | Disable password login once OIDC is configured         |
| `UPDATE_CHECK_ENABLED`          | `true`              | Check GitHub Releases for a newer stable version       |

### Advanced runtime settings

The container image already provides appropriate values for these internal settings. Override them only for a custom runtime, filesystem layout, or source installation.

| Variable        | Default                  | Description                                                           |
| --------------- | ------------------------ | --------------------------------------------------------------------- |
| `ADDR`          | `:8080`                  | Internal HTTP listen address                                          |
| `DATA_DIR`      | `./data`                 | Database, uploaded icons, and cache directory; the image uses `/data` |
| `DATABASE_PATH` | `$DATA_DIR/bookmarks.db` | Custom SQLite database path                                           |

Release builds embed their Git tag and commit in the executable. Check a native installation with `dashlit --version`; the same version is shown in the application footer. Update checks are cached for 12 hours and failures never prevent DashLit from starting. Set `UPDATE_CHECK_ENABLED=false` to disable the outbound GitHub request.

To provision the first administrator non-interactively, set both `INITIAL_ADMIN_USERNAME` and `INITIAL_ADMIN_PASSWORD` before the first start. The password must contain at least six characters. DashLit creates the account only while the users table is empty; after any user exists, both variables are ignored and can be removed from the deployment configuration. They do not reset or update an existing account.

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

The user documentation is a separate VitePress project under `docs/` and does not add Node dependencies to the repository root:

```bash
cd docs
npm install
npm run dev
```

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

Releases use tags such as `v1.0.0`. Each release publishes both the matching versioned container image and the stable `main` image, then creates a GitHub Release. Pushes to the `main` branch publish the development image as `dev`. The `latest` tag remains on the legacy generation and is intentionally not published by these workflows.

See [RELEASING.md](RELEASING.md) for the complete maintainer release procedure.

## Data and backups

Stop the container before copying the data volume, or use SQLite's online backup tooling. At minimum, preserve:

- `bookmarks.db`
- `icons/`
- `icon-cache/` (optional; it can be rebuilt)

## License

DashLit is available under the [MIT License](LICENSE).
