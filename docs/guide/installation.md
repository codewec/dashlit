# Installation

The published container supports `linux/amd64`, `linux/arm64`, and `linux/arm/v7` (armhf). Docker Compose is recommended because it keeps configuration and volume declarations reproducible.

## Docker Compose

For the quickest installation, download the ready-to-use Compose file, generate a random `JWT_SECRET`, and start DashLit:

```bash
mkdir dashlit && cd dashlit
curl -fsSLo docker-compose.yml https://raw.githubusercontent.com/codewec/dashlit/main/docker-compose.yml
printf 'JWT_SECRET=%s\n' "$(openssl rand -hex 32)" > .env
docker compose up -d
```

Open `http://localhost:3000`. Register the first account; it becomes the administrator. Keep the generated `.env` file private and do not commit it.

Alternatively, create the deployment files manually. Save the following as `docker-compose.yml`:

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

Create a `.env` file beside it:

```dotenv
JWT_SECRET=replace-with-a-long-random-secret
```

Use a randomly generated secret and do not commit it. Then start the service:

```bash
docker compose up -d
```

### Bind mounts

The minimal example uses a named volume, which lets Docker initialize `/data` with the ownership stored in the image. If you prefer a host directory:

```yaml
volumes:
  - ./data:/data
```

create it before starting the container and allow writing for everyone. This is the simplest option and keeps the directory accessible to your host user:

```bash
mkdir -p data
sudo chmod -R 777 data
docker compose up -d
```

Docker does not adjust ownership of bind-mounted host directories. If `./data` does not exist, Docker normally creates it as `root:root`, which prevents the non-root DashLit process from writing to it. On SELinux systems, use `./data:/data:Z` as well.

For a more restrictive setup, make the container UID/GID `10001:10001` the owner and limit access. Your regular host user may then lose write access to the directory:

```bash
sudo chown -R 10001:10001 data
sudo chmod -R 750 data
```

DashLit checks the data, database, uploaded-icon, and icon-cache paths before normal startup. If any path is not writable, it serves an error page at the regular DashLit address and also writes the details to its logs. The process stays alive instead of entering a restart loop, and the container becomes `unhealthy`. Fix the permissions and restart it:

```bash
docker compose restart dashlit
```

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

## Existing Linux system

DashLit can be installed as a standalone service on a systemd-based Linux distribution. The installer supports `amd64`, `arm64`, and `armv7`, downloads the matching binary from the latest GitHub Release, and verifies it against the published SHA-256 checksums.

Run it as root:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/install.sh | sudo bash
```

### Dependencies installed by the script

The installer checks for `curl`, `tar`, `sha256sum`, and `openssl`. If any are missing, it uses the available system package manager to install the following packages:

- Debian and Ubuntu: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`;
- Fedora and other systems using DNF: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`;
- RHEL-compatible systems using YUM: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`;
- Alpine: `ca-certificates`, `curl`, `tar`, `coreutils`, `openssl`.

The target system must already use systemd and provide standard account-management commands such as `useradd`, `usermod`, and `groupadd`. No Go compiler, Node.js, npm, pnpm, database server, or Docker runtime is installed. DashLit itself remains a standalone binary with SQLite embedded through its Go driver.

The installation uses these paths:

| Purpose | Path |
| --- | --- |
| Binary and installed version | `/opt/dashlit` |
| Configuration | `/etc/dashlit/dashlit.env` |
| Database and icons | `/var/lib/dashlit` |
| systemd unit | `/etc/systemd/system/dashlit.service` |

DashLit runs as a dedicated system user and listens on port `8080`. Edit `/etc/dashlit/dashlit.env` for OIDC or other settings, then apply changes with:

```bash
sudo systemctl restart dashlit
```

### Updates

The installer adds a `dashlit-update` command. Run it to install the newest release:

```bash
sudo dashlit-update
```

The updater verifies the release checksum, preserves configuration and data, and restores the previous executable if the updated service cannot start. Back up `/var/lib/dashlit` and `/etc/dashlit/dashlit.env` before significant upgrades.

### Uninstall

Remove the systemd service, executable, and update command while preserving configuration and data:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/uninstall.sh | sudo bash
```

This leaves `/etc/dashlit` and `/var/lib/dashlit` in place so DashLit can be reinstalled later. To permanently delete those directories and the `dashlit` system account, use `--purge`:

```bash
curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/uninstall.sh | sudo bash -s -- --purge
```

The purge operation cannot be undone. Back up the data directory first if its contents may still be needed.

## Proxmox VE LXC

Run the following command as root in the Proxmox VE host shell:

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/proxmox-lxc.sh)"
```

The script creates an unprivileged Debian 13 container with 1 CPU, 512 MB RAM, a 4 GB disk, DHCP networking on `vmbr0`, automatic startup, and the `nesting=1` feature required by systemd 257 in current Debian 13 templates. It reuses the newest Debian 13 template of the host architecture already present in the selected template storage. A template is downloaded only when no matching local Debian 13 template exists. It then invokes the regular Linux installer inside the container. DashLit listens on port `80` there and is available at `http://CONTAINER_IP` without a port suffix.

On the Proxmox host, the script only uses tools already supplied by Proxmox VE: `pct`, `pvesh`, `pvesm`, and `pveam`. Inside the new container it initially installs `ca-certificates` and `curl`; the regular installer then checks and installs any missing archive, checksum, or OpenSSL tools listed above.

The following environment variables override the defaults:

| Variable | Default | Purpose |
| --- | --- | --- |
| `DASHLIT_CTID` | Next available ID | Container ID |
| `DASHLIT_HOSTNAME` | `dashlit` | Container hostname |
| `DASHLIT_STORAGE` | First active `rootdir` storage | Root filesystem storage |
| `DASHLIT_TEMPLATE_STORAGE` | First active `vztmpl` storage | Debian template storage |
| `DASHLIT_BRIDGE` | `vmbr0` | Network bridge |
| `DASHLIT_IP_CONFIG` | `dhcp` | Proxmox `ip=` value |
| `DASHLIT_CORES` | `1` | CPU cores |
| `DASHLIT_MEMORY` | `512` | Memory in MB |
| `DASHLIT_DISK` | `4` | Disk size in GB |

For example:

```bash
DASHLIT_CTID=120 DASHLIT_STORAGE=local-lvm DASHLIT_MEMORY=1024 \
  bash -c "$(curl -fsSL https://raw.githubusercontent.com/codewec/dashlit/main/scripts/proxmox-lxc.sh)"
```

The Proxmox console automatically opens a local root session without asking for a password. You can also enter the container from the host with `pct enter CTID`. This console-only autologin does not enable passwordless root access over SSH. Updates are installed from inside the container with `dashlit-update`.

The uninstall command removes DashLit inside the LXC but does not delete the container itself. If the entire container is no longer needed, back it up and remove it through Proxmox VE instead.

### Why a project-owned installer?

The Community Scripts new-application form currently requires at least 1,000 GitHub stars or a comparable public adoption signal. DashLit does not yet meet that requirement. Rather than depend on a permanently modified fork of ProxmoxVE, DashLit publishes a small installer maintained and released with the application itself. The requirement can be reviewed in the [Community Scripts contribution repository](https://github.com/community-scripts/ProxmoxVED/blob/main/.github/ISSUE_TEMPLATE/script_request.yml).

## Reverse proxy

Expose DashLit through an HTTPS reverse proxy for internet or shared-network use. Proxy requests to port `8080` in the container and preserve the original host and protocol headers. WebSocket-specific configuration is not required.

If OIDC is enabled, set `OIDC_REDIRECT_URL` to the public HTTPS callback and register exactly the same URL with the identity provider:

```dotenv
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
```

## Pinning a release

The `main` tag follows the newest release of the current DashLit generation. For predictable upgrades, replace it with a versioned release tag, for example:

```yaml
image: ghcr.io/codewec/dashlit:v1.0.0
```

The `dev` tag is rebuilt after every push to the `main` branch. It may contain changes that have not been included in a release yet and is intended for testing.

The `latest` tag is intentionally not used for the current generation so existing legacy installations are not upgraded automatically.

Read the [changelog](/changelog), back up `/data`, pull the new image, and recreate the container.

## Build from source

Development requires Go 1.25+, Node.js 22+, and pnpm:

```bash
git clone https://github.com/codewec/dashlit.git
cd dashlit
cp .env.example .env
cd frontend && pnpm install && cd ..
make build
./app
```

For live development, run `make dev-backend` and `make dev-frontend` in separate terminals. The frontend development server at port `5173` proxies API requests to the backend at port `8080`.
