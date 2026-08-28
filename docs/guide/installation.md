# Installation

The published container supports `linux/amd64` and `linux/arm64`. Docker Compose is recommended because it keeps configuration and volume declarations reproducible.

## Docker Compose

Create a directory for the deployment and save the following as `docker-compose.yml`:

```yaml
services:
  dashlit:
    image: ghcr.io/codewec/dashlit:beta
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

Open `http://localhost:3000`. Register the first account; it becomes the administrator.

## Docker CLI

```bash
docker pull ghcr.io/codewec/dashlit:beta
docker run -d \
  --name dashlit \
  --restart unless-stopped \
  -p 3000:8080 \
  -e JWT_SECRET='replace-with-a-long-random-secret' \
  -v dashlit-data:/data \
  ghcr.io/codewec/dashlit:beta
```

## Reverse proxy

Expose DashLit through an HTTPS reverse proxy for internet or shared-network use. Proxy requests to port `8080` in the container and preserve the original host and protocol headers. WebSocket-specific configuration is not required.

If OIDC is enabled, set `OIDC_REDIRECT_URL` to the public HTTPS callback and register exactly the same URL with the identity provider:

```dotenv
OIDC_REDIRECT_URL=https://dash.example.com/api/auth/oidc/callback
```

## Pinning a release

The `beta` tag follows the newest beta build. For predictable upgrades, replace it with a versioned prerelease tag, for example:

```yaml
image: ghcr.io/codewec/dashlit:v1.0.0-beta.1
```

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
